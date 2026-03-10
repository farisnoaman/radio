package domain

var Tables = []interface{}{
	// System
	&SysConfig{},
	&SysOpr{},
	&SysOprLog{},
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
	// Agent Hierarchy & Commissions
	&AgentHierarchy{},
	&CommissionLog{},
	&CommissionSummary{},
	// Lifecycle
	&SessionLog{},
	// Billing
	&Invoice{},
}

