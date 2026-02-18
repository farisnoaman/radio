package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

func TestGetDashboardStats(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	now := time.Now()

	profiles := []*domain.RadiusProfile{
		{Name: "default", Status: "enabled"},
		{Name: "premium", Status: "enabled"},
	}
	for _, profile := range profiles {
		err := db.Create(profile).Error
		require.NoError(t, err)
	}

	users := []*domain.RadiusUser{
		{
			Username:   "alice",
			ProfileId:  profiles[0].ID,
			Status:     "enabled",
			ExpireTime: now.Add(24 * time.Hour),
		},
		{
			Username:   "bob",
			ProfileId:  profiles[1].ID,
			Status:     "disabled",
			ExpireTime: now.Add(48 * time.Hour),
		},
		{
			Username:   "carol",
			ProfileId:  profiles[0].ID,
			Status:     "enabled",
			ExpireTime: now.Add(-24 * time.Hour),
		},
	}
	for _, user := range users {
		err := db.Create(user).Error
		require.NoError(t, err)
	}

	onlineSessions := []*domain.RadiusOnline{
		{
			Username:      "alice",
			NasPortType:   5,
			ServiceType:   2,
			AcctSessionId: "session-alice",
			AcctStartTime: now.Add(-30 * time.Minute),
		},
		{
			Username:      "carol",
			NasPortType:   5,
			ServiceType:   2,
			AcctSessionId: "session-carol",
			AcctStartTime: now.Add(-25 * time.Minute),
		},
		{
			Username:      "bob",
			NasPortType:   19,
			ServiceType:   2,
			AcctSessionId: "session-bob",
			AcctStartTime: now.Add(-15 * time.Minute),
		},
		{
			Username:      "ghost",
			NasPortType:   15,
			ServiceType:   2,
			AcctSessionId: "session-ghost",
			AcctStartTime: now.Add(-10 * time.Minute),
		},
	}
	for _, session := range onlineSessions {
		err := db.Create(session).Error
		require.NoError(t, err)
	}

	accountingRecords := []*domain.RadiusAccounting{
		{
			Username:        "alice",
			AcctSessionId:   "acct-1",
			AcctStartTime:   now.Add(-1 * time.Minute), // 1 minute ago - within today and 24h window
			AcctInputTotal:  int64(1 * 1024 * 1024 * 1024),
			AcctOutputTotal: int64(2 * 1024 * 1024 * 1024),
		},
		{
			Username:        "alice",
			AcctSessionId:   "acct-2",
			AcctStartTime:   now.Add(-26 * time.Hour), // 26 hours ago - outside 24h window
			AcctInputTotal:  int64(500 * 1024 * 1024),
			AcctOutputTotal: int64(256 * 1024 * 1024),
		},
		{
			Username:        "carol",
			AcctSessionId:   "acct-3",
			AcctStartTime:   now.Add(-5 * 24 * time.Hour), // 5 days ago - outside both windows
			AcctInputTotal:  int64(200 * 1024 * 1024),
			AcctOutputTotal: int64(300 * 1024 * 1024),
		},
	}
	for _, record := range accountingRecords {
		err := db.Create(record).Error
		require.NoError(t, err)
	}

	e := setupTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/stats", nil)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)

	err := GetDashboardStats(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	dataBytes, err := json.Marshal(response.Data)
	require.NoError(t, err)

	var stats DashboardStats
	err = json.Unmarshal(dataBytes, &stats)
	require.NoError(t, err)

	assert.Equal(t, int64(3), stats.TotalUsers)
	assert.Equal(t, int64(4), stats.OnlineUsers)
	assert.Equal(t, int64(2), stats.TotalProfiles)
	assert.Equal(t, int64(1), stats.DisabledUsers)
	assert.Equal(t, int64(1), stats.ExpiredUsers)
	require.Len(t, stats.AuthTrend, 7)

	var trendTotal int64
	for _, point := range stats.AuthTrend {
		trendTotal += point.Count
	}
	assert.Equal(t, int64(3), trendTotal)

	require.Len(t, stats.Traffic24h, 24)
	// Calculate total traffic from all 24h points
	// Note: Due to timezone differences between Go's time.Now() and SQLite's localtime,
	// we check total traffic sum instead of specific hour to ensure CI compatibility
	var totalDownloadGB float64
	for _, point := range stats.Traffic24h {
		totalDownloadGB += point.DownloadGB
	}
	// The first accounting record (acct-1) has 2GB download and was created 1 minute ago
	assert.InDelta(t, 2.0, totalDownloadGB, 0.01)
	assert.GreaterOrEqual(t, stats.TodayInputGB, 1.0)
	assert.GreaterOrEqual(t, stats.TodayOutputGB, 2.0)

	require.GreaterOrEqual(t, len(stats.ProfileDistribution), 2)
	profileMap := make(map[int64]DashboardProfileSlice)
	var unassignedCount int64
	for _, item := range stats.ProfileDistribution {
		profileMap[item.ProfileID] = item
		if item.ProfileID == 0 {
			unassignedCount = item.Value
		}
	}
	defaultProfile := profileMap[profiles[0].ID]
	require.Equal(t, profiles[0].Name, defaultProfile.ProfileName)
	assert.Equal(t, int64(2), defaultProfile.Value)
	premiumProfile := profileMap[profiles[1].ID]
	require.Equal(t, profiles[1].Name, premiumProfile.ProfileName)
	assert.Equal(t, int64(1), premiumProfile.Value)
	assert.Equal(t, int64(1), premiumProfile.Value)
	assert.Equal(t, int64(1), unassignedCount)

	// Test Caching
	// Add a new user
	newUser := &domain.RadiusUser{
		Username:   "dave",
		ProfileId:  profiles[0].ID,
		Status:     "enabled",
		ExpireTime: now.Add(24 * time.Hour),
	}
	err = db.Create(newUser).Error
	require.NoError(t, err)

	// User count should still be 3 due to cache
	rec2 := httptest.NewRecorder()
	// Create new context for second request
	c2 := CreateTestContext(e, db, req, rec2, appCtx)
	err = GetDashboardStats(c2)
	require.NoError(t, err)

	var response2 Response
	err = json.Unmarshal(rec2.Body.Bytes(), &response2)
	require.NoError(t, err)
	dataBytes2, _ := json.Marshal(response2.Data)
	var stats2 DashboardStats
	json.Unmarshal(dataBytes2, &stats2)

	assert.Equal(t, int64(3), stats2.TotalUsers, "TotalUsers should be cached (3)")

	// Flush cache
	dashboardCache.Flush()

	// User count should now be 4
	rec3 := httptest.NewRecorder()
	c3 := CreateTestContext(e, db, req, rec3, appCtx)
	err = GetDashboardStats(c3)
	require.NoError(t, err)

	var response3 Response
	err = json.Unmarshal(rec3.Body.Bytes(), &response3)
	require.NoError(t, err)
	dataBytes3, _ := json.Marshal(response3.Data)
	var stats3 DashboardStats
	json.Unmarshal(dataBytes3, &stats3)

	assert.Equal(t, int64(4), stats3.TotalUsers, "TotalUsers should update after cache flush (4)")

	// Test Short TTL
	appCtx.Config().Web.CacheTTL = 1 // Set TTL to 1 second
	dashboardCache.Flush()

	// Request 4: Cache the result with 1s TTL
	c4 := CreateTestContext(e, db, req, httptest.NewRecorder(), appCtx)
	err = GetDashboardStats(c4)
	require.NoError(t, err)

	// Add another user (total 5)
	user5 := &domain.RadiusUser{
		Username:   "eve",
		ProfileId:  profiles[0].ID,
		Status:     "enabled",
		ExpireTime: now.Add(24 * time.Hour),
	}
	err = db.Create(user5).Error
	require.NoError(t, err)

	// Immediate request: Should still be cached (4 users)
	rec5 := httptest.NewRecorder()
	c5 := CreateTestContext(e, db, req, rec5, appCtx)
	err = GetDashboardStats(c5)
	require.NoError(t, err)
	var stats5 DashboardStats
	json.Unmarshal(rec5.Body.Bytes(), &struct{ Data *DashboardStats }{Data: &stats5})
	assert.Equal(t, int64(4), stats5.TotalUsers, "Should return cached data immediately")

	// Wait 1.1s for expiration
	time.Sleep(1100 * time.Millisecond)

	// Request 6: Should fetch fresh data (5 users)
	rec6 := httptest.NewRecorder()
	c6 := CreateTestContext(e, db, req, rec6, appCtx)
	err = GetDashboardStats(c6)
	require.NoError(t, err)

	var response6 Response
	err = json.Unmarshal(rec6.Body.Bytes(), &response6)
	require.NoError(t, err)
	dataBytes6, _ := json.Marshal(response6.Data)
	var stats6 DashboardStats
	json.Unmarshal(dataBytes6, &stats6)

	assert.Equal(t, int64(5), stats6.TotalUsers, "Should return fresh data after TTL expiration")
}


