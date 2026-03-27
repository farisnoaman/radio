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
//
// Logic:
//   - On RedeemVoucher, first_use users get ExpireTime = year 9999 (placeholder).
//   - On first RADIUS login, this plugin detects the 9999 year, looks up the batch,
//     and sets User.ExpireTime = first_login_time + ValidityDays (hours).
//   - The calculated window is capped against the voucher's hard expiry date
//     (Voucher.ExpireTime, set from Batch.PrintExpireTime at creation).
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

	// Check if this is a "pending activation" user.
	// RedeemVoucher sets ExpireTime to year 9999 for first_use vouchers.
	if user.ExpireTime.Year() < 9999 {
		// Already activated or not a first-use voucher
		return nil
	}

	// Look up the voucher by username (voucher code = username)
	voucher, err := c.voucherRepo.GetByCode(ctx, user.Username)
	if err != nil {
		// Not found or error -> ignore, treat as normal user
		return nil
	}

	// Get batch to confirm type and validity window
	batch, err := c.voucherRepo.GetBatchByID(ctx, voucher.BatchID)
	if err != nil {
		zap.L().Error("first_use_activator: batch not found",
			zap.String("username", user.Username),
			zap.Int64("batch_id", voucher.BatchID),
			zap.Error(err))
		return nil
	}

	if batch.ExpirationType != "first_use" {
		// Not a first-use batch (shouldn't happen if ExpireTime is 9999, but safe guard)
		return nil
	}

	if batch.ValidityDays <= 0 {
		zap.L().Error("first_use_activator: ValidityDays is 0 or negative, cannot activate window",
			zap.String("username", user.Username),
			zap.Int64("batch_id", batch.ID),
			zap.Int("validity_days", batch.ValidityDays))
		return nil
	}

	// Calculate window expiry from first login time (now)
	now := time.Now()

	// ValidityDays is in HOURS, convert to duration
	windowDuration := time.Duration(batch.ValidityDays) * time.Hour
	windowExpiry := now.Add(windowDuration)

	// Determine the hard deadline: the voucher's original ExpireTime
	// (set from Batch.PrintExpireTime at batch creation time).
	// This is the absolute latest date the voucher can be valid.
	hardDeadline := voucher.ExpireTime
	// If voucher.ExpireTime is far-future default (2999), it means no hard cap was set
	isHardCapped := !hardDeadline.IsZero() && hardDeadline.Year() < 2999

	// Cap: final expiry = min(window_expiry, hard_deadline)
	newExpire := windowExpiry
	if isHardCapped && hardDeadline.Before(windowExpiry) {
		newExpire = hardDeadline
		zap.L().Info("first_use_activator: window capped by voucher expiry deadline",
			zap.String("username", user.Username),
			zap.Time("window_expiry", windowExpiry),
			zap.Time("hard_deadline", hardDeadline),
			zap.Time("final_expire", newExpire))
	}

	zap.L().Info("first_use_activator: activating voucher on first login",
		zap.String("username", user.Username),
		zap.Int64("batch_id", batch.ID),
		zap.Int("validity_hours", batch.ValidityDays),
		zap.Time("first_login_time", now),
		zap.Time("window_expiry", windowExpiry),
		zap.Time("final_expire", newExpire))

	// Update RadiusUser ExpireTime
	if err := c.userRepo.UpdateField(ctx, user.Username, "expire_time", newExpire); err != nil {
		zap.L().Error("first_use_activator: failed to update user expiration",
			zap.String("username", user.Username),
			zap.Error(err))
		return err
	}

	// Update Voucher first_used_at and expire_time
	if err := c.voucherRepo.UpdateFirstUsedAt(ctx, voucher.Code, now, newExpire); err != nil {
		zap.L().Error("first_use_activator: failed to update voucher",
			zap.String("code", voucher.Code),
			zap.Error(err))
		// Log error but proceed since user is already updated
	}

	// Update the user object in the current context so subsequent checkers see the correct time
	user.ExpireTime = newExpire

	return nil
}

