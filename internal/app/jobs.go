package app

import (
	"fmt"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/metrics"
	"go.uber.org/zap"
)

var cronParser = cron.NewParser(
	cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

func (a *Application) initJob() {
	loc, _ := time.LoadLocation(a.appConfig.System.Location)
	a.sched = cron.New(cron.WithLocation(loc), cron.WithParser(cronParser))

	var err error
	_, err = a.sched.AddFunc("@every 30s", func() {
		go a.SchedSystemMonitorTask()
		go a.SchedProcessMonitorTask()
	})
	if err != nil {
		zap.S().Errorf("init job error %s", err.Error())
	}

	_, err = a.sched.AddFunc("@daily", func() {
		go a.SchedLogArchivalTask()
	})
	if err != nil {
		zap.S().Errorf("init job error %s", err.Error())
	}

	_, err = a.sched.AddFunc("@every 1m", func() {
		go a.SchedSubscriptionRenewalTask()
	})
	if err != nil {
		zap.S().Errorf("init job error %s", err.Error())
	}

	// Backup task
	if a.appConfig.Backup.Enabled {
		_, err = a.sched.AddFunc(a.appConfig.Backup.Cron, func() {
			go a.SchedBackupTask()
		})
		if err != nil {
			zap.S().Errorf("init backup job error %s", err.Error())
		}
	}

	// Expire Strategy Task
	_, err = a.sched.AddFunc("@every 1h", func() {
		go a.SchedExpireStrategyTask()
	})
	if err != nil {
		zap.S().Errorf("init job error %s", err.Error())
	}

	a.sched.Start()
}

// SchedSubscriptionRenewalTask processes auto-renewals for active subscriptions
func (a *Application) SchedSubscriptionRenewalTask() {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Error(err)
		}
	}()

	var subs []domain.VoucherSubscription
	now := time.Now()
	// Find active subscriptions due for renewal
	err := a.gormDB.Where("status = ? AND auto_renew = ? AND next_renewal_at <= ? AND is_deleted = ?",
		"active", true, now, false).Find(&subs).Error
	if err != nil {
		zap.S().Errorf("Failed to fetch subscriptions for renewal: %v", err)
		return
	}

	for _, sub := range subs {
		a.processSubscriptionRenewal(sub)
	}
}

func (a *Application) processSubscriptionRenewal(sub domain.VoucherSubscription) {
	tx := a.gormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Get Voucher and verify it's active
	var voucher domain.Voucher
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("code = ?", sub.VoucherCode).First(&voucher).Error; err != nil {
		tx.Rollback()
		zap.S().Errorf("Subscription renewal failed: Voucher %s not found", sub.VoucherCode)
		return
	}

	if voucher.Status != "active" {
		tx.Rollback()
		return // Only renew active vouchers
	}

	// 2. Get Product price
	var product domain.Product
	if err := tx.First(&product, sub.ProductID).Error; err != nil {
		tx.Rollback()
		zap.S().Errorf("Subscription renewal failed: Product %d not found", sub.ProductID)
		return
	}

	// 3. Handle Wallet deduction if agent-owned
	if sub.AgentID > 0 {
		var wallet domain.AgentWallet
		if err := tx.Set("gorm:query_option", "FOR UPDATE").FirstOrCreate(&wallet, domain.AgentWallet{AgentID: sub.AgentID}).Error; err != nil {
			tx.Rollback()
			return
		}

		price := product.CostPrice
		if price <= 0 {
			price = product.Price
		}

		if wallet.Balance < price {
			tx.Rollback()
			zap.S().Warnf("Subscription renewal skipped: Insufficient balance for Agent %d", sub.AgentID)
			return
		}

		// Deduct balance
		wallet.Balance -= price
		if err := tx.Save(&wallet).Error; err != nil {
			tx.Rollback()
			return
		}

		// Log transaction
		tx.Create(&domain.WalletLog{
			AgentID:     sub.AgentID,
			Type:        "purchase",
			Amount:      -price,
			Balance:     wallet.Balance,
			ReferenceID: fmt.Sprintf("sub-%d", sub.ID),
			Remark:      fmt.Sprintf("Subscription renewal for voucher %s", voucher.Code),
			CreatedAt:   time.Now(),
		})
	}

	// 4. Extend RadiusUser expiry
	var user domain.RadiusUser
	if err := tx.Where("username = ?", voucher.Code).First(&user).Error; err == nil {
		newExpire := user.ExpireTime.Add(time.Duration(sub.IntervalDays) * 24 * time.Hour)
		if err := tx.Model(&user).Update("expire_time", newExpire).Error; err != nil {
			tx.Rollback()
			return
		}
		// Also update voucher expiry if it exists
		if !voucher.ExpireTime.IsZero() {
			tx.Model(&voucher).Update("expire_time", newExpire)
		}
	}

	// 5. Update Subscription status
	sub.LastRenewalAt = time.Now()
	sub.NextRenewalAt = sub.NextRenewalAt.Add(time.Duration(sub.IntervalDays) * 24 * time.Hour)
	sub.RenewalCount++
	if err := tx.Save(&sub).Error; err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()
	zap.S().Infof("Voucher subscription renewed: %s, next renewal: %v", sub.VoucherCode, sub.NextRenewalAt)
}

// SchedSystemMonitorTask system monitor
func (a *Application) SchedSystemMonitorTask() {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Error(err)
		}
	}()

	// Collect CPU usage
	_cpuuse, err := cpu.Percent(0, false)
	var cpuUsage int64
	if err == nil && len(_cpuuse) > 0 {
		cpuUsage = int64(_cpuuse[0] * 100)
		metrics.SetGauge("system_cpuuse", cpuUsage) // Store as percentage * 100
	}

	// Collect memory usage
	_meminfo, err := mem.VirtualMemory()
	var memUsage, memTotal int64
	if err == nil {
		memUsage = int64(_meminfo.Used / 1024 / 1024)
		memTotal = int64(_meminfo.Total / 1024 / 1024)
		metrics.SetGauge("system_memuse", memUsage) //nolint:gosec // G115: memory MB value fits in int64
	}

	// Broadcast to WebSocket
	if a.wsHub != nil {
		a.wsHub.Broadcast(map[string]interface{}{
			"type": "system_metrics",
			"data": map[string]interface{}{
				"cpu_usage": float64(cpuUsage) / 100.0,
				"mem_usage": memUsage,
				"mem_total": memTotal,
				"timestamp": time.Now().Unix(),
			},
		})
	}
}

// SchedProcessMonitorTask app process monitor
func (a *Application) SchedProcessMonitorTask() {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Error(err)
		}
	}()

	p, err := process.NewProcess(int32(os.Getpid())) //nolint:gosec // G115: PID is always within int32 range
	if err != nil {
		return
	}

	// Collect process CPU usage
	cpuuse, err := p.CPUPercent()
	if err == nil {
		metrics.SetGauge("toughradius_cpuuse", int64(cpuuse*100)) // Store as percentage * 100
	}

	// Collect process memory usage
	meminfo, err := p.MemoryInfo()
	if err == nil {
		metrics.SetGauge("toughradius_memuse", int64(meminfo.RSS/1024/1024)) //nolint:gosec // G115: memory MB value fits in int64
	}
}

func (a *Application) SchedClearExpireData() {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Error(err)
		}
	}()
	// Clean expire online
	a.gormDB.Where("last_update <= ?",
		time.Now().Add(time.Second*300*-1)).
		Delete(&domain.RadiusOnline{})

	// Clean up accounting logs
	idays := a.ConfigMgr().GetInt("radius", "AccountingHistoryDays")
	if idays == 0 {
		idays = 90
	}
	a.gormDB.
		Where("acct_stop_time < ? ", time.Now().
			Add(-time.Hour*24*time.Duration(idays))).Delete(domain.RadiusAccounting{})
}

func (a *Application) SchedBackupTask() {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Error(err)
		}
	}()
	
	zap.S().Info("Starting scheduled backup...")
	filename, err := a.BackupMgr().CreateBackup()
	if err != nil {
		zap.S().Errorf("Backup failed: %v", err)
		return
	}
	zap.S().Infof("Backup created successfully: %s", filename)
	
	// Prune old backups
	err = a.BackupMgr().PruneBackups(a.appConfig.Backup.MaxBackups)
	if err != nil {
		zap.S().Warnf("Failed to prune backups: %v", err)
	}
}

func (a *Application) SchedLogArchivalTask() {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Error(err)
		}
	}()
	
	// Archive logs older than 90 days (configurable ideally)
	err := a.archivalMgr.ArchiveSystemLogs(90)
	if err != nil {
		zap.S().Errorf("Log archival failed: %v", err)
	}
}

// SchedExpireStrategyTask handles actions for expired users (disable/delete/notify)
func (a *Application) SchedExpireStrategyTask() {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Error(err)
		}
	}()

	// 1. Disable expired users (if not already disabled)
	// Find users with expire_time < now AND status = 'enabled'
	now := time.Now()
	var expiredUsers []domain.RadiusUser
	err := a.gormDB.Where("expire_time < ? AND status = ?", now, "enabled").Find(&expiredUsers).Error
	if err != nil {
		zap.S().Errorf("Failed to fetch expired users: %v", err)
		return
	}

	if len(expiredUsers) > 0 {
		zap.S().Infof("Found %d expired users to disable", len(expiredUsers))
		// Bulk update status to 'disabled'
		err = a.gormDB.Model(&domain.RadiusUser{}).
			Where("expire_time < ? AND status = ?", now, "enabled").
			Update("status", "disabled").Error
		if err != nil {
			zap.S().Errorf("Failed to disable expired users: %v", err)
		} else {
			// Optional: Trigger Disconnect for these users if they are online
			// This would require iterating and calling DisconnectSession
			zap.S().Infof("Successfully disabled %d expired users", len(expiredUsers))
		}
	}

	// 2. (Optional) Archive/Delete old expired users (e.g., expired > 1 year)
	// retentionDate := now.AddDate(-1, 0, 0)
	// err = a.gormDB.Where("expire_time < ?", retentionDate).Delete(&domain.RadiusUser{}).Error
}
