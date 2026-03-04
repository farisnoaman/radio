package analytics

import (
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)


// PredictionRequest holds parameters for a forecast request
type PredictionRequest struct {
	Metric string `json:"metric"` // "traffic", "users"
	Days   int    `json:"days"`   // Number of days to look back
	Future int    `json:"future"` // Number of days to forecast
}

// PredictionResult holds the forecasted data
type PredictionResult struct {
	Metric    string      `json:"metric"`
	Forecast  []DataPoint `json:"forecast"`
	History   []DataPoint `json:"history"`
	Confidence float64     `json:"confidence"` // R-squared or similar confidence score
}

type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type PredictiveEngine struct {
	db *gorm.DB
}

func NewPredictiveEngine(db *gorm.DB) *PredictiveEngine {
	return &PredictiveEngine{db: db}
}

// Forecast generates a prediction based on historical data
func (e *PredictiveEngine) Forecast(req PredictionRequest) (*PredictionResult, error) {
	log.Printf("[analytics] Forecast request: metric=%s, days=%d, future=%d", req.Metric, req.Days, req.Future)
	
	var history []DataPoint
	var err error

	if req.Metric == "traffic" {
		history, err = e.getTrafficHistory(req.Days)
	} else if req.Metric == "users" {
		history, err = e.getUserHistory(req.Days)
	} else {
		return nil, errors.New("unsupported metric")
	}


	if err != nil {
		log.Printf("[analytics] Error getting history: %v", err)
		return nil, err
	}

	log.Printf("[analytics] Got %d history points", len(history))

	if len(history) < 2 {
		return &PredictionResult{
			Metric:   req.Metric,
			History:  history,
			Forecast: []DataPoint{},
		}, nil
	}

	// Simple Linear Regression
	// y = mx + c
	// x is time (unix timestamp), y is value
	
	forecast := e.calculateLinearRegression(history, req.Future)

	return &PredictionResult{
		Metric:   req.Metric,
		History:  history,
		Forecast: forecast,
		Confidence: 0.8, // Mock confidence for simple regression
	}, nil
}

func (e *PredictiveEngine) getTrafficHistory(days int) ([]DataPoint, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	
	// Use string for date because strftime returns "YYYY-MM-DD"
	var results []struct {
		Date  string
		Total float64
	}

	// Try with new column names (acct_input_total) first
	err := e.db.Table("radius_accounting").
		Select("strftime('%Y-%m-%d', acct_start_time, 'localtime') as date, SUM(acct_input_total + acct_output_total) as total").
		Where("acct_start_time > ?", startDate).
		Group("strftime('%Y-%m-%d', acct_start_time, 'localtime')").
		Order("strftime('%Y-%m-%d', acct_start_time, 'localtime')").
		Scan(&results).Error

	if err != nil {
		log.Printf("[analytics] Error with acct_input_total: %v, trying fallback...", err)
		// Fallback: try with old column name pattern
		err = e.db.Table("radius_accounting").
			Select("strftime('%Y-%m-%d', acct_start_time, 'localtime') as date, SUM(AcctInputTotal + AcctOutputTotal) as total").
			Where("acct_start_time > ?", startDate).
			Group("strftime('%Y-%m-%d', acct_start_time, 'localtime')").
			Order("strftime('%Y-%m-%d', acct_start_time, 'localtime')").
			Scan(&results).Error
	}

	if err != nil {
		log.Printf("[analytics] Both column attempts failed: %v", err)
		// Return empty data instead of error - chart will just show no data
		return []DataPoint{}, nil
	}

	dataPoints := make([]DataPoint, 0, len(results))
	for _, r := range results {
		t, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			log.Printf("[analytics] Warning: Failed to parse date %s: %v", r.Date, err)
			continue
		}
		dataPoints = append(dataPoints, DataPoint{
			Timestamp: t,
			Value:     r.Total / 1024 / 1024, // MB
		})
	}
	return dataPoints, nil
}

func (e *PredictiveEngine) getUserHistory(days int) ([]DataPoint, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	
	// Use string for date because strftime returns "YYYY-MM-DD"
	var results []struct {
		Date  string
		Count int64
	}

	err := e.db.Table("radius_user").
		Select("strftime('%Y-%m-%d', created_at, 'localtime') as date, count(*) as count").
		Where("created_at > ?", startDate).
		Group("strftime('%Y-%m-%d', created_at, 'localtime')").
		Order("strftime('%Y-%m-%d', created_at, 'localtime')").
		Scan(&results).Error

	if err != nil {
		log.Printf("[analytics] Error getting user history: %v", err)
		// Return empty data instead of error
		return []DataPoint{}, nil
	}

	// We need cumulative count for user growth
	var cumulative int64
	
	// Get initial count before startDate
	if err := e.db.Table("radius_user").Where("created_at <= ?", startDate).Count(&cumulative).Error; err != nil {
		log.Printf("[analytics] Warning: Failed to get initial user count: %v", err)
		cumulative = 0
	}

	dataPoints := make([]DataPoint, 0, len(results))
	for _, r := range results {
		cumulative += r.Count
		t, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			log.Printf("[analytics] Warning: Failed to parse date %s: %v", r.Date, err)
			continue
		}
		dataPoints = append(dataPoints, DataPoint{
			Timestamp: t,
			Value:     float64(cumulative),
		})
	}
	return dataPoints, nil
}

func (e *PredictiveEngine) calculateLinearRegression(history []DataPoint, futureDays int) []DataPoint {
	if len(history) == 0 {
		return []DataPoint{}
	}

	n := float64(len(history))
	var sumX, sumY, sumXY, sumX2 float64
	
	startTime := history[0].Timestamp.Unix()

	for _, p := range history {
		x := float64(p.Timestamp.Unix() - startTime) // Normalize time to start from 0 for stability
		y := p.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Guard against division by zero (when all X values are identical)
	denominator := n*sumX2 - sumX*sumX
	var slope float64
	if denominator == 0 {
		log.Printf("[analytics] Warning: Division by zero in linear regression - all data points have same timestamp")
		slope = 0
	} else {
		slope = (n*sumXY - sumX*sumY) / denominator
	}
	intercept := (sumY - slope*sumX) / n

	var forecast []DataPoint
	lastTime := history[len(history)-1].Timestamp
	
	for i := 1; i <= futureDays; i++ {
		futureTime := lastTime.AddDate(0, 0, i)
		x := float64(futureTime.Unix() - startTime)
		y := slope*x + intercept
		if y < 0 {
			y = 0
		}
		forecast = append(forecast, DataPoint{
			Timestamp: futureTime,
			Value:     y,
		})
	}

	return forecast
}
