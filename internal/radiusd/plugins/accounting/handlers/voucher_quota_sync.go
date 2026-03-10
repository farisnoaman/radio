package handlers

import (
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
)

type VoucherQuotaSyncHandler struct {
	db *gorm.DB
}

func NewVoucherQuotaSyncHandler(db *gorm.DB) *VoucherQuotaSyncHandler {
	return &VoucherQuotaSyncHandler{db: db}
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

	return nil
}
