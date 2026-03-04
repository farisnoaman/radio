package adminapi

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/analytics"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

func registerAnalyticsRoutes() {
	webserver.ApiGET("/analytics/forecast", GetForecast)
}

func GetForecast(c echo.Context) error {
	metric := c.QueryParam("metric")
	if metric == "" {
		metric = "traffic"
	}
	
	days, _ := strconv.Atoi(c.QueryParam("days"))
	if days == 0 {
		days = 30
	}

	future, _ := strconv.Atoi(c.QueryParam("future"))
	if future == 0 {
		future = 7
	}

	appCtx := GetAppContext(c)
	engine := analytics.NewPredictiveEngine(appCtx.DB())
	
	result, err := engine.Forecast(analytics.PredictionRequest{
		Metric: metric,
		Days:   days,
		Future: future,
	})
	
	if err != nil {
		// Log the actual error for debugging
		return fail(c, http.StatusInternalServerError, "ANALYSIS_ERROR", "Failed to generate forecast: "+err.Error(), nil)
	}

	return ok(c, result)
}
