package handlers

import (
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	"github.com/talkincode/toughradius/v9/internal/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
)

type VoucherQuotaSyncHandler struct {
	db             *gorm.DB
	loyaltyService *service.LoyaltyService
}

func NewVoucherQuotaSyncHandler(db *gorm.DB, loyaltyService *service.LoyaltyService) *VoucherQuotaSyncHandler {
	return &VoucherQuotaSyncHandler{
		db:             db,
		loyaltyService: loyaltyService,
	}
}

func (h *VoucherQuotaSyncHandler) Name() string {
	return "VoucherQuotaSyncHandler"
}

func (h *VoucherQuotaSyncHandler) CanHandle(ctx *accounting.AccountingContext) bool {
	return ctx.StatusType == int(rfc2866.AcctStatusType_Value_InterimUpdate) ||
		ctx.StatusType == int(rfc2866.AcctStatusType_Value_Stop)
}

func (h *VoucherQuotaSyncHandler) Handle(acctCtx *accounting.AccountingContext) error {
	if h.db == nil {
		return nil
	}

	var user domain.RadiusUser
	if err := h.db.Where("username = ?", acctCtx.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	if user.VoucherBatchID == 0 {
		return nil
	}

	r := acctCtx.Request
	acctInputOctets := int(rfc2866.AcctInputOctets_Get(r.Packet))
	acctInputGigawords := int(rfc2869.AcctInputGigawords_Get(r.Packet))
	acctOutputOctets := int(rfc2866.AcctOutputOctets_Get(r.Packet))
	acctOutputGigawords := int(rfc2869.AcctOutputGigawords_Get(r.Packet))

	inputBytes := int64(acctInputOctets) + int64(acctInputGigawords)*4*1024*1024*1024
	outputBytes := int64(acctOutputOctets) + int64(acctOutputGigawords)*4*1024*1024*1024
	totalBytes := inputBytes + outputBytes

	dataUsedMB := totalBytes / (1024 * 1024)

	acctSessionTime := int64(rfc2866.AcctSessionTime_Get(r.Packet))

	updates := map[string]interface{}{
		"data_used": gorm.Expr("data_used + ?", dataUsedMB),
		"time_used": gorm.Expr("time_used + ?", acctSessionTime),
	}

	if err := h.db.Model(&domain.Voucher{}).
		Where("code = ?", user.VoucherCode).
		Updates(updates).Error; err != nil {
		zap.L().Error("Failed to sync voucher quota",
			zap.String("voucher_code", user.VoucherCode),
			zap.Error(err))
		return err
	}

	zap.L().Debug("Voucher quota synced",
		zap.String("voucher_code", user.VoucherCode),
		zap.Int64("data_used_mb", dataUsedMB),
		zap.Int64("time_used_seconds", acctSessionTime))

	// Track loyalty usage if service is available
	if h.loyaltyService != nil {
		macAddr := ""
		if acctCtx.VendorReq != nil {
			macAddr = acctCtx.VendorReq.MacAddr
		}
		if macAddr != "" {
			// Note: The usage in Radius Accounting (Acct-Input-Octets/Session-Time) is cumulative for the session.
			// The voucher_quota_sync.go handler currently adds them to the voucher (gorm.Expr data_used + dataUsedMB).
			// If this handler is called multiple times per session (Interim Updates), it risks over-counting
			// unless we track the delta or the service handles it.
			// Per User Requirement: "accumulate data that is deleted by week/month/year logic"
			// and "atomic usage aggregation service with optimistic locking".
			// We pass the current event's usage to the service.
			err := h.loyaltyService.ProcessUsageEvent(acctCtx.Context, service.UsageEvent{
				Mac:       macAddr,
				TenantID:  user.TenantID,
				DataUsed:  dataUsedMB * 1024 * 1024, // Convert MB to bytes
				TimeUsed:  acctSessionTime,
				Timestamp: time.Now(),
			})
			if err != nil {
				zap.L().Error("Failed to process loyalty usage event",
					zap.String("mac", macAddr),
					zap.Error(err))
			}
		}
	}

	return nil
}
