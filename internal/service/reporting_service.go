package service

import (
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ReportingService struct {
	db *gorm.DB
}

func NewReportingService(db *gorm.DB) *ReportingService {
	return &ReportingService{db: db}
}

type ReportingSummary struct {
	Period    string         `json:"period"`
	StartDate string         `json:"start_date"`
	EndDate   string         `json:"end_date"`
	Users     UserMetrics    `json:"users"`
	Sessions  SessionMetrics `json:"sessions"`
	Data      DataMetrics    `json:"data"`
	Network   NetworkMetrics `json:"network"`
	Agents    AgentMetrics   `json:"agents"`
	Issues    IssueMetrics   `json:"issues"`
}

type UserMetrics struct {
	TotalUsers      int `json:"total_users"`
	ActiveUsers     int `json:"active_users"`
	NewMonthlyUsers int `json:"new_monthly_users"`
	NewVoucherUsers int `json:"new_voucher_users"`
}

type SessionMetrics struct {
	TotalSessions  int `json:"total_sessions"`
	ActiveSessions int `json:"active_sessions"`
}

type DataMetrics struct {
	MonthlyDataUsedGB float64 `json:"monthly_data_used_gb"`
	VoucherDataUsedGB float64 `json:"voucher_data_used_gb"`
}

type NetworkMetrics struct {
	Nodes   DeviceCount `json:"nodes"`
	Servers DeviceCount `json:"servers"`
	CPEs    DeviceCount `json:"cpes"`
}

type DeviceCount struct {
	Active int `json:"active"`
	Total  int `json:"total"`
}

type AgentMetrics struct {
	TotalAgents   int     `json:"total_agents"`
	TotalBatches int     `json:"total_batches"`
	Revenue       float64 `json:"revenue"`
	MRR           float64 `json:"mrr"`
}

type IssueMetrics struct {
	DeviceIssues int `json:"device_issues"`
	NetworkIssues int `json:"network_issues"`
}

func (s *ReportingService) GetSummary(providerID int64, period string, startDate, endDate time.Time) (*ReportingSummary, error) {
	summary := &ReportingSummary{
		Period:    period,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
	}

	var snapshots []domain.DailySnapshot
	err := s.db.Where("provider_id = ? AND snapshot_date BETWEEN ? AND ?",
		providerID, startDate, endDate).Find(&snapshots).Error
	if err != nil {
		return nil, err
	}

	if len(snapshots) > 0 {
		for _, snap := range snapshots {
			summary.Users.TotalUsers += snap.TotalUsers
			summary.Users.ActiveUsers += snap.ActiveUsers
			summary.Users.NewMonthlyUsers += snap.NewMonthlyUsers
			summary.Users.NewVoucherUsers += snap.NewVoucherUsers
			summary.Sessions.TotalSessions += snap.TotalSessions
			summary.Sessions.ActiveSessions += snap.ActiveSessions
			summary.Data.MonthlyDataUsedGB += float64(snap.MonthlyDataUsedBytes) / (1024 * 1024 * 1024)
			summary.Data.VoucherDataUsedGB += float64(snap.VoucherDataUsedBytes) / (1024 * 1024 * 1024)
			summary.Network.Nodes.Active += snap.ActiveNodes
			summary.Network.Nodes.Total += snap.TotalNodes
			summary.Network.Servers.Active += snap.ActiveServers
			summary.Network.Servers.Total += snap.TotalServers
			summary.Network.CPEs.Active += snap.ActiveCPEs
			summary.Network.CPEs.Total += snap.TotalCPEs
			summary.Agents.TotalAgents = maxInt(summary.Agents.TotalAgents, snap.TotalAgents)
			summary.Agents.TotalBatches += snap.TotalBatches
			summary.Agents.Revenue += snap.AgentRevenue
			summary.Agents.MRR += snap.MRR
			summary.Issues.DeviceIssues += snap.DeviceIssues
			summary.Issues.NetworkIssues += snap.NetworkIssues
		}
		return summary, nil
	}

	return s.getRealTimeSummary(providerID)
}

func (s *ReportingService) getRealTimeSummary(providerID int64) (*ReportingSummary, error) {
	summary := &ReportingSummary{
		Period:    "realtime",
		StartDate: time.Now().Format("2006-01-02"),
		EndDate:   time.Now().Format("2006-01-02"),
	}

	var userCount, activeUserCount, sessionCount int64
	s.db.Model(&domain.RadiusUser{}).Where("tenant_id = ?", providerID).Count(&userCount)
	s.db.Model(&domain.RadiusUser{}).Where("tenant_id = ? AND status = ?", providerID, "active").Count(&activeUserCount)
	s.db.Model(&domain.RadiusAccounting{}).Where("tenant_id = ? AND acctstoptime IS NULL", providerID).Count(&sessionCount)

	summary.Users.TotalUsers = int(userCount)
	summary.Users.ActiveUsers = int(activeUserCount)
	summary.Sessions.ActiveSessions = int(sessionCount)

	s.getNetworkStatus(providerID, summary)

	agentMetrics, err := s.GetAgentMetrics(providerID)
	if err == nil {
		summary.Agents = *agentMetrics
	}

	return summary, nil
}

func (s *ReportingService) GetNetworkStatus(providerID int64) (*NetworkMetrics, error) {
	metrics := &NetworkMetrics{}

	var totalNodes, activeNodes, totalServers, activeServers int64

	s.db.Model(&domain.NetNode{}).Where("tenant_id = ?", providerID).Count(&totalNodes)
	s.db.Model(&domain.NetNode{}).Where("tenant_id = ? AND status = ?", providerID, "active").Count(&activeNodes)
	s.db.Model(&domain.Server{}).Where("tenant_id = ?", providerID).Count(&totalServers)
	s.db.Model(&domain.Server{}).Where("tenant_id = ? AND status = ?", providerID, "active").Count(&activeServers)

	metrics.Nodes.Total = int(totalNodes)
	metrics.Nodes.Active = int(activeNodes)
	metrics.Servers.Total = int(totalServers)
	metrics.Servers.Active = int(activeServers)
	metrics.CPEs.Total = 0
	metrics.CPEs.Active = 0

	return metrics, nil
}

func (s *ReportingService) getNetworkStatus(providerID int64, summary *ReportingSummary) error {
	var totalNodes, activeNodes, totalServers, activeServers int64

	s.db.Model(&domain.NetNode{}).Where("tenant_id = ?", providerID).Count(&totalNodes)
	s.db.Model(&domain.NetNode{}).Where("tenant_id = ? AND status = ?", providerID, "active").Count(&activeNodes)
	s.db.Model(&domain.Server{}).Where("tenant_id = ?", providerID).Count(&totalServers)
	s.db.Model(&domain.Server{}).Where("tenant_id = ? AND status = ?", providerID, "active").Count(&activeServers)

	summary.Network.Nodes.Total = int(totalNodes)
	summary.Network.Nodes.Active = int(activeNodes)
	summary.Network.Servers.Total = int(totalServers)
	summary.Network.Servers.Active = int(activeServers)
	summary.Network.CPEs.Total = 0
	summary.Network.CPEs.Active = 0

	return nil
}

func (s *ReportingService) GetAgentMetrics(providerID int64) (*AgentMetrics, error) {
	metrics := &AgentMetrics{}

	var batchCount, agentCount int64
	var revenue, mrr float64

	s.db.Model(&domain.VoucherBatch{}).Where("tenant_id = ?", providerID).Count(&batchCount)

	s.db.Model(&domain.SysOpr{}).Where("tenant_id = ? AND level = ?", providerID, "agent").Count(&agentCount)

	s.db.Model(&domain.CommissionLog{}).
		Joins("JOIN sys_opr ON sys_opr.id = commission_log.agent_id").
		Where("sys_opr.tenant_id = ? AND sys_opr.level = ?", providerID, "agent").
		Select("COALESCE(SUM(commission_log.amount), 0)").
		Scan(&revenue)

	s.db.Model(&domain.VoucherSubscription{}).Where("tenant_id = ? AND status = ?", providerID, "active").Select("COALESCE(SUM(price), 0)").Scan(&mrr)

	metrics.TotalAgents = int(agentCount)
	metrics.TotalBatches = int(batchCount)
	metrics.Revenue = revenue
	metrics.MRR = mrr

	return metrics, nil
}

func (s *ReportingService) GetOpenIssues(providerID int64, limit int) ([]domain.NetworkIssue, error) {
	var issues []domain.NetworkIssue
	err := s.db.Where("provider_id = ? AND status = ?", providerID, "open").
		Order("created_at DESC").
		Limit(limit).
		Find(&issues).Error
	return issues, err
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (s *ReportingService) AggregateDailySnapshots() error {
	yesterday := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)

	var providers []domain.Provider
	if err := s.db.Where("status = ?", "active").Find(&providers).Error; err != nil {
		return err
	}

	for _, provider := range providers {
		if err := s.createOrUpdateSnapshot(provider.ID, yesterday); err != nil {
			zap.S().Errorf("failed to create snapshot for provider %d: %s", provider.ID, err.Error())
		}
	}

	return nil
}

func (s *ReportingService) createOrUpdateSnapshot(providerID int64, snapshotDate time.Time) error {
	snapshot := domain.DailySnapshot{
		ProviderID:     providerID,
		SnapshotDate:   snapshotDate,
	}

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	_ = snapshotDate // Used for date context

	var userCount, activeUserCount int64
	s.db.Model(&domain.RadiusUser{}).Where("tenant_id = ?", providerID).Count(&userCount)
	s.db.Model(&domain.RadiusUser{}).Where("tenant_id = ? AND status = ?", providerID, "active").Count(&activeUserCount)
	snapshot.TotalUsers = int(userCount)
	snapshot.ActiveUsers = int(activeUserCount)

	var newMonthlyUsers int64
	s.db.Model(&domain.RadiusUser{}).Where("tenant_id = ? AND created_at >= ? AND created_at < ?", providerID, monthStart, now).Count(&newMonthlyUsers)
	snapshot.NewMonthlyUsers = int(newMonthlyUsers)

	var activeSessions int64
	s.db.Model(&domain.RadiusAccounting{}).Where("tenant_id = ? AND acctstoptime IS NULL", providerID).Count(&activeSessions)
	snapshot.ActiveSessions = int(activeSessions)

	var monthlyData int64
	s.db.Model(&domain.RadiusAccounting{}).
		Where("tenant_id = ? AND acctstarttime >= ?", providerID, monthStart).
		Select("COALESCE(SUM(acctinputoctets + acctoutputoctets), 0)").
		Scan(&monthlyData)
	snapshot.MonthlyDataUsedBytes = monthlyData

	snapshot.NewVoucherUsers = 0
	snapshot.TotalSessions = 0

	var totalNodes, activeNodes, totalServers, activeServers int64
	s.db.Model(&domain.NetNode{}).Where("tenant_id = ?", providerID).Count(&totalNodes)
	s.db.Model(&domain.NetNode{}).Where("tenant_id = ? AND status = ?", providerID, "active").Count(&activeNodes)
	s.db.Model(&domain.Server{}).Where("tenant_id = ?", providerID).Count(&totalServers)
	s.db.Model(&domain.Server{}).Where("tenant_id = ? AND status = ?", providerID, "active").Count(&activeServers)
	snapshot.TotalNodes = int(totalNodes)
	snapshot.ActiveNodes = int(activeNodes)
	snapshot.TotalServers = int(totalServers)
	snapshot.ActiveServers = int(activeServers)

	var batchCount int64
	var revenue, mrr float64
	s.db.Model(&domain.VoucherBatch{}).Where("tenant_id = ?", providerID).Count(&batchCount)
	s.db.Model(&domain.CommissionLog{}).
		Joins("JOIN sys_opr ON sys_opr.id = commission_log.agent_id").
		Where("sys_opr.tenant_id = ? AND sys_opr.level = ?", providerID, "agent").
		Select("COALESCE(SUM(commission_log.amount), 0)").
		Scan(&revenue)
	s.db.Model(&domain.VoucherSubscription{}).Where("tenant_id = ? AND status = ?", providerID, "active").Select("COALESCE(SUM(price), 0)").Scan(&mrr)
	snapshot.TotalBatches = int(batchCount)
	snapshot.AgentRevenue = revenue
	snapshot.MRR = mrr

	var openIssues int64
	s.db.Model(&domain.NetworkIssue{}).Where("provider_id = ? AND status = ?", providerID, "open").Count(&openIssues)
	snapshot.DeviceIssues = int(openIssues)
	snapshot.NetworkIssues = 0

	return s.db.Save(&snapshot).Error
}
