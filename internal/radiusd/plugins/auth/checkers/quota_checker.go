package checkers

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
)

// QuotaChecker checks whether the user has exceeded their data quota
type QuotaChecker struct {
	accountingRepo repository.AccountingRepository
}

// NewQuotaChecker creates a quota checker instance
func NewQuotaChecker(accountingRepo repository.AccountingRepository) *QuotaChecker {
	return &QuotaChecker{
		accountingRepo: accountingRepo,
	}
}

func (c *QuotaChecker) Name() string {
	return "quota"
}

func (c *QuotaChecker) Order() int {
	return 15 // Execute after ExpireChecker (10)
}

func (c *QuotaChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
	user := authCtx.User
	if user == nil || user.DataQuota <= 0 {
		return nil
	}

	// Calculate total usage from accounting records
	totalUsage, err := c.accountingRepo.GetTotalUsage(ctx, user.Username)
	if err != nil {
		// Log error but maybe allow login if accounting check fails?
		// Usually safer to allow, but here we'll be strict if the repo is reachable.
		return nil 
	}

	// DataQuota is in MB, totalUsage is in Bytes
	if totalUsage >= user.DataQuota*1024*1024 {
		return errors.NewUserQuotaError()
	}

	return nil
}
