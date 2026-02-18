package app

import (
	"github.com/robfig/cron/v3"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app/backup"
	"github.com/talkincode/toughradius/v9/internal/app/logging"
	"github.com/talkincode/toughradius/v9/internal/app/maintenance"
	"github.com/talkincode/toughradius/v9/internal/app/tunnel"

	"github.com/talkincode/toughradius/v9/internal/app/websocket"



	"gorm.io/gorm"
)

// DBProvider provides database access
type DBProvider interface {
	DB() *gorm.DB
}

// ConfigProvider provides application configuration
type ConfigProvider interface {
	Config() *config.AppConfig
}

// SettingsProvider provides system settings access
type SettingsProvider interface {
	GetSettingsStringValue(category, key string) string
	GetSettingsInt64Value(category, key string) int64
	GetSettingsBoolValue(category, key string) bool
	SaveSettings(settings map[string]interface{}) error
}

// SchedulerProvider provides task scheduling capability
type SchedulerProvider interface {
	Scheduler() *cron.Cron
}

// ConfigManagerProvider provides configuration manager access
type ConfigManagerProvider interface {
	ConfigMgr() *ConfigManager
}

// ProfileCacheProvider provides profile cache access
type ProfileCacheProvider interface {
	ProfileCache() *ProfileCache
}

// ArchivalProvider provides archival manager access
type ArchivalProvider interface {
	ArchivalMgr() *logging.ArchivalManager
}

// AppContext combines all provider interfaces for full application context
// Services should depend on specific providers or this combined interface
type AppContext interface {
	DBProvider
	ConfigProvider
	SettingsProvider
	SchedulerProvider
	ConfigManagerProvider
	ProfileCacheProvider
	ArchivalProvider

	// Application lifecycle methods
	MigrateDB(track bool) error
	InitDb()
	DropAll()
	MaintMgr() *maintenance.MaintenanceManager

	BackupMgr() backup.BackupManager
	WsHub() *websocket.Hub
	TunnelMgr() tunnel.TunnelManager
}

