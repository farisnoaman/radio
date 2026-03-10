package adminapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	customValidator "github.com/talkincode/toughradius/v9/pkg/validator"
	"gorm.io/gorm"
)

// setupTestEcho creates an Echo instance with a validator
func setupTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = customValidator.NewValidator()
	return e
}

// setupTestDB creates an in-memory test database with all required tables
func setupTestDB(t *testing.T) *gorm.DB {
	// Use a file-based temporary database with unique name for each test
	// Include timestamp with nanosecond precision to ensure uniqueness
	uniqueID := time.Now().Format("20060102150405.000000")
	dbPath := "/tmp/testdb_" + t.Name() + "_" + uniqueID + ".db"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)

	// Automatically migrate all tables needed for testing
	err = db.AutoMigrate(
		&domain.RadiusProfile{},
		&domain.RadiusUser{},
		&domain.NetNode{},
		&domain.NetNas{},
		&domain.RadiusAccounting{},
		&domain.RadiusOnline{},
		&domain.SysOpr{},
		&domain.SysConfig{},
		&domain.SysOprLog{},
		&domain.VoucherBatch{},
		&domain.Voucher{},
		&domain.AgentWallet{},
		&domain.WalletLog{},
		&domain.Product{},
		&domain.VoucherTopup{},
		&domain.VoucherSubscription{},
		&domain.VoucherBundle{},
		&domain.VoucherBundleItem{},
		// Agent hierarchy & commission tables for new features
		&domain.AgentHierarchy{},
		&domain.CommissionLog{},
		&domain.CommissionSummary{},
		&domain.PayoutLog{},
		&domain.Invoice{},
		&domain.Server{},
	)
	require.NoError(t, err)

	// Clean up the database file after the test
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	})

	return db
}

// setupTestApp creates a test application context and sets it globally
// Returns: app context for injecting into echo context
func setupTestApp(_ *testing.T, db *gorm.DB) app.AppContext {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestApp",
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/toughradius-test",
			Debug:    true,
		},
		Web: config.WebConfig{
			Secret: "test-secret-key-for-jwt",
		},
		Database: config.DBConfig{
			Type: "sqlite",
			Name: ":memory:",
		},
	}

	// Create application but don't call Init() which would create a new DB
	testApp := app.NewApplication(cfg)
	testApp.Init(cfg)
	testApp.OverrideDB(db)

	return testApp
}

// CreateTestAppContext creates a test application context with an in-memory SQLite database
// Returns: db, echo instance, and app context
func CreateTestAppContext(t *testing.T) (*gorm.DB, *echo.Echo, app.AppContext) {
	// Use unique database path for each test with nanosecond precision
	uniqueID := time.Now().Format("20060102150405.000000")
	dbPath := "/tmp/testdb_" + t.Name() + "_" + uniqueID + ".db"

	cfg := &config.AppConfig{
		System: config.SysConfig{
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/toughradius-test",
		},
		Database: config.DBConfig{
			Type: "sqlite",
			Name: dbPath,
		},
		Web: config.WebConfig{
			Secret: "test-secret-key-for-jwt",
		},
	}

	testApp := app.NewApplication(cfg)
	testApp.Init(cfg)

	// Migrate all test tables
	db := testApp.DB()
	err := db.AutoMigrate(
		&domain.RadiusProfile{},
		&domain.RadiusUser{},
		&domain.NetNode{},
		&domain.NetNas{},
		&domain.RadiusAccounting{},
		&domain.RadiusOnline{},
		&domain.SysOpr{},
		&domain.SysConfig{},
		&domain.SysOprLog{},
		&domain.VoucherBatch{},
		&domain.Voucher{},
		&domain.AgentWallet{},
		&domain.WalletLog{},
		&domain.Product{},
		&domain.VoucherTopup{},
		&domain.VoucherSubscription{},
		&domain.VoucherBundle{},
		&domain.VoucherBundleItem{},
		// Agent hierarchy & commission tables for new features
		&domain.AgentHierarchy{},
		&domain.CommissionLog{},
		&domain.CommissionSummary{},
		&domain.PayoutRequest{},
		&domain.PayoutLog{},
		&domain.Server{},
	)
	require.NoError(t, err)

	// Clean up the database file after the test
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	})

	e := setupTestEcho()

	return db, e, testApp
}

// CreateTestContext creates an echo context with appCtx injected
func CreateTestContext(e *echo.Echo, db *gorm.DB, req *http.Request, rec *httptest.ResponseRecorder, appCtx app.AppContext) echo.Context {
	c := e.NewContext(req, rec)
	c.Set("appCtx", appCtx)
	c.Set("db", db)
	// Inject a default super admin for tests that require authentication
	c.Set("current_operator", &domain.SysOpr{
		ID:       1,
		Username: "superadmin",
		Level:    "super",
		Status:   "enabled",
	})
	return c
}

// CreateTestContextWithApp is a helper that combines setupTestEcho, setupTestDB, and setupTestApp
// for backward compatibility with existing tests
func CreateTestContextWithApp(t *testing.T, req *http.Request, rec *httptest.ResponseRecorder) (echo.Context, *gorm.DB, app.AppContext) {
	e := setupTestEcho()
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	c := CreateTestContext(e, db, req, rec, appCtx)
	return c, db, appCtx
}
