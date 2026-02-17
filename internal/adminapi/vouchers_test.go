package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

// Helper to setup DB with voucher tables since test_helpers.go doesn't include them
func setupTestDBForVouchers(t *testing.T) *gorm.DB {
	db := setupTestDB(t)
	err := db.AutoMigrate(
		&domain.VoucherBatch{},
		&domain.Voucher{},
		&domain.AgentWallet{},
		&domain.WalletLog{},
		&domain.Product{}, // Need product for tests
	)
	assert.NoError(t, err)
	return db
}

func TestCreateVoucherBatch(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Create a product
	product := domain.Product{
		Name:            "Test Product",
		Price:           10.0,
		CostPrice:       8.0,
		ValiditySeconds: 3600,
		Status:          "enabled",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	db.Create(&product)

	// Create an agent and wallet
	agent := domain.SysOpr{
		Username: "agent1",
		Password: "password",
		Level:    "agent",
		Status:   "enabled",
	}
	db.Create(&agent)

	wallet := domain.AgentWallet{
		AgentID: agent.ID,
		Balance: 100.0,
	}
	db.Create(&wallet)

	t.Run("Create Batch Success (Admin)", func(t *testing.T) {
		req := VoucherBatchRequest{
			Name:      "Admin Batch",
			ProductID: "1", // Assuming product ID 1
			Count:     10,
			Prefix:    "ADM",
			Length:    12,
			Type:      "mixed",
		}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/voucher-batches", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"})
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, CreateVoucherBatch(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("Create Batch Success (Agent with Funds)", func(t *testing.T) {
		req := VoucherBatchRequest{
			Name:      "Agent Batch",
			ProductID: "1",
			Count:     5,
			Prefix:    "AGT",
			Length:    12,
			Type:      "mixed",
		}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/voucher-batches", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		// Set context operator to the agent created above
		c.Set("current_operator", &agent)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, CreateVoucherBatch(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			
			// Verify Wallet Deduction
			var updatedWallet domain.AgentWallet
			db.First(&updatedWallet, "agent_id = ?", agent.ID)
			expectedBalance := 100.0 - (5 * 8.0) // 5 vouchers * cost 8.0
			assert.Equal(t, expectedBalance, updatedWallet.Balance)
		}
	})
}

func TestVoucherLifecycle(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup: Profile, Product, Batch, Vouchers
	profile := domain.RadiusProfile{Name: "Basic", AddrPool: "pool1"}
	db.Create(&profile)
	
	product := domain.Product{Name: "1Hour", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	db.Create(&product)

	batch := domain.VoucherBatch{Name: "LifeBatch", ProductID: product.ID, Count: 2}
	db.Create(&batch)

	voucher := domain.Voucher{
		BatchID: batch.ID,
		Code:    "TESTCODE123",
		Status:  "unused",
		Price:   10,
	}
	db.Create(&voucher)

	t.Run("Redeem Voucher", func(t *testing.T) {
		req := VoucherRedeemRequest{Code: "TESTCODE123"}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/redeem", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, RedeemVoucher(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			
			// Verify User Created
			var user domain.RadiusUser
			db.First(&user, "username = ?", "TESTCODE123")
			assert.Equal(t, "TESTCODE123", user.Username)
			assert.Equal(t, "enabled", user.Status)
			
			// Verify Voucher Status Updated
			var updatedVoucher domain.Voucher
			db.First(&updatedVoucher, "code = ?", "TESTCODE123")
			assert.Equal(t, "used", updatedVoucher.Status)
		}
	})

	t.Run("Redeem Invalid Voucher", func(t *testing.T) {
		req := VoucherRedeemRequest{Code: "INVALID"}
        jsonReq, _ := json.Marshal(req)

        rec := httptest.NewRecorder()
        c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/redeem", strings.NewReader(string(jsonReq))), rec)
        c.Set("db", db)
        c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

        _ = RedeemVoucher(c)
        assert.Equal(t, http.StatusNotFound, rec.Code)
    })
}

func TestExtendVoucher(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup: Profile, Product, Batch, Voucher, User
	profile := domain.RadiusProfile{Name: "TestProfile", AddrPool: "pool1"}
	db.Create(&profile)

	product := domain.Product{Name: "TestProduct", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	db.Create(&product)

	batch := domain.VoucherBatch{Name: "ExtendTest", ProductID: product.ID, Count: 1}
	db.Create(&batch)

	// Create a used voucher
	expireTime := time.Now().Add(24 * time.Hour)
	voucher := domain.Voucher{
		BatchID:    batch.ID,
		Code:       "EXTENDTEST001",
		Status:     "used",
		Price:      10,
		ExpireTime: expireTime,
	}
	db.Create(&voucher)

	// Create associated user
	user := domain.RadiusUser{
		Username:   "EXTENDTEST001",
		Password:   "EXTENDTEST001",
		ProfileId:  profile.ID,
		Status:     "enabled",
		ExpireTime: expireTime,
	}
	db.Create(&user)

	t.Run("Extend Voucher Success", func(t *testing.T) {
		req := VoucherExtendRequest{Code: "EXTENDTEST001", ValidityDays: 7}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/extend", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"})
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, ExtendVoucher(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify user expiration extended
			var updatedUser domain.RadiusUser
			db.First(&updatedUser, "username = ?", "EXTENDTEST001")
			expectedExpire := expireTime.Add(7 * 24 * time.Hour)
			assert.WithinDuration(t, expectedExpire, updatedUser.ExpireTime, time.Minute)

			// Verify voucher extension count
			var updatedVoucher domain.Voucher
			db.First(&updatedVoucher, "code = ?", "EXTENDTEST001")
			assert.Equal(t, 1, updatedVoucher.ExtendedCount)
		}
	})

	t.Run("Extend Invalid Voucher", func(t *testing.T) {
		req := VoucherExtendRequest{Code: "INVALID", ValidityDays: 7}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/extend", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = ExtendVoucher(c)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Extend Unused Voucher Fails", func(t *testing.T) {
		// Create unused voucher
		unusedVoucher := domain.Voucher{
			BatchID: batch.ID,
			Code:    "UNUSED001",
			Status:  "unused",
			Price:   10,
		}
		db.Create(&unusedVoucher)

		req := VoucherExtendRequest{Code: "UNUSED001", ValidityDays: 7}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/extend", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = ExtendVoucher(c)
		assert.Equal(t, http.StatusConflict, rec.Code)
	})
}
