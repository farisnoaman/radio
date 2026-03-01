package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

func TestInvoiceEndpoints(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDB(t)

	// Create a test user for invoices
	testUser := domain.RadiusUser{
		Username:           "postpaiduser1",
		BillingType:        domain.BillingTypePostpaid,
		SubscriptionStatus: domain.SubscriptionSuspended,
	}
	db.Create(&testUser)

	// Create test invoices
	invoice1 := domain.Invoice{
		ID:       1001,
		Username: "postpaiduser1",
		Amount:   50.0,
		Status:   domain.InvoiceUnpaid,
		DueDate:  time.Now().AddDate(0, 0, -1),
	}
	invoice2 := domain.Invoice{
		ID:       1002,
		Username: "postpaiduser2",
		Amount:   100.0,
		Status:   domain.InvoicePaid,
	}
	db.Create(&invoice1)
	db.Create(&invoice2)

	t.Run("ListInvoices", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/invoices", nil), rec)
		c.Set("db", db)
		
		err := ListInvoices(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var res map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &res)
		
		data := res["data"].([]interface{})
		assert.Len(t, data, 2)
	})

	t.Run("ListInvoices Filter Status", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/invoices?status=unpaid", nil), rec)
		c.Set("db", db)
		
		err := ListInvoices(c)
		require.NoError(t, err)
		
		var res map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &res)
		
		data := res["data"].([]interface{})
		assert.Len(t, data, 1)
		first := data[0].(map[string]interface{})
		assert.Equal(t, "postpaiduser1", first["username"])
	})

	t.Run("GetInvoice", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
		c.SetPath("/api/v1/invoices/:id")
		c.SetParamNames("id")
		c.SetParamValues("1002")
		c.Set("db", db)

		err := GetInvoice(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GetInvoice NotFound", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
		c.SetPath("/api/v1/invoices/:id")
		c.SetParamNames("id")
		c.SetParamValues("9999")
		c.Set("db", db)

		err := GetInvoice(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("PayInvoice", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/", nil), rec)
		c.SetPath("/api/v1/invoices/:id/pay")
		c.SetParamNames("id")
		c.SetParamValues("1001")
		c.Set("db", db)

		err := PayInvoice(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var updatedInvoice domain.Invoice
		db.First(&updatedInvoice, 1001)
		assert.Equal(t, domain.InvoicePaid, updatedInvoice.Status)

		var updatedUser domain.RadiusUser
		db.First(&updatedUser, "username = ?", "postpaiduser1")
		assert.Equal(t, domain.SubscriptionActive, updatedUser.SubscriptionStatus)
	})

	t.Run("PayInvoice NotFound", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/", nil), rec)
		c.SetPath("/api/v1/invoices/:id/pay")
		c.SetParamNames("id")
		c.SetParamValues("9999")
		c.Set("db", db)

		err := PayInvoice(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
