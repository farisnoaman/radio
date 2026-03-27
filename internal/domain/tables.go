package domain

var Tables = []interface{}{
	// System
	&SysConfig{},
	&SysOpr{},
	&SysOprLog{},
	// Multi-Tenant Platform
	&Provider{},
	&ProviderRegistration{},
	// Network
	&NetNode{},
	&NetNas{},
	&Server{},
	// Radius
	&RadiusAccounting{},
	&RadiusOnline{},
	&RadiusProfile{},
	&RadiusUser{},
	// Commercial
	&Product{},
	&VoucherBatch{},
	&Voucher{},
	&AgentWallet{},
	&WalletLog{},
	&VoucherTopup{},
	&VoucherSubscription{},
	&VoucherBundle{},
	&VoucherBundleItem{},
	&VoucherTemplate{},
	// Agent Hierarchy & Commissions
	&AgentHierarchy{},
	&CommissionLog{},
	&CommissionSummary{},
	// Lifecycle
	&SessionLog{},
	// Billing
	&Invoice{},
	// Backup
	&BackupConfig{},
	&BackupRecord{},
	// NAS Templates (Phase 1)
	&NASTemplate{},
	// Device Management (Phase 1)
	&DeviceConfigBackup{},
	&SpeedTestResult{},
	&Location{},
	&NetworkDevice{},
	&NetworkDeviceMetric{},
	&NetworkDeviceAlert{},
	// Environment Monitoring
	&EnvironmentMetric{},
	&EnvironmentAlert{},
	&AlertConfig{},
	// RADIUS Proxy (Phase 2)
	&RadiusProxyServer{},
	&RadiusProxyRealm{},
	&ProxyRequestLog{},
}
