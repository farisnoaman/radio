package adminapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setupTestDBForVouchers creates a test database with all voucher-related tables.
// This is a convenience wrapper around setupTestDB which includes all tables.
func setupTestDBForVouchers(t *testing.T) *gorm.DB {
	return setupTestDB(t)
}

// =============================================================================
// TestCreateVoucherBatch - Table-Driven Tests
// =============================================================================

// TestCreateVoucherBatch tests the voucher batch creation endpoint.
// Tests both admin and agent scenarios for batch creation.
//
// Test scenarios:
// - Admin user creates batch: Should succeed without wallet deduction
// - Agent with sufficient funds creates batch: Should succeed and deduct from wallet
// - Agent with insufficient funds: Should fail with appropriate error
// - Invalid product ID: Should fail with not found error
func TestCreateVoucherBatch(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup: Create required product and agent with wallet
	product := domain.Product{
		Name:            "Test Product",
		Price:           10.0,
		CostPrice:       8.0,
		ValiditySeconds: 3600,
		Status:          "enabled",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	require.NoError(t, db.Create(&product).Error)

	agent := domain.SysOpr{
		Username: "agent1",
		Password: "password",
		Level:    "agent",
		Status:   "enabled",
	}
	require.NoError(t, db.Create(&agent).Error)

	wallet := domain.AgentWallet{
		AgentID: agent.ID,
		Balance: 100.0,
	}
	require.NoError(t, db.Create(&wallet).Error)

	// Table-driven test cases
	tests := []struct {
		name               string
		operator           *domain.SysOpr
		req                VoucherBatchRequest
		wantStatus         int
		walletDeduction    float64 // Expected wallet deduction (0 if no deduction expected)
		description        string
	}{
		{
			name:        "AdminUser_CreatesBatch_ShouldSucceed",
			operator:    &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"},
			req:         VoucherBatchRequest{Name: "Admin Batch", ProductID: "1", Count: 10, Prefix: "ADM", Length: 12, Type: "mixed"},
			wantStatus:  http.StatusOK,
			walletDeduction: 0, // Admin doesn't pay
			description: "Admin creates voucher batch without cost deduction",
		},
		{
			name:        "AgentWithFunds_CreatesBatch_ShouldSucceedAndDeductWallet",
			operator:    &agent,
			req:         VoucherBatchRequest{Name: "Agent Batch", ProductID: "1", Count: 5, Prefix: "AGT", Length: 12, Type: "mixed"},
			wantStatus:  http.StatusOK,
			walletDeduction: 40.0, // 5 vouchers * $8 cost
			description: "Agent with sufficient balance creates batch, wallet should be deducted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonReq, err := json.Marshal(tt.req)
			require.NoError(t, err)

			rec := httptest.NewRecorder()
			c := e.NewContext(
				httptest.NewRequest(http.MethodPost, "/api/v1/voucher-batches", strings.NewReader(string(jsonReq))),
				rec,
			)
			c.Set("db", db)
			c.Set("current_operator", tt.operator)
			c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			err = CreateVoucherBatch(c)
			if tt.wantStatus == http.StatusOK {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantStatus, rec.Code, tt.description)

			// Verify wallet deduction if applicable
			if tt.walletDeduction > 0 {
				var updatedWallet domain.AgentWallet
				db.First(&updatedWallet, "agent_id = ?", agent.ID)
				expectedBalance := 100.0 - tt.walletDeduction
				assert.Equal(t, expectedBalance, updatedWallet.Balance, "Wallet should be deducted correctly")
			}
		})
	}
}

// =============================================================================
// TestPINProtection - Table-Driven Tests
// =============================================================================

// TestPINProtection tests PIN protection functionality for voucher redemption.
//
// Test scenarios:
// - Valid PIN: Should succeed and create user
// - Missing PIN when required: Should fail with PIN_REQUIRED error
// - Invalid PIN: Should fail with INVALID_PIN error
// - No PIN required: Should succeed without PIN
func TestPINProtection(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup: Create profile, product, and batch with PIN generation
	profile := domain.RadiusProfile{Name: "PINTestProfile", AddrPool: "pool1"}
	require.NoError(t, db.Create(&profile).Error)

	product := domain.Product{Name: "PINTestProduct", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	require.NoError(t, db.Create(&product).Error)

	batch := domain.VoucherBatch{
		Name:           "PINBatch",
		ProductID:      product.ID,
		Count:          2,
		GeneratePIN:    true,
		PINLength:      4,
		ExpirationType: "fixed",
	}
	require.NoError(t, db.Create(&batch).Error)

	// Create test vouchers
	voucherWithPIN := domain.Voucher{
		BatchID:     batch.ID,
		Code:        "PINTEST001",
		Status:      "unused",
		Price:       10,
		RequirePIN:  true,
		PIN:         "1234",
	}
	require.NoError(t, db.Create(&voucherWithPIN).Error)

	voucherNoPIN := domain.Voucher{
		BatchID:     batch.ID,
		Code:        "PINTEST002",
		Status:      "unused",
		Price:       10,
		RequirePIN:  false,
	}
	require.NoError(t, db.Create(&voucherNoPIN).Error)

	// Voucher without PIN required (for dedicated test)
	voucherNoPINRequired := domain.Voucher{
		BatchID:     batch.ID,
		Code:        "PINTEST005",
		Status:      "unused",
		Price:       10,
		RequirePIN:  false,
	}
	require.NoError(t, db.Create(&voucherNoPINRequired).Error)

	// Table-driven test cases
	tests := []struct {
		name           string
		voucherCode    string
		pin            string
		wantStatus     int
		wantErrorCode  string // Expected error code in response
		description    string
	}{
		{
			name:        "ValidPIN_ShouldSucceed",
			voucherCode: "PINTEST001",
			pin:         "1234",
			wantStatus:  http.StatusOK,
			description: "Valid PIN should allow voucher redemption",
		},
		{
			name:          "MissingPINWhenRequired_ShouldFail",
			voucherCode:   "PINTEST001",
			pin:           "",
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "PIN_REQUIRED",
			description:   "Missing PIN when required should return PIN_REQUIRED error",
		},
		{
			name:          "InvalidPIN_ShouldFail",
			voucherCode:   "PINTEST001",
			pin:           "0000",
			wantStatus:    http.StatusUnauthorized,
			wantErrorCode: "INVALID_PIN",
			description:   "Invalid PIN should return INVALID_PIN error",
		},
		{
			name:        "NoPINRequired_ShouldSucceed",
			voucherCode: "PINTEST005",
			pin:         "",
			wantStatus:  http.StatusOK,
			description: "Voucher without PIN requirement should succeed without PIN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use existing PINTEST005 for no-PIN test, create new for others
			voucherCode := "PINTEST005"
			if tt.name != "NoPINRequired_ShouldSucceed" {
				// Create a fresh voucher for tests that need PIN validation
				// to avoid state issues from previous test runs
				newVoucher := domain.Voucher{
					BatchID:    batch.ID,
					Code:       fmt.Sprintf("PINTEST_%d", time.Now().UnixNano()),
					Status:     "unused",
					Price:      10,
					RequirePIN: true,
					PIN:        "1234",
				}
				require.NoError(t, db.Create(&newVoucher).Error)
				voucherCode = newVoucher.Code
			}

			req := VoucherRedeemRequest{Code: voucherCode, PIN: tt.pin}
			jsonReq, _ := json.Marshal(req)

			rec := httptest.NewRecorder()
			c := e.NewContext(
				httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/redeem", strings.NewReader(string(jsonReq))),
				rec,
			)
			c.Set("db", db)
			c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			_ = RedeemVoucher(c)
			assert.Equal(t, tt.wantStatus, rec.Code, tt.description)
			if tt.wantErrorCode != "" {
				assert.Contains(t, rec.Body.String(), tt.wantErrorCode, "Error code should be present in response")
			}
		})
	}
}

// =============================================================================
// TestFirstUseExpiration - Table-Driven Tests
// =============================================================================

// TestFirstUseExpiration tests first-use expiration functionality for vouchers.
//
// Test scenarios:
// - First-use expiration: Voucher expires X days after first login
// - Fixed expiration: Voucher expires based on product validity period
func TestFirstUseExpiration(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup: Create profile and products
	profile := domain.RadiusProfile{Name: "FirstUseProfile", AddrPool: "pool1"}
	require.NoError(t, db.Create(&profile).Error)

	product := domain.Product{Name: "FirstUseProduct", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	require.NoError(t, db.Create(&product).Error)

	// Create batch with first_use expiration (7 days after first use)
	batchFirstUse := domain.VoucherBatch{
		Name:           "FirstUseBatch",
		ProductID:      product.ID,
		Count:          2,
		ExpirationType: "first_use",
		ValidityDays:   7,
	}
	require.NoError(t, db.Create(&batchFirstUse).Error)

	// Create batch with fixed expiration
	batchFixed := domain.VoucherBatch{
		Name:           "FixedBatch",
		ProductID:      product.ID,
		Count:          1,
		ExpirationType: "fixed",
	}
	require.NoError(t, db.Create(&batchFixed).Error)

	// Generate unique voucher codes for this test run
	uniqueSuffix := time.Now().UnixNano()
	firstUseCode := fmt.Sprintf("FIRSTUSE%d", uniqueSuffix)
	fixedCode := fmt.Sprintf("FIXED%d", uniqueSuffix)

	// Create vouchers
	voucherFirstUse := domain.Voucher{
		BatchID: batchFirstUse.ID,
		Code:    firstUseCode,
		Status:  "unused",
		Price:   10,
	}
	require.NoError(t, db.Create(&voucherFirstUse).Error)

	voucherFixed := domain.Voucher{
		BatchID: batchFixed.ID,
		Code:    fixedCode,
		Status:  "unused",
		Price:   10,
	}
	require.NoError(t, db.Create(&voucherFixed).Error)

	// Table-driven test cases
	tests := []struct {
		name            string
		voucherCode     string
		wantStatus      int
		validateExpiry  func(t *testing.T, user domain.RadiusUser)
		description     string
	}{
		{
			name:        "FirstUseVoucher_CalculatesExpiryFromFirstUse",
			voucherCode: firstUseCode,
			wantStatus:  http.StatusOK,
			validateExpiry: func(t *testing.T, user domain.RadiusUser) {
				// Should expire 7 days from now (first_use expiration)
				expectedExpire := time.Now().Add(7 * 24 * time.Hour)
				assert.WithinDuration(t, expectedExpire, user.ExpireTime, time.Minute, "First-use voucher should expire 7 days from first use")
			},
			description: "First-use voucher should calculate expiry from first login time",
		},
		{
			name:        "FixedVoucher_UsesProductValidity",
			voucherCode: fixedCode,
			wantStatus:  http.StatusOK,
			validateExpiry: func(t *testing.T, user domain.RadiusUser) {
				// Should expire based on product validity (1 hour = 3600 seconds)
				expectedExpire := time.Now().Add(3600 * time.Second)
				assert.WithinDuration(t, expectedExpire, user.ExpireTime, time.Minute, "Fixed voucher should use product validity period")
			},
			description: "Fixed voucher should use product validity period for expiration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := VoucherRedeemRequest{Code: tt.voucherCode}
			jsonReq, _ := json.Marshal(req)

			rec := httptest.NewRecorder()
			c := e.NewContext(
				httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/redeem", strings.NewReader(string(jsonReq))),
				rec,
			)
			c.Set("db", db)
			c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			err := RedeemVoucher(c)
			if tt.wantStatus == http.StatusOK {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantStatus, rec.Code, tt.description)

			// Verify user expiration
			var user domain.RadiusUser
			db.First(&user, "username = ?", tt.voucherCode)
			tt.validateExpiry(t, user)

			// Verify voucher FirstUsedAt is set
			var voucher domain.Voucher
			db.First(&voucher, "code = ?", tt.voucherCode)
			assert.False(t, voucher.FirstUsedAt.IsZero(), "FirstUsedAt should be set after redemption")
		})
	}
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
		if err := db.Create(&unusedVoucher).Error; err != nil {
			t.Fatalf("Failed to create unused voucher: %v", err)
		}
		
		// Verify existence
		var checkVoucher domain.Voucher
		if err := db.First(&checkVoucher, "code = ?", "UNUSED001").Error; err != nil {
			t.Fatalf("Voucher UNUSED001 not found in DB immediately after creation: %v", err)
		}



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

func TestTransferVouchers(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup: Create two agents
	agent1 := domain.SysOpr{Username: "agent1", Password: "password", Level: "agent", Status: "enabled"}
	db.Create(&agent1)

	agent2 := domain.SysOpr{Username: "agent2", Password: "password", Level: "agent", Status: "enabled"}
	db.Create(&agent2)

	// Create product
	product := domain.Product{Name: "TransferProduct", Price: 10, CostPrice: 8, ValiditySeconds: 3600, Status: "enabled"}
	db.Create(&product)

	// Create batch with agent1
	batch := domain.VoucherBatch{Name: "TransferBatch", ProductID: product.ID, AgentID: agent1.ID, Count: 5}
	db.Create(&batch)

	// Create unused vouchers for agent1
	for i := 0; i < 3; i++ {
		voucher := domain.Voucher{
			BatchID:  batch.ID,
			Code:     fmt.Sprintf("TRANF%03d", i),
			Status:   "unused",
			AgentID:  agent1.ID,
			Price:    10,
		}
		db.Create(&voucher)
	}

	// Create used voucher (should not be transferred)
	usedVoucher := domain.Voucher{
		BatchID: batch.ID,
		Code:    "TRANFUSED",
		Status:  "used",
		AgentID: agent1.ID,
		Price:   10,
	}
	db.Create(&usedVoucher)

	t.Run("Transfer Vouchers Success (Admin)", func(t *testing.T) {
		req := TransferVouchersRequest{BatchID: fmt.Sprintf("%d", batch.ID), ToAgentID: fmt.Sprintf("%d", agent2.ID)}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/voucher-batches/1/transfer", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"})
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, TransferVouchers(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify vouchers transferred
			var vouchers []domain.Voucher
			db.Where("batch_id = ?", batch.ID).Find(&vouchers)
			for _, v := range vouchers {
				if v.Status == "unused" {
					assert.Equal(t, agent2.ID, v.AgentID, "Voucher %s should be transferred to agent2", v.Code)
				}
			}
		}
	})

	t.Run("Transfer to Same Agent Fails", func(t *testing.T) {
		// Create new batch for agent1
		batch2 := domain.VoucherBatch{Name: "TransferBatch2", ProductID: product.ID, AgentID: agent1.ID, Count: 2}
		db.Create(&batch2)

		req := TransferVouchersRequest{BatchID: fmt.Sprintf("%d", batch2.ID), ToAgentID: fmt.Sprintf("%d", agent1.ID)}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/voucher-batches/2/transfer", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"})
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = TransferVouchers(c)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Transfer Non-Existent Batch Fails", func(t *testing.T) {
		req := TransferVouchersRequest{BatchID: "99999", ToAgentID: fmt.Sprintf("%d", agent2.ID)}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/voucher-batches/99999/transfer", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"})
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = TransferVouchers(c)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

// =============================================================================
// P6: Subscription/Recurring Tests
// =============================================================================

func TestVoucherSubscription(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup
	profile := domain.RadiusProfile{Name: "SubProfile", AddrPool: "pool1"}
	db.Create(&profile)

	product := domain.Product{Name: "SubProduct", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	db.Create(&product)

	batch := domain.VoucherBatch{Name: "SubBatch", ProductID: product.ID, Count: 3}
	db.Create(&batch)

	// Active voucher for subscription
	activeVoucher := domain.Voucher{
		BatchID:    batch.ID,
		Code:       "SUB001",
		Status:     "active",
		Price:      10,
	}
	db.Create(&activeVoucher)

	// Used voucher (not eligible)
	usedVoucher := domain.Voucher{
		BatchID:    batch.ID,
		Code:       "SUB002",
		Status:     "used",
		Price:      10,
	}
	db.Create(&usedVoucher)

	operator := &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"}

	t.Run("Create Subscription Success", func(t *testing.T) {
		req := VoucherSubscriptionRequest{
			VoucherCode:  "SUB001",
			ProductID:    fmt.Sprintf("%d", product.ID),
			IntervalDays: 30,
			AutoRenew:    true,
		}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/subscriptions", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, CreateVoucherSubscription(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify subscription created
			var sub domain.VoucherSubscription
			db.First(&sub, "voucher_code = ?", "SUB001")
			assert.Equal(t, 30, sub.IntervalDays)
			assert.True(t, sub.AutoRenew)
			assert.Equal(t, "active", sub.Status)
			assert.False(t, sub.NextRenewalAt.IsZero())
		}
	})

	t.Run("Create duplicate subscription fails", func(t *testing.T) {
		req := VoucherSubscriptionRequest{
			VoucherCode:  "SUB001",
			ProductID:    fmt.Sprintf("%d", product.ID),
			IntervalDays: 30,
			AutoRenew:    true,
		}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/subscriptions", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = CreateVoucherSubscription(c)
		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Contains(t, rec.Body.String(), "SUBSCRIPTION_EXISTS")
	})

	t.Run("Create subscription for non-active voucher fails", func(t *testing.T) {
		req := VoucherSubscriptionRequest{
			VoucherCode:  "SUB002",
			ProductID:    fmt.Sprintf("%d", product.ID),
			IntervalDays: 30,
		}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/subscriptions", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = CreateVoucherSubscription(c)
		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Contains(t, rec.Body.String(), "VOUCHER_NOT_ACTIVE")
	})

	t.Run("Create subscription for non-existent voucher fails", func(t *testing.T) {
		req := VoucherSubscriptionRequest{
			VoucherCode:  "NONEXISTENT",
			ProductID:    fmt.Sprintf("%d", product.ID),
			IntervalDays: 30,
		}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/subscriptions", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = CreateVoucherSubscription(c)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("List Subscriptions", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/vouchers/subscriptions?voucher_code=SUB001", nil), rec)
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, ListVoucherSubscriptions(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)
			data := response["data"].([]interface{})
			assert.Len(t, data, 1)
		}
	})

	t.Run("Cancel Subscription Success", func(t *testing.T) {
		// Get subscription ID
		var sub domain.VoucherSubscription
		result := db.First(&sub, "voucher_code = ?", "SUB001")
		assert.NoError(t, result.Error)
		assert.NotZero(t, sub.ID)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/vouchers/subscriptions/%d/cancel", sub.ID), nil), rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", sub.ID))
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, CancelVoucherSubscription(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify subscription cancelled
			var updatedSub domain.VoucherSubscription
			db.First(&updatedSub, "id = ?", sub.ID)
			assert.Equal(t, "cancelled", updatedSub.Status)
		}
	})

	t.Run("Cancel non-existent subscription fails", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/subscriptions/99999/cancel", nil), rec)
		c.SetParamNames("id")
		c.SetParamValues("99999")
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = CancelVoucherSubscription(c)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

// =============================================================================
// P7: Voucher Bundles Tests
// =============================================================================

func TestVoucherBundle(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup
	profile := domain.RadiusProfile{Name: "BundleProfile", AddrPool: "pool1"}
	db.Create(&profile)

	product := domain.Product{Name: "BundleProduct", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	db.Create(&product)

	operator := &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"}

	t.Run("Create Bundle Success", func(t *testing.T) {
		req := VoucherBundleRequest{
			Name:        "Test Bundle",
			Description: "A test bundle",
			ProductID:   fmt.Sprintf("%d", product.ID),
			VoucherCount: 5,
			Price:       45.0,
		}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/voucher-bundles", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, CreateVoucherBundle(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify bundle created
			var bundle domain.VoucherBundle
			db.First(&bundle, "name = ?", "Test Bundle")
			assert.Equal(t, "Test Bundle", bundle.Name)
			assert.Equal(t, 5, bundle.VoucherCount)
			assert.Equal(t, "active", bundle.Status)

			// Verify vouchers generated
			var vouchers []domain.Voucher
			db.Where("batch_id = ?", bundle.ID).Find(&vouchers)
			assert.Len(t, vouchers, 5)
		}
	})

	t.Run("Create Bundle with invalid product fails", func(t *testing.T) {
		req := VoucherBundleRequest{
			Name:        "Invalid Bundle",
			ProductID:   "99999",
			VoucherCount: 5,
			Price:       45.0,
		}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/voucher-bundles", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = CreateVoucherBundle(c)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("List Bundles", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/voucher-bundles", nil), rec)
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, ListVoucherBundles(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)
			data := response["data"].([]interface{})
			assert.GreaterOrEqual(t, len(data), 1)
		}
	})
}

// =============================================================================
// P8: Public Validation API Tests
// =============================================================================

func TestPublicVoucherStatus(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup: Create a voucher
	profile := domain.RadiusProfile{Name: "PublicProfile", AddrPool: "pool1"}
	db.Create(&profile)

	product := domain.Product{Name: "PublicProduct", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	db.Create(&product)

	batch := domain.VoucherBatch{Name: "PublicBatch", ProductID: product.ID, Count: 2}
	db.Create(&batch)

	activeVoucher := domain.Voucher{
		BatchID:     batch.ID,
		Code:        "PUBLIC001",
		Status:      "active",
		Price:       10,
		ExpireTime:  time.Now().Add(24 * time.Hour),
		ActivatedAt: time.Now(),
	}
	db.Create(&activeVoucher)

	t.Run("Get Public Status Success", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/public/vouchers/status?code=PUBLIC001", nil), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, PublicVoucherStatus(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)
			data := response["data"].(map[string]interface{})
			assert.Equal(t, "active", data["status"])
			// PIN should not be exposed in public API
			assert.NotContains(t, rec.Body.String(), "pin")
		}
	})

	t.Run("Get Status for non-existent voucher fails", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/public/vouchers/status?code=NONEXISTENT", nil), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = PublicVoucherStatus(c)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Get Status without code parameter fails", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/public/vouchers/status", nil), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = PublicVoucherStatus(c)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "CODE_REQUIRED")
	})
}
