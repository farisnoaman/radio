package adminapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

func TestCampaignWorkflow(t *testing.T) {
	// Mock DB logic would go here. 
	// Since we are inside adminapi package, we can test internal logic if we extract it, 
	// or integration test if we have a test DB.
	// For now, simple struct validation.

	campaign := domain.VoucherCampaign{
		Name:      "Test Campaign",
		Count:     100,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	assert.Equal(t, "pending", campaign.Status)
	assert.Equal(t, 100, campaign.Count)
}
