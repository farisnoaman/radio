package checkers

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"go.uber.org/zap"
)

// FirstUseActivator checks if a user is logging in for the first time via a first-use voucher
// and activates the expiration countdown.
type FirstUseActivator struct {
	voucherRepo repository.VoucherRepository
	userRepo    repository.UserRepository
}

// NewFirstUseActivator creates a first-use activator instance
func NewFirstUseActivator(
	voucherRepo repository.VoucherRepository,
	userRepo repository.UserRepository,
) *FirstUseActivator {
	return &FirstUseActivator{
		voucherRepo: voucherRepo,
		userRepo:    userRepo,
	}
}

func (c *FirstUseActivator) Name() string {
	return "first_use_activator"
}

func (c *FirstUseActivator) Order() int {
	// Execute before ExpireChecker (which is 10) to ensure we update expiration before checking it
	return 5
}

func (c *FirstUseActivator) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User
	if user == nil {
		return nil
	}

	// Check if this is a "pending activation" user
	// We set ExpireTime to year 9999 in RedeemVoucher for first_use vouchers
	if user.ExpireTime.Year() < 9999 {
		// Already activated or not a first-use voucher
		return nil
	}

	// Double check by looking up the voucher
	voucher, err := c.voucherRepo.GetByCode(ctx, user.Username)
	if err != nil {
		// Not found or error -> ignore, treat as normal user
		return nil
	}

	// Get batch to confirm type and validity
	batch, err := c.voucherRepo.GetBatchByID(ctx, voucher.BatchID)
	if err != nil {
		zap.L().Error("first_use_activator: batch not found",
			zap.String("username", user.Username),
			zap.Int64("batch_id", voucher.BatchID),
			zap.Error(err))
		return nil
	}

	if batch.ExpirationType != "first_use" {
		// Should generally not happen if ExpireTime is 9999, but safe to ignore
		return nil
	}

	// Calculate new expiration
	now := time.Now()
	validityDuration := time.Duration(batch.ValidityDays) * 24 * time.Hour
	newExpire := now.Add(validityDuration)

	// Update RadiusUser
	if err := c.userRepo.UpdateField(ctx, user.Username, "expire_time", newExpire); err != nil {
		zap.L().Error("first_use_activator: failed to update user expiration",
			zap.String("username", user.Username),
			zap.Error(err))
		// If DB update fails, we should probably return error to prevent free access
		return err
	}

	// Update Voucher
	if err := c.voucherRepo.UpdateFirstUsedAt(ctx, voucher.Code, now, newExpire); err != nil {
		zap.L().Error("first_use_activator: failed to update voucher",
			zap.String("code", voucher.Code),
			zap.Error(err))
		// Log error but proceed since user is updated
	}

	// Update the user object in the current context so subsequent checkers see the correct time
	user.ExpireTime = newExpire

	zap.L().Info("first_use_activator: voucher activated on first login",
		zap.String("username", user.Username),
		zap.Time("activated_at", now),
		zap.Time("new_expire", newExpire))

	return nil
}
