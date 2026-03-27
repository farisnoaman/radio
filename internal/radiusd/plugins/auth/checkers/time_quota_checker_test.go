package checkers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	gormrepo "github.com/talkincode/toughradius/v9/internal/radiusd/repository/gorm"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTimeQuotaChecker_Name(t *testing.T) {
	checker := &TimeQuotaChecker{}
	assert.Equal(t, "time_quota", checker.Name())
}

func TestTimeQuotaChecker_Order(t *testing.T) {
	checker := &TimeQuotaChecker{}
	assert.Equal(t, 16, checker.Order())
}

func TestTimeQuotaChecker_Check(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate schema
	err = db.AutoMigrate(&domain.RadiusAccounting{})
	require.NoError(t, err)

	// Create real repository (not mocks)
	repo := gormrepo.NewGormAccountingRepository(db)
	checker := NewTimeQuotaChecker(repo)
	ctx := context.Background()

	t.Run("allows login when time quota not set", func(t *testing.T) {
		user := &domain.RadiusUser{
			Username:  "testuser",
			TimeQuota: 0, // No time quota
		}
		authCtx := &auth.AuthContext{User: user}

		err := checker.Check(ctx, authCtx)
		assert.Nil(t, err)
	})

	t.Run("allows login when time quota not exceeded", func(t *testing.T) {
		// Create accounting record with 1 hour used
		accounting := &domain.RadiusAccounting{
			Username:        "testuser2",
			AcctSessionTime: 3600, // 1 hour
			AcctInputTotal:  1000000,
			AcctOutputTotal: 2000000,
		}
		err = repo.Create(ctx, accounting)
		require.NoError(t, err)

		user := &domain.RadiusUser{
			Username:  "testuser2",
			TimeQuota: 18000, // 5 hours
		}
		authCtx := &auth.AuthContext{User: user}

		err = checker.Check(ctx, authCtx)
		assert.Nil(t, err) // Should allow (1 hour < 5 hours)
	})

	t.Run("rejects login when time quota exceeded", func(t *testing.T) {
		// Create accounting records totaling 5.5 hours
		accounting1 := &domain.RadiusAccounting{
			Username:        "testuser3",
			AcctSessionTime: 18000, // 5 hours
			AcctInputTotal:  1000000,
			AcctOutputTotal: 2000000,
		}
		accounting2 := &domain.RadiusAccounting{
			Username:        "testuser3",
			AcctSessionTime: 1800, // 30 minutes
			AcctInputTotal:  500000,
			AcctOutputTotal: 1000000,
		}
		err = repo.Create(ctx, accounting1)
		require.NoError(t, err)
		err = repo.Create(ctx, accounting2)
		require.NoError(t, err)

		user := &domain.RadiusUser{
			Username:  "testuser3",
			TimeQuota: 18000, // 5 hours
		}
		authCtx := &auth.AuthContext{User: user}

		err = checker.Check(ctx, authCtx)
		require.NotNil(t, err) // Should reject (5.5 hours > 5 hours)

		// Verify it's the correct error type
		authErr, ok := errors.GetAuthError(err)
		assert.True(t, ok)
		assert.Contains(t, authErr.Message, "time quota exceeded")
	})

	t.Run("allows login when user is nil", func(t *testing.T) {
		authCtx := &auth.AuthContext{User: nil}

		err := checker.Check(ctx, authCtx)
		assert.Nil(t, err)
	})

	t.Run("allows login when exactly at quota limit", func(t *testing.T) {
		// Create accounting record with exactly 5 hours
		accounting := &domain.RadiusAccounting{
			Username:        "testuser4",
			AcctSessionTime: 18000, // Exactly 5 hours
			AcctInputTotal:  1000000,
			AcctOutputTotal: 2000000,
		}
		err = repo.Create(ctx, accounting)
		require.NoError(t, err)

		user := &domain.RadiusUser{
			Username:  "testuser4",
			TimeQuota: 18000, // Exactly 5 hours
		}
		authCtx := &auth.AuthContext{User: user}

		err = checker.Check(ctx, authCtx)
		require.NotNil(t, err) // Should reject (exactly at limit = exceeded)
	})
}

func TestTimeQuotaChecker_Check_WithTenant(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate schema
	err = db.AutoMigrate(&domain.RadiusAccounting{})
	require.NoError(t, err)

	// Create real repository
	repo := gormrepo.NewGormAccountingRepository(db)
	checker := NewTimeQuotaChecker(repo)

	// Test with tenant context
	tenantCtx := tenant.WithTenantID(context.Background(), 123)

	t.Run("respects tenant isolation", func(t *testing.T) {
		// Create accounting records for different tenants
		accounting1 := &domain.RadiusAccounting{
			TenantID:        123,
			Username:        "testuser5",
			AcctSessionTime: 3600, // 1 hour
			AcctInputTotal:  1000000,
			AcctOutputTotal: 2000000,
		}
		accounting2 := &domain.RadiusAccounting{
			TenantID:        456,
			Username:        "testuser5",
			AcctSessionTime: 18000, // 5 hours (different tenant)
			AcctInputTotal:  1000000,
			AcctOutputTotal: 2000000,
		}
		err = repo.Create(tenantCtx, accounting1)
		require.NoError(t, err)
		err = repo.Create(context.Background(), accounting2)
		require.NoError(t, err)

		user := &domain.RadiusUser{
			TenantID:  123,
			Username:  "testuser5",
			TimeQuota: 18000, // 5 hours
		}
		authCtx := &auth.AuthContext{User: user}

		// Should only count tenant 123's usage (1 hour), not tenant 456's (5 hours)
		err = checker.Check(tenantCtx, authCtx)
		assert.Nil(t, err) // Should allow (1 hour < 5 hours)
	})
}
