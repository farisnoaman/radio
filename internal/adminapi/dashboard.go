package adminapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"go.uber.org/zap"
	"gorm.io/gorm"
)



var dashboardCache = cache.New(5*time.Minute, 10*time.Minute)


// DashboardStats represents the dashboard statistics structure
type DashboardStats struct {
	TotalUsers          int64                     `json:"total_users"`          // Total number of users
	OnlineUsers         int64                     `json:"online_users"`         // Currently online users
	TodayAuthCount      int64                     `json:"today_auth_count"`     // Authentication count for today
	TodayAcctCount      int64                     `json:"today_acct_count"`     // Accounting record count for today
	TotalProfiles       int64                     `json:"total_profiles"`       // Total number of profiles
	DisabledUsers       int64                     `json:"disabled_users"`       // Disabled users
	ExpiredUsers        int64                     `json:"expired_users"`        // Expired users
	TodayInputGB        float64                   `json:"today_input_gb"`       // Today's upstream traffic (GB)
	TodayOutputGB       float64                   `json:"today_output_gb"`      // Today's downstream traffic (GB)
	AuthTrend           []DashboardAuthTrendPoint `json:"auth_trend"`           // Daily authentication trend (last 7 days)
	Traffic24h          []DashboardTrafficPoint   `json:"traffic_24h"`          // Hourly traffic statistics (last 24 hours)
	ProfileDistribution []DashboardProfileSlice   `json:"profile_distribution"` // Online users grouped by profile
}

// DashboardAuthTrendPoint represents authentication count per day
type DashboardAuthTrendPoint struct {
	Date  string `json:"date"`  // Date label formatted as YYYY-MM-DD
	Count int64  `json:"count"` // Authentication count for the day
}

// DashboardTrafficPoint represents hourly upload/download traffic
type DashboardTrafficPoint struct {
	Hour       string  `json:"hour"`        // Hour label formatted as YYYY-MM-DD HH:00
	UploadGB   float64 `json:"upload_gb"`   // Upload traffic in GB within the hour
	DownloadGB float64 `json:"download_gb"` // Download traffic in GB within the hour
}

// DashboardProfileSlice represents online user distribution grouped by profile
type DashboardProfileSlice struct {
	ProfileID   int64  `json:"profile_id"`
	ProfileName string `json:"profile_name"`
	Value       int64  `json:"value"`
}

const (
	dateKeyFormat = "2006-01-02"
	hourKeyFormat = "2006-01-02 15:00"
	bytesInGB     = float64(1024 * 1024 * 1024)
)

// GetDashboardStats retrieves dashboard statistics
// @Summary get dashboard statistics
// @Tags Dashboard
// @Accept json
// @Produce json
// @Success 200 {object} DashboardStats
// @Router /api/v1/dashboard/stats [get]
func GetDashboardStats(c echo.Context) error {
	db := GetDB(c).WithContext(c.Request().Context())
	now := time.Now()
	todayStart := startOfDay(now)

	// Check cache
	cacheKey := "dashboard_stats"
	if cached, found := dashboardCache.Get(cacheKey); found {
		zap.L().Info("dashboard cache hit")
		return ok(c, cached)
	}
	zap.L().Info("dashboard cache miss")

	stats := &DashboardStats{}



	// 1. Total users
	if err := db.Model(&domain.RadiusUser{}).Count(&stats.TotalUsers).Error; err != nil {
		zap.L().Error("dashboard: failed to count total users", zap.Error(err))
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get user count", err)
	}

	// 2. Online users
	if err := db.Model(&domain.RadiusOnline{}).Count(&stats.OnlineUsers).Error; err != nil {
		zap.L().Error("dashboard: failed to count online users", zap.Error(err))
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get online user count", err)
	}

	// 3. Total profiles
	if err := db.Model(&domain.RadiusProfile{}).Count(&stats.TotalProfiles).Error; err != nil {
		zap.L().Error("dashboard: failed to count profiles", zap.Error(err))
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get profile count", err)
	}

	// 4. Disabled users
	if err := db.Model(&domain.RadiusUser{}).Where("status = ?", "disabled").Count(&stats.DisabledUsers).Error; err != nil {
		zap.L().Error("dashboard: failed to count disabled users", zap.Error(err))
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get disabled user count", err)
	}

	// 5. Expired users
	if err := db.Model(&domain.RadiusUser{}).Where("expire_time < ?", now).Count(&stats.ExpiredUsers).Error; err != nil {
		zap.L().Error("dashboard: failed to count expired users", zap.Error(err))
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get expired user count", err)
	}

	// 6. Today's authentication count (estimated from today's new online sessions)
	if err := db.Model(&domain.RadiusOnline{}).Where("acct_start_time >= ?", todayStart).Count(&stats.TodayAuthCount).Error; err != nil {
		zap.L().Error("dashboard: failed to count today's auth", zap.Error(err))
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get today's auth count", err)
	}

	// 7. Today's accounting record count
	if err := db.Model(&domain.RadiusAccounting{}).Where("acct_start_time >= ?", todayStart).Count(&stats.TodayAcctCount).Error; err != nil {
		zap.L().Error("dashboard: failed to count today's accounting", zap.Error(err))
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get today's accounting count", err)
	}

	// 8. Today's traffic statistics (bytes to GB)
	var flowStats struct {
		TotalInput  int64
		TotalOutput int64
	}
	if err := db.Model(&domain.RadiusAccounting{}).
		Select("COALESCE(SUM(acct_input_total), 0) as total_input, COALESCE(SUM(acct_output_total), 0) as total_output").
		Where("acct_start_time >= ?", todayStart).
		Scan(&flowStats).Error; err != nil {
		zap.L().Error("dashboard: failed to get traffic stats", zap.Error(err))
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get traffic statistics", err)
	}

	stats.TodayInputGB = float64(flowStats.TotalInput) / bytesInGB
	stats.TodayOutputGB = float64(flowStats.TotalOutput) / bytesInGB

	// Fetch additional stats
	authTrend, err := fetchAuthTrend(db, now)
	if err != nil {
		zap.L().Error("dashboard: failed to fetch auth trend", zap.Error(err))
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get auth trend", err)
	}
	stats.AuthTrend = authTrend

	trafficStats, err := fetchTrafficStats(db, now)
	if err != nil {
		zap.L().Error("dashboard: failed to fetch traffic stats", zap.Error(err))
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get traffic stats", err)
	}
	stats.Traffic24h = trafficStats

	profileDist, err := fetchProfileDistribution(db)
	if err != nil {
		zap.L().Error("dashboard: failed to fetch profile distribution", zap.Error(err))
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get profile distribution", err)
	}
	stats.ProfileDistribution = profileDist

	stats.ProfileDistribution = profileDist

	stats.ProfileDistribution = profileDist

	// Cache the result
	// Get TTL from config, default to 60s if 0 (though default config sets it)
	ttl := time.Duration(GetAppContext(c).Config().Web.CacheTTL) * time.Second
	if ttl == 0 {
		ttl = 60 * time.Second
	}
	dashboardCache.Set(cacheKey, stats, ttl)

	return ok(c, stats)



}

// registerDashboardRoutes registers the dashboard routes
func registerDashboardRoutes() {
	webserver.ApiGET("/dashboard/stats", GetDashboardStats)
}

func fetchAuthTrend(db *gorm.DB, now time.Time) ([]DashboardAuthTrendPoint, error) {
	const days = 7
	result := make([]DashboardAuthTrendPoint, days)
	seriesEnd := startOfDay(now).Add(24 * time.Hour)
	seriesStart := seriesEnd.AddDate(0, 0, -days)
	bucketExpr := dateBucketExpression(db, "acct_start_time", "day")
	var rows []struct {
		Bucket string
		Count  int64
	}
	if err := db.Model(&domain.RadiusAccounting{}).
		Select(fmt.Sprintf("%s AS bucket, COUNT(*) AS count", bucketExpr)).
		Where("acct_start_time >= ? AND acct_start_time < ?", seriesStart, seriesEnd).
		Group("bucket").
		Order("bucket").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("auth trend query failed: %w", err)
	}
	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Bucket] = row.Count
	}
	for i := 0; i < days; i++ {
		day := seriesStart.AddDate(0, 0, i)
		key := day.Format(dateKeyFormat)
		result[i] = DashboardAuthTrendPoint{
			Date:  key,
			Count: counts[key],
		}
	}
	return result, nil
}

func fetchTrafficStats(db *gorm.DB, now time.Time) ([]DashboardTrafficPoint, error) {
	const hours = 24
	result := make([]DashboardTrafficPoint, hours)
	hourEnd := startOfHour(now).Add(time.Hour)
	hourStart := hourEnd.Add(-hours * time.Hour)
	bucketExpr := dateBucketExpression(db, "acct_start_time", "hour")
	var rows []struct {
		Bucket   string
		Upload   float64
		Download float64
	}
	if err := db.Model(&domain.RadiusAccounting{}).
		Select(fmt.Sprintf("%s AS bucket, COALESCE(SUM(acct_input_total), 0) AS upload, COALESCE(SUM(acct_output_total), 0) AS download", bucketExpr)).
		Where("acct_start_time >= ? AND acct_start_time < ?", hourStart, hourEnd).
		Group("bucket").
		Order("bucket").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("traffic stats query failed: %w", err)
	}
	lookup := make(map[string]struct {
		Upload   float64
		Download float64
	})
	for _, row := range rows {
		lookup[row.Bucket] = struct {
			Upload   float64
			Download float64
		}{Upload: row.Upload / bytesInGB, Download: row.Download / bytesInGB}
	}
	for i := 0; i < hours; i++ {
		hour := hourStart.Add(time.Duration(i) * time.Hour)
		key := hour.Format(hourKeyFormat)
		if val, ok := lookup[key]; ok {
			result[i] = DashboardTrafficPoint{Hour: key, UploadGB: val.Upload, DownloadGB: val.Download}
			continue
		}
		result[i] = DashboardTrafficPoint{Hour: key, UploadGB: 0, DownloadGB: 0}
	}
	return result, nil
}

func fetchProfileDistribution(db *gorm.DB) ([]DashboardProfileSlice, error) {
	var rows []struct {
		ProfileID   int64
		ProfileName string
		Count       int64
	}
	onlineTable := domain.RadiusOnline{}.TableName()
	userTable := domain.RadiusUser{}.TableName()
	profileTable := domain.RadiusProfile{}.TableName()
	if err := db.Table(fmt.Sprintf("%s AS o", onlineTable)).
		Select("COALESCE(u.profile_id, 0) AS profile_id, COALESCE(p.name, '') AS profile_name, COUNT(*) AS count").
		Joins(fmt.Sprintf("LEFT JOIN %s AS u ON u.username = o.username", userTable)).
		Joins(fmt.Sprintf("LEFT JOIN %s AS p ON p.id = u.profile_id", profileTable)).
		Group("profile_id, profile_name").
		Order("count DESC").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("profile distribution query failed: %w", err)
	}
	result := make([]DashboardProfileSlice, 0, len(rows))
	for _, row := range rows {
		result = append(result, DashboardProfileSlice{
			ProfileID:   row.ProfileID,
			ProfileName: row.ProfileName,
			Value:       row.Count,
		})
	}
	return result, nil
}

func dateBucketExpression(db *gorm.DB, field, granularity string) string {
	switch db.Name() { //nolint:staticcheck
	case "postgres":
		switch granularity {
		case "day":
			return fmt.Sprintf("DATE(%s)", field)
		case "hour":
			return fmt.Sprintf("TO_CHAR(date_trunc('hour', %s), 'YYYY-MM-DD HH24:00')", field)
		}
	default:
		switch granularity {
		case "day":
			return fmt.Sprintf("strftime('%%Y-%%m-%%d', %s, 'localtime')", field)
		case "hour":
			return fmt.Sprintf("strftime('%%Y-%%m-%%d %%H:00', %s, 'localtime')", field)
		}
	}
	return field
}

func startOfDay(t time.Time) time.Time {
	loc := t.Location()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func startOfHour(t time.Time) time.Time {
	loc := t.Location()
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, loc)
}
