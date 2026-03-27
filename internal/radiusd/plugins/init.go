package plugins

import (
	"time"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting/handlers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/checkers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/enhancers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/guards"

	// "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/guards"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/validators"
	eaphandlers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/handlers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"gorm.io/gorm"
)

// InitPlugins initializes all plugins
// sessionRepo and accountingRepo must be supplied externally to support dependency injection for plugins
func InitPlugins(
	appCtx app.ConfigManagerProvider,
	db *gorm.DB,
	sessionRepo repository.SessionRepository,
	accountingRepo repository.AccountingRepository,

	voucherRepo repository.VoucherRepository,
	userRepo repository.UserRepository,
) {
	// Register password validators (stateless plugins)
	registry.RegisterPasswordValidator(&validators.PAPValidator{})
	registry.RegisterPasswordValidator(&validators.CHAPValidator{})
	registry.RegisterPasswordValidator(&validators.MSCHAPValidator{})

	// Register profile checkers (mostly stateless)
	registry.RegisterPolicyChecker(&checkers.StatusChecker{})
	registry.RegisterPolicyChecker(&checkers.ExpireChecker{})
	registry.RegisterPolicyChecker(&checkers.MacBindChecker{})
	registry.RegisterPolicyChecker(&checkers.VlanBindChecker{})

	// Checkers that require dependency injection
	if sessionRepo != nil {
		registry.RegisterPolicyChecker(checkers.NewOnlineCountChecker(sessionRepo))
	}

	if accountingRepo != nil {
		registry.RegisterPolicyChecker(checkers.NewQuotaChecker(accountingRepo))
		registry.RegisterPolicyChecker(checkers.NewTimeQuotaChecker(accountingRepo))
	}

	if voucherRepo != nil && userRepo != nil {
		registry.RegisterPolicyChecker(checkers.NewFirstUseActivator(voucherRepo, userRepo))
	}

	// Initialize voucher batch cache and register voucher auth checker
	if voucherRepo != nil {
		var cacheDB *gorm.DB
		if db != nil {
			cacheDB = db
		}
		checkers.InitVoucherBatchCache(2*time.Minute, cacheDB)
		registry.RegisterPolicyChecker(checkers.NewVoucherAuthChecker(voucherRepo, checkers.GetVoucherBatchCache()))
	}

	// Register response enhancers
	registry.RegisterResponseEnhancer(enhancers.NewDefaultAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewHuaweiAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewH3CAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewZTEAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewMikrotikAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewIkuaiAcceptEnhancer())

	// Register authentication guards
	var cfgGetter interface{ GetInt64(string, string) int64 }
	if appCtx != nil {
		cfgGetter = appCtx.ConfigMgr()
	}
	registry.RegisterAuthGuard(guards.NewRejectDelayGuard(cfgGetter))

	// Register accounting handlers (dependency injection required)
	if sessionRepo != nil && accountingRepo != nil {
		registry.RegisterAccountingHandler(handlers.NewStartHandler(sessionRepo, accountingRepo))
		registry.RegisterAccountingHandler(handlers.NewUpdateHandler(sessionRepo))
		registry.RegisterAccountingHandler(handlers.NewStopHandler(sessionRepo, accountingRepo))
		registry.RegisterAccountingHandler(handlers.NewStopHandler(sessionRepo, accountingRepo))
		registry.RegisterAccountingHandler(handlers.NewNasStateHandler(sessionRepo))
		
		// Advanced Lifecycle: Session Logging
		if db != nil {
			registry.RegisterAccountingHandler(handlers.NewSessionLogHandler(db))
			// Voucher quota sync - updates voucher usage from accounting
			registry.RegisterAccountingHandler(handlers.NewVoucherQuotaSyncHandler(db))
			// Quota enforcement - disconnects users when time/data quota is exceeded during active sessions
			registry.RegisterAccountingHandler(handlers.NewQuotaEnforcementHandler(db, accountingRepo))
		}
	}


	// Register EAP handlers
	registry.RegisterEAPHandler(eaphandlers.NewMD5Handler())
	registry.RegisterEAPHandler(eaphandlers.NewOTPHandler())
	registry.RegisterEAPHandler(eaphandlers.NewMSCHAPv2Handler())

	// Register EAP-TLS handler (requires database for certificate validation)
	if db != nil {
		registry.RegisterEAPHandler(eaphandlers.NewTLSHandler(db))
	}

	// Vendor parsers under vendor/parsers register themselves via init()
}
