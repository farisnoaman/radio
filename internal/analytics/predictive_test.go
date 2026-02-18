package analytics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLinearRegression(t *testing.T) {
	engine := &PredictiveEngine{}
	
	now := time.Now()
	history := []DataPoint{
		{Timestamp: now.AddDate(0, 0, -4), Value: 10},
		{Timestamp: now.AddDate(0, 0, -3), Value: 20},
		{Timestamp: now.AddDate(0, 0, -2), Value: 30},
		{Timestamp: now.AddDate(0, 0, -1), Value: 40},
		{Timestamp: now, Value: 50},
	}
	
	// Expect next value to be around 60
	forecast := engine.calculateLinearRegression(history, 1)
	
	assert.Equal(t, 1, len(forecast))
	// Basic float check, might differ slightly due to time precision
	assert.InDelta(t, 60.0, forecast[0].Value, 1.0)
}

func TestLinearRegression_Empty(t *testing.T) {
	engine := &PredictiveEngine{}
	forecast := engine.calculateLinearRegression([]DataPoint{}, 5)
	assert.Empty(t, forecast)
}
