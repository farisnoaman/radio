package adminapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"gorm.io/gorm"
)

func registerEnvMonitoringRoutes() {
	webserver.ApiGET("/network/nas/metrics", GetAllMetricsHandler)
	webserver.ApiGET("/network/nas/alerts", GetAllAlertsHandler)
}

func GetAllMetricsHandler(c echo.Context) error {
	db := GetDB(c)
	tenantID := GetTenantID(c)

	var metrics []domain.EnvironmentMetric
	subQuery := db.Model(&domain.EnvironmentMetric{}).
		Select("nas_id, metric_type, MAX(collected_at) as max_collected").
		Group("nas_id, metric_type")

	if err := db.Raw(`
		SELECT em.* FROM environment_metrics em
		INNER JOIN (?) tm ON em.nas_id = tm.nas_id AND em.metric_type = tm.metric_type AND em.collected_at = tm.max_collected
		INNER JOIN net_nas nas ON em.nas_id = nas.id AND nas.tenant_id = ?
	`, subQuery, tenantID).Find(&metrics).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error(), nil)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"metrics": metrics,
		"count":   len(metrics),
	})
}

func GetAllAlertsHandler(c echo.Context) error {
	db := GetDB(c)
	tenantID := GetTenantID(c)

	var alerts []domain.EnvironmentAlert
	if err := db.Where("tenant_id = ?", tenantID).
		Order("fired_at DESC").Limit(100).Find(&alerts).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error(), nil)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

func EnvMetricHandlers(db *gorm.DB, webserver *echo.Echo) {
	group := webserver.Group("/api/v1/network/nas/:id")

	group.GET("/metrics", func(c echo.Context) error {
		nasID := uint(c.Get("nas_id").(float64))

		var metrics []domain.EnvironmentMetric
		if err := db.Where("nas_id = ?", nasID).
			Order("collected_at DESC").Limit(100).Find(&metrics).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error(), nil)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"metrics": metrics,
			"count":   len(metrics),
		})
	})

	group.GET("/metrics/history", func(c echo.Context) error {
		nasID := uint(c.Get("nas_id").(float64))
		metricType := c.QueryParam("type")
		days := 7
		if d := c.QueryParam("days"); d != "" {
			if parsed, err := strconv.Atoi(d); err == nil {
				days = parsed
			}
		}

		since := time.Now().AddDate(0, 0, -days)

		var metrics []domain.EnvironmentMetric
		query := db.Where("nas_id = ? AND collected_at > ?", nasID, since)
		if metricType != "" {
			query = query.Where("metric_type = ?", metricType)
		}
		if err := query.Order("collected_at DESC").Find(&metrics).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error(), nil)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"metrics": metrics,
			"count":   len(metrics),
		})
	})

	group.GET("/alerts", func(c echo.Context) error {
		nasID := uint(c.Get("nas_id").(float64))

		var alerts []domain.EnvironmentAlert
		if err := db.Where("nas_id = ?", nasID).
			Order("fired_at DESC").Limit(100).Find(&alerts).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error(), nil)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"alerts": alerts,
			"count":  len(alerts),
		})
	})

	group.GET("/alerts/config", func(c echo.Context) error {
		nasID := uint(c.Get("nas_id").(float64))

		var configs []domain.AlertConfig
		if err := db.Where("nas_id = ?", nasID).Find(&configs).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error(), nil)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"configs": configs,
			"count":   len(configs),
		})
	})

	group.PUT("/alerts/config", func(c echo.Context) error {
		nasID := uint(c.Get("nas_id").(float64))

		var req struct {
			MetricType     string  `json:"metric_type"`
			ThresholdType string  `json:"threshold_type"`
			ThresholdValue float64 `json:"threshold_value"`
			Severity       string  `json:"severity"`
			Enabled        bool    `json:"enabled"`
			NotifyEmail    bool    `json:"notify_email"`
			NotifyWebhook  bool    `json:"notify_webhook"`
			WebhookURL     string  `json:"webhook_url"`
		}
		if err := c.Bind(&req); err != nil {
			return fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		}

		var cfg domain.AlertConfig
		err := db.Where("nas_id = ? AND metric_type = ?", nasID, req.MetricType).First(&cfg).Error
		if err == gorm.ErrRecordNotFound {
			cfg = domain.AlertConfig{
				TenantID:       GetTenantID(c),
				NasID:          nasID,
				MetricType:     req.MetricType,
				ThresholdType:  req.ThresholdType,
				ThresholdValue: req.ThresholdValue,
				Severity:       req.Severity,
				Enabled:        req.Enabled,
				NotifyEmail:    req.NotifyEmail,
				NotifyWebhook:  req.NotifyWebhook,
				WebhookURL:     req.WebhookURL,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			if err := db.Create(&cfg).Error; err != nil {
				return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error(), nil)
			}
			return c.JSON(http.StatusCreated, map[string]interface{}{"config": cfg})
		}

		cfg.ThresholdType = req.ThresholdType
		cfg.ThresholdValue = req.ThresholdValue
		cfg.Severity = req.Severity
		cfg.Enabled = req.Enabled
		cfg.NotifyEmail = req.NotifyEmail
		cfg.NotifyWebhook = req.NotifyWebhook
		cfg.WebhookURL = req.WebhookURL
		cfg.UpdatedAt = time.Now()

		if err := db.Save(&cfg).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error(), nil)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"config": cfg})
	})

	group.POST("/alerts/:alertId/ack", func(c echo.Context) error {
		nasID := uint(c.Get("nas_id").(float64))
		alertID := uint(c.Get("alert_id").(float64))
		username := c.Get("username").(string)

		var alert domain.EnvironmentAlert
		if err := db.Where("id = ? AND nas_id = ?", alertID, nasID).First(&alert).Error; err != nil {
			return fail(c, http.StatusNotFound, "NOT_FOUND", "Alert not found", nil)
		}

		alert.Status = domain.AlertStatusAcknowledged
		alert.AcknowledgedBy = &username
		now := time.Now()
		alert.AcknowledgedAt = &now

		if err := db.Save(&alert).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error(), nil)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"alert": alert})
	})
}
