package analytics

import (
	"errors"
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
		return nil, err
	}

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
	var results []struct {
		Date  time.Time
		Total float64
	}

	// Group by date and sum traffic
	// Using generic SQL compatible with SQLite/PG for date grouping
	err := e.db.Table("tr_radius_accountings").
		Select("date(acct_start_time) as date, sum(acct_input_octets + acct_output_octets) as total").
		Where("acct_start_time > ?", startDate).
		Group("date").
		Order("date").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	dataPoints := make([]DataPoint, len(results))
	for i, r := range results {
		dataPoints[i] = DataPoint{
			Timestamp: r.Date,
			Value:     r.Total / 1024 / 1024, // MB
		}
	}
	return dataPoints, nil
}

func (e *PredictiveEngine) getUserHistory(days int) ([]DataPoint, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	var results []struct {
		Date  time.Time
		Count int64
	}

	err := e.db.Table("tr_radius_users").
		Select("date(created_at) as date, count(*) as count").
		Where("created_at > ?", startDate).
		Group("date").
		Order("date").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// We need cumulative count for user growth
	var cumulative int64
	
	// Get initial count before startDate
	e.db.Table("tr_radius_users").Where("created_at <= ?", startDate).Count(&cumulative)

	dataPoints := make([]DataPoint, len(results))
	for i, r := range results {
		cumulative += r.Count
		dataPoints[i] = DataPoint{
			Timestamp: r.Date,
			Value:     float64(cumulative),
		}
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

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
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
