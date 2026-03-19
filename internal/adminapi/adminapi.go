package adminapi

import (
	"github.com/talkincode/toughradius/v9/internal/app"
)

// Init registers all admin API routes
func Init(appCtx app.AppContext) {
	registerAuthRoutes()
	registerUserRoutes()
	registerDashboardRoutes()
	registerProfileRoutes()
	registerAccountingRoutes()
	registerSessionRoutes()
	registerNASRoutes()
	registerServerRoutes()
	registerDiscoveryRoutes()
	registerSettingsRoutes()
	registerNodesRoutes()
	registerOperatorsRoutes()
	registerProductRoutes()
	registerVoucherRoutes()
	registerVoucherTemplateRoutes()
	registerAgentRoutes()
	registerAgentHierarchyRoutes()
	registerFinancialRoutes()
	registerSystemLogRoutes()
	registerPortalSessionRoutes()
	registerPortalVoucherRoutes()
	registerPortalUserRoutes()
	registerBackupRoutes()
	registerMaintenanceRoutes()
	registerWebsocketRoutes()
	registerPrivacyRoutes()
	registerTopologyRoutes()
	registerTunnelRoutes()
	registerAnalyticsRoutes()
	registerCoARoutes()
	registerCPERoutes()
	registerInvoiceRoutes()
	registerProviderRoutes()
}


