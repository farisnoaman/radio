package checkers

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/radiusd/cache"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"go.uber.org/zap"
)

var voucherBatchCache *cache.VoucherBatchCache

func InitVoucherBatchCache(ttl time.Duration) {
	voucherBatchCache = cache.NewVoucherBatchCache(ttl)
}

func GetVoucherBatchCache() *cache.VoucherBatchCache {
	return voucherBatchCache
}

type VoucherAuthChecker struct {
	voucherRepo    repository.VoucherRepository
	activeBatches  *cache.VoucherBatchCache
}

func NewVoucherAuthChecker(voucherRepo repository.VoucherRepository, activeBatches *cache.VoucherBatchCache) *VoucherAuthChecker {
	return &VoucherAuthChecker{
		voucherRepo:   voucherRepo,
		activeBatches: activeBatches,
	}
}

func (c *VoucherAuthChecker) Name() string {
	return "voucher_auth"
}

func (c *VoucherAuthChecker) Order() int {
	return 3
}

func (c *VoucherAuthChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User
	if user == nil {
		return nil
	}

	if user.VoucherBatchID == 0 {
		return nil
	}

	if !c.activeBatches.IsActive(user.VoucherBatchID) {
		zap.L().Debug("voucher auth: batch not active",
			zap.Int64("batch_id", user.VoucherBatchID),
			zap.String("username", user.Username))
		return radiuserrors.NewAuthErrorWithStage(
			"voucher_batch_inactive",
			"Voucher batch is not active. Please contact administrator.",
			"voucher_auth",
		)
	}

	voucher, err := c.voucherRepo.GetByCode(ctx, user.VoucherCode)
	if err != nil {
		zap.L().Error("voucher auth: voucher not found",
			zap.String("voucher_code", user.VoucherCode),
			zap.Error(err))
		return radiuserrors.NewAuthErrorWithStage(
			"voucher_not_found",
			"Voucher not found",
			"voucher_auth",
		)
	}

	if voucher.Status != "active" {
		zap.L().Debug("voucher auth: voucher not active",
			zap.String("voucher_code", voucher.Code),
			zap.String("status", voucher.Status))
		return radiuserrors.NewAuthErrorWithStage(
			"voucher_not_active",
			"Voucher is not active. Please contact administrator.",
			"voucher_auth",
		)
	}

	if !voucher.ExpireTime.IsZero() && voucher.ExpireTime.Before(time.Now()) {
		return radiuserrors.NewAuthErrorWithStage(
			"voucher_expired",
			"Voucher has expired. Please contact administrator.",
			"voucher_auth",
		)
	}

	if voucher.DataQuota > 0 && voucher.DataUsed >= voucher.DataQuota {
		return radiuserrors.NewAuthErrorWithStage(
			"voucher_data_quota_exceeded",
			"Voucher data quota exceeded. Please contact administrator.",
			"voucher_auth",
		)
	}

	if voucher.TimeQuota > 0 && voucher.TimeUsed >= voucher.TimeQuota {
		return radiuserrors.NewAuthErrorWithStage(
			"voucher_time_quota_exceeded",
			"Voucher time quota exceeded. Please contact administrator.",
			"voucher_auth",
		)
	}

	zap.L().Debug("voucher auth: passed",
		zap.Int64("batch_id", user.VoucherBatchID),
		zap.String("voucher_code", voucher.Code),
		zap.String("username", user.Username))

	return nil
}
