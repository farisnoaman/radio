package checkers

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
)

// TimeQuotaChecker checks whether the user has exceeded their time quota
type TimeQuotaChecker struct {
	accountingRepo repository.AccountingRepository
}

// NewTimeQuotaChecker creates a time quota checker instance
func NewTimeQuotaChecker(accountingRepo repository.AccountingRepository) *TimeQuotaChecker {
	return &TimeQuotaChecker{
		accountingRepo: accountingRepo,
	}
}

func (c *TimeQuotaChecker) Name() string {
	return "time_quota"
}

func (c *TimeQuotaChecker) Order() int {
	return 16 // Execute after QuotaChecker (15)
}

func (c *TimeQuotaChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User
	if user == nil || user.TimeQuota <= 0 {
		return nil // No time quota configured, allow login
	}

	// Get total session time from accounting records
	totalTime, err := c.accountingRepo.GetTotalSessionTime(ctx, user.Username)
	if err != nil {
		// Log error but allow login on check failure
		// (safer to allow than to block on accounting errors)
		return nil
	}

	// TimeQuota is in seconds, totalTime is in seconds
	if totalTime >= user.TimeQuota {
		return errors.NewTimeQuotaError()
	}

	if authCtx.Metadata == nil {
		authCtx.Metadata = make(map[string]interface{})
	}
	authCtx.Metadata["remaining_time_quota"] = user.TimeQuota - totalTime

	return nil
}
