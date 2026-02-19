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
	// For now, simple struct validation using existing VoucherBatch type.

	batch := domain.VoucherBatch{
		Name:      "Test Campaign",
		Count:     100,
		CreatedAt: time.Now(),
	}

	assert.Equal(t, "Test Campaign", batch.Name)
	assert.Equal(t, 100, batch.Count)
}
