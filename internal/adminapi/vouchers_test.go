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
		&domain.Product{},
		&domain.VoucherTopup{},
		&domain.VoucherSubscription{},
		&domain.VoucherBundle{},
		&domain.VoucherBundleItem{},
		&domain.RadiusUser{},
		&domain.SysOpr{},
		&domain.SysOprLog{},

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
// P3: PIN Protection Tests
// =============================================================================

func TestPINProtection(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup: Profile, Product, Batch with PIN generation
	profile := domain.RadiusProfile{Name: "PINTestProfile", AddrPool: "pool1"}
	db.Create(&profile)

	product := domain.Product{Name: "PINTestProduct", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	db.Create(&product)

	// Create batch with PIN generation enabled
	batch := domain.VoucherBatch{
		Name:         "PINBatch",
		ProductID:    product.ID,
		Count:        2,
		GeneratePIN:  true,
		PINLength:    4,
		ExpirationType: "fixed",
	}
	db.Create(&batch)

	// Create voucher with PIN
	voucherWithPIN := domain.Voucher{
		BatchID:     batch.ID,
		Code:        "PINTEST001",
		Status:      "unused",
		Price:       10,
		RequirePIN:  true,
		PIN:         "1234",
	}
	db.Create(&voucherWithPIN)

	// Create voucher without PIN (should still work)
	voucherNoPIN := domain.Voucher{
		BatchID:    batch.ID,
		Code:      "PINTEST002",
		Status:    "unused",
		Price:     10,
		RequirePIN: false,
	}
	db.Create(&voucherNoPIN)

	t.Run("Redeem with valid PIN", func(t *testing.T) {
		req := VoucherRedeemRequest{Code: "PINTEST001", PIN: "1234"}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/redeem", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, RedeemVoucher(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify user created
			var user domain.RadiusUser
			db.First(&user, "username = ?", "PINTEST001")
			assert.Equal(t, "PINTEST001", user.Username)

			// Verify voucher FirstUsedAt is set
			var updatedVoucher domain.Voucher
			db.First(&updatedVoucher, "code = ?", "PINTEST001")
			assert.Equal(t, "used", updatedVoucher.Status)
			assert.False(t, updatedVoucher.FirstUsedAt.IsZero())
		}
	})

	t.Run("Redeem without PIN when required", func(t *testing.T) {
		// Create a new voucher with PIN for this test
		voucherNoPINUse := domain.Voucher{
			BatchID:    batch.ID,
			Code:       "PINTEST004",
			Status:     "unused",
			Price:      10,
			RequirePIN: true,
			PIN:        "9999",
		}
		db.Create(&voucherNoPINUse)

		req := VoucherRedeemRequest{Code: "PINTEST004"}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/redeem", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = RedeemVoucher(c)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PIN_REQUIRED")
	})

	t.Run("Redeem with invalid PIN", func(t *testing.T) {
		// Create another voucher with PIN
		voucher := domain.Voucher{
			BatchID:    batch.ID,
			Code:       "PINTEST003",
			Status:     "unused",
			Price:      10,
			RequirePIN: true,
			PIN:        "5678",
		}
		db.Create(&voucher)

		req := VoucherRedeemRequest{Code: "PINTEST003", PIN: "0000"}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/redeem", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = RedeemVoucher(c)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "INVALID_PIN")
	})

	t.Run("Redeem voucher without PIN required", func(t *testing.T) {
		req := VoucherRedeemRequest{Code: "PINTEST002"}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/redeem", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, RedeemVoucher(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})
}

// =============================================================================
// P4: Data Top-Up Tests
// =============================================================================

func TestVoucherTopup(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup: Profile, Product, Batch, Voucher (active)
	profile := domain.RadiusProfile{Name: "TopupProfile", AddrPool: "pool1"}
	db.Create(&profile)

	product := domain.Product{Name: "TopupProduct", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	db.Create(&product)

	batch := domain.VoucherBatch{Name: "TopupBatch", ProductID: product.ID, Count: 2}
	db.Create(&batch)

	// Active voucher for topup
	activeVoucher := domain.Voucher{
		BatchID:    batch.ID,
		Code:       "TOPUP001",
		Status:     "active",
		Price:      10,
		ExpireTime: time.Now().Add(24 * time.Hour),
	}
	db.Create(&activeVoucher)

	// Used voucher (not eligible for topup)
	usedVoucher := domain.Voucher{
		BatchID:    batch.ID,
		Code:       "TOPUP002",
		Status:     "used",
		Price:      10,
	}
	db.Create(&usedVoucher)

	// Operator for auth
	operator := &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"}

	t.Run("Create Topup Success", func(t *testing.T) {
		req := VoucherTopupRequest{
			VoucherCode: "TOPUP001",
			DataQuota:   1024, // 1GB
			TimeQuota:   3600,  // 1 hour
			Price:       5.0,
		}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/topup", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, CreateVoucherTopup(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify topup created
			var topup domain.VoucherTopup
			db.First(&topup, "voucher_code = ?", "TOPUP001")
			assert.Equal(t, int64(1024), topup.DataQuota)
			assert.Equal(t, int64(3600), topup.TimeQuota)
			assert.Equal(t, "active", topup.Status)

			// Verify voucher expiry extended
			var voucher domain.Voucher
			db.First(&voucher, "code = ?", "TOPUP001")
			assert.True(t, voucher.ExpireTime.After(activeVoucher.ExpireTime))
		}
	})

	t.Run("Topup for non-existent voucher fails", func(t *testing.T) {
		req := VoucherTopupRequest{
			VoucherCode: "NONEXISTENT",
			DataQuota:   1024,
		}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/topup", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = CreateVoucherTopup(c)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Topup for inactive voucher fails", func(t *testing.T) {
		req := VoucherTopupRequest{
			VoucherCode: "TOPUP002",
			DataQuota:   1024,
		}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/topup", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = CreateVoucherTopup(c)
		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Contains(t, rec.Body.String(), "VOUCHER_NOT_ACTIVE")
	})

	t.Run("List Topups", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/vouchers/topups?voucher_code=TOPUP001", nil), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, ListVoucherTopups(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)
			data := response["data"].([]interface{})
			assert.Len(t, data, 1)
		}
	})

	t.Run("List Topups without voucher_code fails", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/vouchers/topups", nil), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		_ = ListVoucherTopups(c)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// =============================================================================
// P5: First-Use Expiration Tests
// =============================================================================

func TestFirstUseExpiration(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup: Profile, Product, Batch with first_use expiration
	profile := domain.RadiusProfile{Name: "FirstUseProfile", AddrPool: "pool1"}
	db.Create(&profile)

	product := domain.Product{Name: "FirstUseProduct", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	db.Create(&product)

	// Batch with first_use expiration (valid for 7 days after first use)
	batchFirstUse := domain.VoucherBatch{
		Name:           "FirstUseBatch",
		ProductID:      product.ID,
		Count:          2,
		ExpirationType: "first_use",
		ValidityDays:   7,
	}
	db.Create(&batchFirstUse)

	// Batch with fixed expiration
	batchFixed := domain.VoucherBatch{
		Name:           "FixedBatch",
		ProductID:      product.ID,
		Count:          1,
		ExpirationType: "fixed",
	}
	db.Create(&batchFixed)

	// Voucher in first_use batch
	voucherFirstUse := domain.Voucher{
		BatchID: batchFirstUse.ID,
		Code:   "FIRSTUSE001",
		Status: "unused",
		Price:  10,
	}
	db.Create(&voucherFirstUse)

	// Voucher in fixed batch
	voucherFixed := domain.Voucher{
		BatchID: batchFixed.ID,
		Code:   "FIXED001",
		Status: "unused",
		Price:  10,
	}
	db.Create(&voucherFixed)

	t.Run("Redeem first_use voucher calculates expiry from first use", func(t *testing.T) {
		req := VoucherRedeemRequest{Code: "FIRSTUSE001"}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/redeem", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, RedeemVoucher(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify user has expiry 7 days from now
			var user domain.RadiusUser
			db.First(&user, "username = ?", "FIRSTUSE001")
			expectedExpire := time.Now().Add(7 * 24 * time.Hour)
			assert.WithinDuration(t, expectedExpire, user.ExpireTime, time.Minute)

			// Verify voucher FirstUsedAt is set
			var voucher domain.Voucher
			db.First(&voucher, "code = ?", "FIRSTUSE001")
			assert.False(t, voucher.FirstUsedAt.IsZero())
		}
	})

	t.Run("Redeem fixed voucher uses product validity", func(t *testing.T) {
		req := VoucherRedeemRequest{Code: "FIXED001"}
		jsonReq, _ := json.Marshal(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/vouchers/redeem", strings.NewReader(string(jsonReq))), rec)
		c.Set("db", db)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, RedeemVoucher(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify user has expiry based on product validity (1 hour = 3600 seconds)
			var user domain.RadiusUser
			db.First(&user, "username = ?", "FIXED001")
			expectedExpire := time.Now().Add(3600 * time.Second)
			assert.WithinDuration(t, expectedExpire, user.ExpireTime, time.Minute)
		}
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
