package handlers

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/coa"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"layeh.com/radius/rfc2866"
)

// QuotaEnforcementHandler checks user quotas during interim updates
// and sends CoA Disconnect-Request to the NAS when quota is exceeded.
// This ensures users are disconnected when their time or data quota runs out,
// even if they are still actively connected.
type QuotaEnforcementHandler struct {
	db             *gorm.DB
	accountingRepo repository.AccountingRepository
	coaClient      *coa.Client
}

// NewQuotaEnforcementHandler creates a new quota enforcement handler
func NewQuotaEnforcementHandler(
	db *gorm.DB,
	accountingRepo repository.AccountingRepository,
) *QuotaEnforcementHandler {
	coaClient := coa.NewClient(coa.Config{
		Timeout:    5 * time.Second,
		RetryCount: 2,
		RetryDelay: 500 * time.Millisecond,
	})

	return &QuotaEnforcementHandler{
		db:             db,
		accountingRepo: accountingRepo,
		coaClient:      coaClient,
	}
}

func (h *QuotaEnforcementHandler) Name() string {
	return "QuotaEnforcementHandler"
}

func (h *QuotaEnforcementHandler) CanHandle(ctx *accounting.AccountingContext) bool {
	return ctx.StatusType == int(rfc2866.AcctStatusType_Value_InterimUpdate)
}

func (h *QuotaEnforcementHandler) Handle(acctCtx *accounting.AccountingContext) error {
	if h.db == nil {
		return nil
	}

	// Get the current session time from the interim update packet
	currentSessionTime := int64(rfc2866.AcctSessionTime_Get(acctCtx.Request.Packet))

	// Look up the user to check their quota settings
	var user domain.RadiusUser
	if err := h.db.Where("username = ?", acctCtx.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // User not found, skip
		}
		return err
	}

	// Check time quota enforcement
	if user.TimeQuota > 0 {
		h.enforceTimeQuota(acctCtx, &user, currentSessionTime)
	}

	// Check data quota enforcement
	if user.DataQuota > 0 {
		h.enforceDataQuota(acctCtx, &user)
	}

	return nil
}

// enforceTimeQuota checks if the user has exceeded their time quota and disconnects if so
func (h *QuotaEnforcementHandler) enforceTimeQuota(
	acctCtx *accounting.AccountingContext,
	user *domain.RadiusUser,
	currentSessionTime int64,
) {
	// Get total session time from completed accounting records
	totalPastTime, err := h.accountingRepo.GetTotalSessionTime(acctCtx.Context, user.Username)
	if err != nil {
		zap.L().Error("Failed to get total session time for quota enforcement",
			zap.String("username", user.Username),
			zap.Error(err))
		return
	}

	// Total time = past completed sessions + current active session time
	totalTime := totalPastTime + currentSessionTime

	zap.L().Debug("Quota enforcement check",
		zap.String("username", user.Username),
		zap.Int64("time_quota", user.TimeQuota),
		zap.Int64("total_past_time", totalPastTime),
		zap.Int64("current_session_time", currentSessionTime),
		zap.Int64("total_time", totalTime),
	)

	if totalTime >= user.TimeQuota {
		zap.L().Warn("Time quota exceeded, sending disconnect",
			zap.String("username", user.Username),
			zap.Int64("time_quota_sec", user.TimeQuota),
			zap.Int64("total_used_sec", totalTime),
		)
		h.disconnectUser(acctCtx)
	}
}

// enforceDataQuota checks if the user has exceeded their data quota and disconnects if so
func (h *QuotaEnforcementHandler) enforceDataQuota(
	acctCtx *accounting.AccountingContext,
	user *domain.RadiusUser,
) {
	// Get total data usage from completed accounting records
	totalPastUsage, err := h.accountingRepo.GetTotalUsage(acctCtx.Context, user.Username)
	if err != nil {
		zap.L().Error("Failed to get total usage for quota enforcement",
			zap.String("username", user.Username),
			zap.Error(err))
		return
	}

	// Convert data quota from MB to bytes for comparison
	dataQuotaBytes := user.DataQuota * 1024 * 1024

	if totalPastUsage >= dataQuotaBytes {
		zap.L().Warn("Data quota exceeded, sending disconnect",
			zap.String("username", user.Username),
			zap.Int64("data_quota_mb", user.DataQuota),
			zap.Int64("total_used_bytes", totalPastUsage),
		)
		h.disconnectUser(acctCtx)
	}
}

// disconnectUser sends a CoA Disconnect-Request to the NAS to terminate the user's session
func (h *QuotaEnforcementHandler) disconnectUser(acctCtx *accounting.AccountingContext) {
	nas := acctCtx.NAS
	if nas == nil {
		zap.L().Error("Cannot disconnect user: NAS info not available",
			zap.String("username", acctCtx.Username))
		return
	}

	nasIP := nas.Ipaddr
	if nasIP == "" {
		nasIP = acctCtx.NASIP
	}
	if nasIP == "" {
		zap.L().Error("Cannot disconnect user: NAS IP not available",
			zap.String("username", acctCtx.Username))
		return
	}

	secret := nas.Secret
	if secret == "" {
		zap.L().Error("Cannot disconnect user: NAS secret not available",
			zap.String("username", acctCtx.Username))
		return
	}

	coaPort := nas.CoaPort
	if coaPort <= 0 {
		coaPort = 3799 // Default CoA port
	}

	acctSessionID := rfc2866.AcctSessionID_GetString(acctCtx.Request.Packet)

	// Send disconnect request asynchronously to avoid blocking accounting processing
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp := h.coaClient.SendDisconnect(ctx, coa.DisconnectRequest{
			NASIP:         nasIP,
			NASPort:       coaPort,
			Secret:        secret,
			Username:      acctCtx.Username,
			AcctSessionID: acctSessionID,
		})

		if resp.Success {
			zap.L().Info("Quota exceeded: user disconnected successfully",
				zap.String("username", acctCtx.Username),
				zap.String("nas_ip", nasIP),
				zap.String("session_id", acctSessionID),
				zap.Duration("duration", resp.Duration),
			)
		} else {
			zap.L().Error("Quota exceeded: failed to disconnect user",
				zap.String("username", acctCtx.Username),
				zap.String("nas_ip", nasIP),
				zap.String("session_id", acctSessionID),
				zap.Error(resp.Error),
			)
		}
	}()
}
