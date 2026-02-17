package adminapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

func TestBulkActivateVouchers(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup: Profile, Product, Batch (unused)
	profile := domain.RadiusProfile{Name: "BulkProfile", AddrPool: "pool1"}
	db.Create(&profile)

	product := domain.Product{Name: "BulkProduct", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	db.Create(&product)

	batch := domain.VoucherBatch{Name: "BulkBatch", ProductID: product.ID, Count: 2, ExpirationType: "fixed"}
	db.Create(&batch)

	voucher1 := domain.Voucher{BatchID: batch.ID, Code: "BULK001", Status: "unused", Price: 10}
	db.Create(&voucher1)

	voucher2 := domain.Voucher{BatchID: batch.ID, Code: "BULK002", Status: "unused", Price: 10}
	db.Create(&voucher2)

	// Admin operator
	operator := &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"}

	t.Run("Bulk Activate Success", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/voucher-batches/%d/activate", batch.ID), nil), rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", batch.ID))
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, BulkActivateVouchers(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify vouchers activated
			var v1, v2 domain.Voucher
			db.First(&v1, "code = ?", "BULK001")
			db.First(&v2, "code = ?", "BULK002")

			assert.Equal(t, "active", v1.Status)
			assert.False(t, v1.ActivatedAt.IsZero())
			assert.False(t, v1.ExpireTime.IsZero()) // Should be set for fixed expiration

			assert.Equal(t, "active", v2.Status)
			assert.False(t, v2.ActivatedAt.IsZero())
			assert.False(t, v2.ExpireTime.IsZero())
		}
	})
}

func TestBulkActivateFirstUseVouchers(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup: Profile, Product, Batch (first_use)
	profile := domain.RadiusProfile{Name: "BulkFirstUseProfile", AddrPool: "pool1"}
	db.Create(&profile)

	product := domain.Product{Name: "BulkFirstUseProduct", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	db.Create(&product)

	batch := domain.VoucherBatch{Name: "BulkFirstUseBatch", ProductID: product.ID, Count: 1, ExpirationType: "first_use", ValidityDays: 7}
	db.Create(&batch)

	voucher := domain.Voucher{BatchID: batch.ID, Code: "BULKFIRST001", Status: "unused", Price: 10}
	db.Create(&voucher)

	operator := &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"}

	t.Run("Bulk Activate First-Use Success", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/voucher-batches/%d/activate", batch.ID), nil), rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", batch.ID))
		c.Set("db", db)
		c.Set("current_operator", operator)
		c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		if assert.NoError(t, BulkActivateVouchers(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Verify voucher activated but no expiry set yet
			var v domain.Voucher
			db.First(&v, "code = ?", "BULKFIRST001")

			assert.Equal(t, "active", v.Status)
			assert.False(t, v.ActivatedAt.IsZero())
			assert.True(t, v.ExpireTime.IsZero()) // Should be zero for first_use untill redemption/login
		}
	})
}

func TestExportVoucherBatch(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForVouchers(t)

	// Setup data
	profile := domain.RadiusProfile{Name: "ExportProfile", AddrPool: "pool1"}
	db.Create(&profile)

	product := domain.Product{Name: "ExportProduct", Price: 10, RadiusProfileID: profile.ID, ValiditySeconds: 3600}
	db.Create(&product)

	batch := domain.VoucherBatch{Name: "ExportBatch", ProductID: product.ID, Count: 2, ExpirationType: "fixed"}
	db.Create(&batch)

	// Active voucher with dates
	activeVoucher := domain.Voucher{
		BatchID:     batch.ID,
		Code:        "EXP001",
		Status:      "active",
		Price:       10,
		ActivatedAt: time.Now(),
		ExpireTime:  time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(&activeVoucher)

	// Unused voucher with zero dates
	unusedVoucher := domain.Voucher{
		BatchID:   batch.ID,
		Code:      "EXP002",
		Status:    "unused",
		Price:     10,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(&unusedVoucher)

	operator := &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"}

	t.Run("Export CSV Success", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/voucher-batches/%d/export", batch.ID), nil), rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", batch.ID))
		c.Set("db", db)
		c.Set("current_operator", operator)

		if assert.NoError(t, ExportVoucherBatch(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "text/csv", rec.Header().Get(echo.HeaderContentType))

			csvContent := rec.Body.String()
			lines := strings.Split(csvContent, "\n")
			
			// Verify Header (13 columns)
			header := lines[0]
			assert.Contains(t, header, "batch_id,code,radius_user,status,agent_id,price,activated_at,expire_time,extended_count,last_extended_at,is_deleted,created_at,updated_at")

			// Verify Row 1 (Active)
			// checking for non-empty dates
			assert.Contains(t, csvContent, "EXP001")
			
			// Verify Row 2 (Unused)
			// checking for empty dates (implicit check via not containing 0001-01-01)
			assert.NotContains(t, csvContent, "0001-01-01")
			assert.NotContains(t, csvContent, "1/1/2001") // User complaint format
		}
	})
}
