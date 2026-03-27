package adminapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

type TrafficStatsResponse struct {
	TotalBytes   uint64          `json:"total_bytes"`
	TotalPackets uint64          `json:"total_packets"`
	TotalFlows   uint64          `json:"total_flows"`
	DurationSec  int             `json:"duration_sec"`
	AvgMBPS      float64         `json:"avg_mbps"`
	AvgPPS       float64         `json:"avg_pps"`
	TopSources   []IPStats       `json:"top_sources"`
	TopDestins   []IPStats       `json:"top_destinations"`
	TopProtocols []ProtocolStats `json:"top_protocols"`
}

type IPStats struct {
	IP      string  `json:"ip"`
	Bytes   uint64  `json:"bytes"`
	Packets uint64  `json:"packets"`
	Flows   uint64  `json:"flows"`
	Percent float64 `json:"percent"`
}

type ProtocolStats struct {
	Protocol string  `json:"protocol"`
	Bytes    uint64  `json:"bytes"`
	Packets  uint64  `json:"packets"`
	Flows    uint64  `json:"flows"`
	Percent  float64 `json:"percent"`
}

type UserTrafficResponse struct {
	UserID        int64            `json:"user_id"`
	Username      string           `json:"username"`
	TotalBytes    uint64           `json:"total_bytes"`
	TotalSessions int              `json:"total_sessions"`
	AvgMBPS       float64          `json:"avg_mbps"`
	TopApps       []AppStats       `json:"top_applications"`
	SessionList   []SessionTraffic `json:"sessions"`
}

type AppStats struct {
	AppName string  `json:"app_name"`
	Bytes   uint64  `json:"bytes"`
	Percent float64 `json:"percent"`
}

type SessionTraffic struct {
	SessionID   string    `json:"session_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Bytes       uint64    `json:"bytes"`
	Packets     uint64    `json:"packets"`
	DurationSec int       `json:"duration_sec"`
}

type ApplicationStats struct {
	Name    string  `json:"name"`
	Bytes   uint64  `json:"bytes"`
	Packets uint64  `json:"packets"`
	Flows   uint64  `json:"flows"`
	Percent float64 `json:"percent"`
}

type LiveMetricsResponse struct {
	AuthRate          float64 `json:"auth_rate"`
	ActiveSessions    int64   `json:"active_sessions"`
	ThroughputInMBPS  float64 `json:"throughput_in_mbps"`
	ThroughputOutMBPS float64 `json:"throughput_out_mbps"`
	CPUUsage          float64 `json:"cpu_usage"`
	MemoryUsage       float64 `json:"memory_usage"`
	DiskUsage         float64 `json:"disk_usage"`
}

func GetTrafficStats(c echo.Context) error {
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())
	timeRange := c.QueryParam("time_range")
	if timeRange == "" {
		timeRange = "day"
	}

	db := GetDB(c)
	timeFilter := getTimeFilter(timeRange)

	type statsResult struct {
		TotalBytes   uint64
		TotalPackets uint64
		TotalFlows   int64
		FirstSeen    time.Time
		LastSeen     time.Time
	}

	var result statsResult
	db.Model(&domain.NetFlowRecord{}).
		Select("COALESCE(SUM(bytes), 0) as total_bytes, COALESCE(SUM(packets), 0) as total_packets, COUNT(*) as total_flows, MIN(first_switched) as first_seen, MAX(last_switched) as last_seen").
		Where("tenant_id = ? AND created_at >= ?", tenantID, timeFilter).
		Scan(&result)

	durationSec := 0
	if !result.FirstSeen.IsZero() && !result.LastSeen.IsZero() {
		durationSec = int(result.LastSeen.Sub(result.FirstSeen).Seconds())
	}

	mbps := 0.0
	pps := 0.0
	if durationSec > 0 {
		mbps = float64(result.TotalBytes) * 8 / float64(durationSec) / 1_000_000
		pps = float64(result.TotalPackets) / float64(durationSec)
	}

	return ok(c, TrafficStatsResponse{
		TotalBytes:   result.TotalBytes,
		TotalPackets: result.TotalPackets,
		TotalFlows:   uint64(result.TotalFlows),
		DurationSec:  durationSec,
		AvgMBPS:      mbps,
		AvgPPS:       pps,
	})
}

func GetUserTraffic(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid user ID", nil)
	}

	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())
	timeRange := c.QueryParam("time_range")
	if timeRange == "" {
		timeRange = "day"
	}

	db := GetDB(c)
	timeFilter := getTimeFilter(timeRange)

	type userTrafficResult struct {
		TotalBytes   uint64
		SessionCount int64
	}

	var result userTrafficResult
	db.Model(&domain.NetFlowRecord{}).
		Select("COALESCE(SUM(bytes), 0) as total_bytes, COUNT(DISTINCT session_id) as session_count").
		Where("tenant_id = ? AND user_id = ? AND created_at >= ?", tenantID, userID, timeFilter).
		Scan(&result)

	return ok(c, UserTrafficResponse{
		UserID:        userID,
		TotalBytes:    result.TotalBytes,
		TotalSessions: int(result.SessionCount),
		AvgMBPS:       0,
	})
}

func GetTopApplications(c echo.Context) error {
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	timeRange := c.QueryParam("time_range")
	if timeRange == "" {
		timeRange = "day"
	}

	db := GetDB(c)
	timeFilter := getTimeFilter(timeRange)

	type AppResult struct {
		AppName string
		Bytes   uint64
	}

	var results []AppResult
	db.Model(&domain.NetFlowRecord{}).
		Select("application_name as app_name, SUM(bytes) as bytes").
		Where("tenant_id = ? AND created_at >= ? AND application_name != ''", tenantID, timeFilter).
		Group("application_name").
		Order("bytes DESC").
		Limit(limit).
		Scan(&results)

	var totalBytes uint64
	for _, r := range results {
		totalBytes += r.Bytes
	}

	apps := make([]ApplicationStats, 0, len(results))
	for _, r := range results {
		percent := 0.0
		if totalBytes > 0 {
			percent = float64(r.Bytes) / float64(totalBytes) * 100
		}
		apps = append(apps, ApplicationStats{
			Name:    r.AppName,
			Bytes:   r.Bytes,
			Percent: percent,
		})
	}

	return ok(c, apps)
}

func GetLiveMetrics(c echo.Context) error {
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())
	db := GetDB(c)

	oneHourAgo := time.Now().Add(-1 * time.Hour)
	var authCount int64
	db.Model(&domain.RadiusOnline{}).Where("tenant_id = ? AND created_at >= ?", tenantID, oneHourAgo).Count(&authCount)

	var activeSessions int64
	db.Model(&domain.RadiusOnline{}).Where("tenant_id = ?", tenantID).Count(&activeSessions)

	type throughputResult struct {
		InputBytes  uint64
		OutputBytes uint64
	}

	var throughput throughputResult
	hourAgo := time.Now().Add(-1 * time.Hour)
	db.Model(&domain.RadiusAccounting{}).
		Select("COALESCE(SUM(input_octets), 0) as input_bytes, COALESCE(SUM(output_octets), 0) as output_bytes").
		Where("tenant_id = ? AND AcctStartTime >= ?", tenantID, hourAgo).
		Scan(&throughput)

	inputMBPS := float64(throughput.InputBytes) * 8 / 3600 / 1_000_000
	outputMBPS := float64(throughput.OutputBytes) * 8 / 3600 / 1_000_000

	return ok(c, LiveMetricsResponse{
		AuthRate:          float64(authCount),
		ActiveSessions:    activeSessions,
		ThroughputInMBPS:  inputMBPS,
		ThroughputOutMBPS: outputMBPS,
		CPUUsage:          0,
		MemoryUsage:       0,
		DiskUsage:         0,
	})
}

func getTimeFilter(timeRange string) time.Time {
	now := time.Now()
	switch timeRange {
	case "hour":
		return now.Add(-1 * time.Hour)
	case "day":
		return now.Add(-24 * time.Hour)
	case "week":
		return now.Add(-7 * 24 * time.Hour)
	case "month":
		return now.Add(-30 * 24 * time.Hour)
	default:
		return now.Add(-24 * time.Hour)
	}
}

func registerTrafficAnalysisRoutes() {
	webserver.ApiGET("/traffic/stats", GetTrafficStats)
	webserver.ApiGET("/traffic/user/:id", GetUserTraffic)
	webserver.ApiGET("/traffic/applications", GetTopApplications)
	webserver.ApiGET("/traffic/live", GetLiveMetrics)
}
