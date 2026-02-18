package adminapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
)

func registerCampaignRoutes() {
	webserver.ApiGET("/campaigns", ListCampaigns)
	webserver.ApiPOST("/campaigns", CreateCampaign)
	webserver.ApiPOST("/campaigns/:id/start", StartCampaign)
}

// ListCampaigns retrieves the campaign list
func ListCampaigns(c echo.Context) error {
	db := GetDB(c)
	var campaigns []domain.VoucherCampaign
	if err := db.Order("id desc").Find(&campaigns).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to list campaigns", err.Error())
	}
	return ok(c, campaigns)
}

// CreateCampaign defines a new voucher generation campaign
func CreateCampaign(c echo.Context) error {
	var req struct {
		Name   string `json:"name"`
		Prefix string `json:"prefix"`
		Length int    `json:"length"`
		Count  int    `json:"count"`
		PlanId int64  `json:"plan_id"`
		Value  int64  `json:"value"`
	}
	
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
	}

	campaign := domain.VoucherCampaign{
		Name:      req.Name,
		Prefix:    req.Prefix,
		Length:    req.Length,
		Count:     req.Count,
		PlanId:    req.PlanId,
		Value:     req.Value,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	db := GetDB(c)
	if err := db.Create(&campaign).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to create campaign", err.Error())
	}

	return ok(c, campaign)
}

// StartCampaign triggers the async generation
func StartCampaign(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db := GetDB(c)
	
	var campaign domain.VoucherCampaign
	if err := db.First(&campaign, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Campaign not found", nil)
	}

	if campaign.Status != "pending" {
		return fail(c, http.StatusBadRequest, "INVALID_STATUS", "Campaign is not pending", nil)
	}

	// Update status to generating
	campaign.Status = "generating"
	db.Save(&campaign)

	// Async generation
	go func() {
		// Mock generation for now (or call specific generation logic)
		// Real implementation would batch insert vouchers similar to CreateVoucherBatch
		
		tx := db.Begin()
		
		// Create a batch record for this campaign
		batch := domain.VoucherBatch{
			Name:      campaign.Name,
			ProductID: campaign.PlanId, // Assuming PlanId maps to ProductID roughly for this context
			Count:     campaign.Count,
			Prefix:    campaign.Prefix,
			CreatedAt: time.Now(),
			Remark:    fmt.Sprintf("Campaign %d", campaign.ID),
		}
		
		if err := tx.Create(&batch).Error; err != nil {
			tx.Rollback()
			zap.S().Errorf("Campaign %d failed to create batch: %v", campaign.ID, err)
			return
		}

		// Generate codes
		vouchers := make([]domain.Voucher, 0, campaign.Count)
		for i := 0; i < campaign.Count; i++ {
			code := campaign.Prefix + common.GenerateVoucherCode(campaign.Length, "mixed")

			vouchers = append(vouchers, domain.Voucher{
				BatchID:   batch.ID,
				Code:      code,
				Status:    "unused",
				Price:     float64(campaign.Value),
				CreatedAt: time.Now(),
			})
		}
		
		if err := tx.CreateInBatches(vouchers, 1000).Error; err != nil {
			tx.Rollback()
			campaign.Status = "failed"
			db.Save(&campaign)
			zap.S().Errorf("Campaign %d failed generation: %v", campaign.ID, err)
			return
		}
		
		tx.Commit()
		
		campaign.Status = "completed"
		db.Save(&campaign)
		zap.S().Infof("Campaign %d completed successfully", campaign.ID)
	}()

	return ok(c, campaign)
}
