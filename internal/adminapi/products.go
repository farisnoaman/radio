package adminapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/repository"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// ListProducts retrieves the product list
// @Summary get the product list
// @Tags Product
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort direction"
// @Success 200 {object} ListResponse
// @Router /api/v1/products [get]
func ListProducts(c echo.Context) error {
	db := GetDB(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	sortField := c.QueryParam("sort")
	order := c.QueryParam("order")
	if sortField == "" {
		sortField = "id"
	}
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	var total int64
	var products []domain.Product

	query := db.Model(&domain.Product{}).Scopes(repository.TenantScope)

	// Filter by name
	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		if strings.EqualFold(db.Name(), "postgres") {
			query = query.Where("name ILIKE ?", "%"+name+"%")
		} else {
			query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")
		}
	}

	// Filter by status
	if status := strings.TrimSpace(c.QueryParam("status")); status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&products)

	return paged(c, products, total, page, perPage)
}

// GetProduct retrieves a single product
// @Summary get product detail
// @Tags Product
// @Param id path int true "Product ID"
// @Success 200 {object} domain.Product
// @Router /api/v1/products/{id} [get]
func GetProduct(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID", nil)
	}

	var product domain.Product
	if err := GetDB(c).Scopes(repository.TenantScope).First(&product, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Product not found", nil)
	}

	return ok(c, product)
}

// ProductRequest represents the JSON body for creating/updating a product
type ProductRequest struct {
	Name            string  `json:"name" validate:"required,min=1,max=100"`
	RadiusProfileID string  `json:"radius_profile_id" validate:"required"`
	Price           float64 `json:"price" validate:"gte=0"`
	CostPrice       float64 `json:"cost_price" validate:"gte=0"`
	UpRate          int     `json:"up_rate" validate:"gte=0"`
	DownRate        int     `json:"down_rate" validate:"gte=0"`
	DataQuota       int64   `json:"data_quota" validate:"gte=0"`
	ValiditySeconds int64   `json:"validity_seconds" validate:"gte=0"`
	IdleTimeout     int     `json:"idle_timeout" validate:"gte=0"`
	SessionTimeout  int     `json:"session_timeout" validate:"gte=0"`
	Status          string  `json:"status"`
	Color           string  `json:"color"`
	Remark          string  `json:"remark" validate:"omitempty,max=500"`
}

// CreateProduct creates a product
// @Summary create a product
// @Tags Product
// @Param product body ProductRequest true "Product information"
// @Success 201 {object} domain.Product
// @Router /api/v1/products [post]
func CreateProduct(c echo.Context) error {
	var req ProductRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get tenant ID from context
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

	profileID, _ := strconv.ParseInt(req.RadiusProfileID, 10, 64)

	// Validate Profile Exists and belongs to same tenant
	var profileCount int64
	GetDB(c).Model(&domain.RadiusProfile{}).Where("tenant_id = ? AND id = ?", tenantID, profileID).Count(&profileCount)
	if profileCount == 0 {
		return fail(c, http.StatusBadRequest, "INVALID_PROFILE", "Radius Profile not found or not accessible", nil)
	}

	// Check whether a product with the same name already exists (business logic validation)
	var count int64
	GetDB(c).Model(&domain.Product{}).Where("tenant_id = ? AND name = ?", tenantID, req.Name).Count(&count)
	if count > 0 {
		return fail(c, http.StatusConflict, "NAME_EXISTS", "Product name already exists", nil)
	}

	product := domain.Product{
		Name:            req.Name,
		TenantID:        tenantID,
		RadiusProfileID: profileID,
		Price:           req.Price,
		CostPrice:       req.CostPrice,
		UpRate:          req.UpRate,
		DownRate:        req.DownRate,
		DataQuota:       req.DataQuota,
		ValiditySeconds: req.ValiditySeconds,
		IdleTimeout:     req.IdleTimeout,
		SessionTimeout:  req.SessionTimeout,
		Status:          req.Status,
		Color:           req.Color,
		Remark:          req.Remark,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if product.Status == "" {
		product.Status = "enabled"
	}

	if product.Color == "" {
		product.Color = "#1976d2" // Default blue color
	}

	if err := GetDB(c).Create(&product).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create product", err.Error())
	}

	return ok(c, product)
}

// UpdateProduct updates a product
// @Summary update a product
// @Tags Product
// @Param id path int true "Product ID"
// @Param product body ProductRequest true "Product information"
// @Success 200 {object} domain.Product
// @Router /api/v1/products/{id} [put]
func UpdateProduct(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID", nil)
	}

	// Get tenant ID from context
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

	var product domain.Product
	if err := GetDB(c).Scopes(repository.TenantScope).First(&product, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Product not found", nil)
	}

	var req ProductRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	profileID, _ := strconv.ParseInt(req.RadiusProfileID, 10, 64)

	// Validate Profile Exists and belongs to same tenant if changed
	if profileID != product.RadiusProfileID {
		var profileCount int64
		GetDB(c).Model(&domain.RadiusProfile{}).Where("tenant_id = ? AND id = ?", tenantID, profileID).Count(&profileCount)
		if profileCount == 0 {
			return fail(c, http.StatusBadRequest, "INVALID_PROFILE", "Radius Profile not found or not accessible", nil)
		}
	}

	// Check whether another product with the same name already exists (business logic validation)
	var count int64
	GetDB(c).Model(&domain.Product{}).Where("tenant_id = ? AND name = ? AND id != ?", tenantID, req.Name, id).Count(&count)
	if count > 0 {
		return fail(c, http.StatusConflict, "NAME_EXISTS", "Product name already exists", nil)
	}

	product.Name = req.Name
	product.RadiusProfileID = profileID
	product.Price = req.Price
	product.CostPrice = req.CostPrice
	product.UpRate = req.UpRate
	product.DownRate = req.DownRate
	product.DataQuota = req.DataQuota
	product.ValiditySeconds = req.ValiditySeconds
	product.IdleTimeout = req.IdleTimeout
	product.SessionTimeout = req.SessionTimeout
	product.Status = req.Status
	product.Color = req.Color
	product.Remark = req.Remark
	product.UpdatedAt = time.Now()

	if err := GetDB(c).Save(&product).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update product", err.Error())
	}

	return ok(c, product)
}

// DeleteProduct delete a product
// @Summary delete a product
// @Tags Product
// @Param id path int true "Product ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/products/{id} [delete]
func DeleteProduct(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID", nil)
	}

	// Get tenant ID from context
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

	var product domain.Product
	if err := GetDB(c).Scopes(repository.TenantScope).First(&product, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Product not found", nil)
	}

	// Check usage in Vouchers
	var voucherCount int64
	GetDB(c).Model(&domain.VoucherBatch{}).Where("product_id = ? AND tenant_id = ?", id, tenantID).Count(&voucherCount)
	if voucherCount > 0 {
		return fail(c, http.StatusConflict, "IN_USE", "Product is used by voucher batches", nil)
	}

	if err := GetDB(c).Delete(&domain.Product{}, id).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete product", err.Error())
	}

	return ok(c, map[string]interface{}{
		"message": "Deletion successful",
	})
}

func registerProductRoutes() {
	webserver.ApiGET("/products", ListProducts)
	webserver.ApiGET("/products/:id", GetProduct)
	webserver.ApiPOST("/products", CreateProduct)
	webserver.ApiPUT("/products/:id", UpdateProduct)
	webserver.ApiDELETE("/products/:id", DeleteProduct)
}
