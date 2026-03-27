package adminapi

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/service"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

type ReportingHandler struct {
	reportingService    *service.ReportingService
	fraudService       *service.FraudDetectionService
	notificationSvc    *service.ProviderNotificationService
}

func NewReportingHandler(db *gorm.DB) *ReportingHandler {
	return &ReportingHandler{
		reportingService: service.NewReportingService(db),
		fraudService:     service.NewFraudDetectionService(db),
		notificationSvc:  service.NewProviderNotificationService(db, nil),
	}
}

func registerReportingRoutes(appCtx app.AppContext) {
	handler := NewReportingHandler(appCtx.DB())

	webserver.ApiGET("/reporting/summary", handler.GetSummary)
	webserver.ApiGET("/reporting/network-status", handler.GetNetworkStatus)
	webserver.ApiGET("/reporting/agents", handler.GetAgentMetrics)
	webserver.ApiGET("/reporting/issues", handler.GetIssues)
	webserver.ApiGET("/reporting/fraud", handler.GetFraudLogs)
	webserver.ApiGET("/reporting/export", handler.ExportCSV)
	webserver.ApiGET("/reporting/notifications/preferences", handler.GetNotificationPreferences)
	webserver.ApiPUT("/reporting/notifications/preferences", handler.UpdateNotificationPreferences)
}

func (h *ReportingHandler) GetSummary(c echo.Context) error {
	providerID := GetOperatorTenantID(c)

	period := c.QueryParam("period")
	if period == "" {
		period = "daily"
	}

	startDate := parseDate(c.QueryParam("start_date"), time.Now().AddDate(0, -1, 0))
	endDate := parseDate(c.QueryParam("end_date"), time.Now())

	summary, err := h.reportingService.GetSummary(providerID, period, startDate, endDate)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "REPORTING_ERROR", err.Error(), nil)
	}

	return ok(c, summary)
}

func (h *ReportingHandler) GetNetworkStatus(c echo.Context) error {
	providerID := GetOperatorTenantID(c)

	metrics, err := h.reportingService.GetNetworkStatus(providerID)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "NETWORK_ERROR", err.Error(), nil)
	}

	return ok(c, metrics)
}

func (h *ReportingHandler) GetAgentMetrics(c echo.Context) error {
	providerID := GetOperatorTenantID(c)

	metrics, err := h.reportingService.GetAgentMetrics(providerID)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "AGENTS_ERROR", err.Error(), nil)
	}

	return ok(c, metrics)
}

func (h *ReportingHandler) GetIssues(c echo.Context) error {
	providerID := GetOperatorTenantID(c)
	limit := parseInt(c.QueryParam("limit"), 50)

	issues, err := h.reportingService.GetOpenIssues(providerID, limit)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "ISSUES_ERROR", err.Error(), nil)
	}

	return ok(c, issues)
}

func (h *ReportingHandler) GetFraudLogs(c echo.Context) error {
	providerID := GetOperatorTenantID(c)
	limit := parseInt(c.QueryParam("limit"), 50)

	logs, err := h.fraudService.GetFraudLogs(providerID, limit)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "FRAUD_ERROR", err.Error(), nil)
	}

	return ok(c, logs)
}

func (h *ReportingHandler) GetNotificationPreferences(c echo.Context) error {
	providerID := GetOperatorTenantID(c)

	pref, err := h.notificationSvc.GetPreferences(providerID)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "PREFERENCES_ERROR", err.Error(), nil)
	}

	return ok(c, pref)
}

func (h *ReportingHandler) UpdateNotificationPreferences(c echo.Context) error {
	providerID := GetOperatorTenantID(c)

	var req service.NotificationPreferenceUpdate
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
	}

	if err := h.notificationSvc.UpdatePreferences(providerID, &req); err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_ERROR", err.Error(), nil)
	}

	return ok(c, map[string]string{"message": "Preferences updated"})
}

func (h *ReportingHandler) ExportCSV(c echo.Context) error {
	providerID := GetOperatorTenantID(c)

	startDate := parseDate(c.QueryParam("start_date"), time.Now().AddDate(0, -1, 0))
	endDate := parseDate(c.QueryParam("end_date"), time.Now())

	summary, err := h.reportingService.GetSummary(providerID, "daily", startDate, endDate)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "EXPORT_ERROR", err.Error(), nil)
	}

	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=report_%s.csv", time.Now().Format("20060102")))

	writer := csv.NewWriter(c.Response().Writer)
	defer writer.Flush()

	writer.Write([]string{"Metric", "Value"})
	writer.Write([]string{"Period", summary.Period})
	writer.Write([]string{"Start Date", summary.StartDate})
	writer.Write([]string{"End Date", summary.EndDate})
	writer.Write([]string{""})
	writer.Write([]string{"Users"})
	writer.Write([]string{"Total Users", strconv.Itoa(summary.Users.TotalUsers)})
	writer.Write([]string{"Active Users", strconv.Itoa(summary.Users.ActiveUsers)})
	writer.Write([]string{"New Monthly Users", strconv.Itoa(summary.Users.NewMonthlyUsers)})
	writer.Write([]string{"New Voucher Users", strconv.Itoa(summary.Users.NewVoucherUsers)})
	writer.Write([]string{""})
	writer.Write([]string{"Sessions"})
	writer.Write([]string{"Active Sessions", strconv.Itoa(summary.Sessions.ActiveSessions)})
	writer.Write([]string{""})
	writer.Write([]string{"Data"})
	writer.Write([]string{"Monthly Data (GB)", fmt.Sprintf("%.2f", summary.Data.MonthlyDataUsedGB)})
	writer.Write([]string{"Voucher Data (GB)", fmt.Sprintf("%.2f", summary.Data.VoucherDataUsedGB)})

	return nil
}

func parseDate(s string, defaultVal time.Time) time.Time {
	if s == "" {
		return defaultVal
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return defaultVal
	}
	return t
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}
