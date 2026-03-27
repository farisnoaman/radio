package adminapi

import (
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/backup"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// backupService is the global backup service instance
var backupService *backup.BackupService

// Init registers all admin API routes
func Init(appCtx app.AppContext) {
	// Initialize backup service
	db := appCtx.DB()
	backupService = backup.NewBackupService(db, nil)

	// Add backup service to the app context for retrieval in handlers
	webserver.SetBackupService(backupService)

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
	registerPortalLoyaltyRoutes()
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
	registerProviderRegistrationRoutes()
	registerBillingRoutes()
	registerMonitoringRoutes()
	registerProviderBackupRoutes()
	registerPortalNotificationRoutes()
	registerReportingRoutes(appCtx)
	registerNASTemplateRoutes()
	registerDeviceManagementRoutes()
	registerDeviceRoutes()
	registerProxyRoutes()
	registerCertificateRoutes()
	registerTrafficAnalysisRoutes()
	registerEnvMonitoringRoutes()
}
