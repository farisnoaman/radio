package adminapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/backup"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// registerProviderBackupRoutes registers provider backup routes
func registerProviderBackupRoutes() {
	// Provider routes (tenant-isolated)
	webserver.ApiGET("/provider/backup", ListProviderBackups)
	webserver.ApiPOST("/provider/backup", CreateProviderBackup)
	webserver.ApiPOST("/provider/backup/:id/restore", RestoreProviderBackup)

	// Admin route (override)
	webserver.ApiPOST("/admin/provider/backup", AdminCreateBackup)
}

// ListProviderBackups lists backups for current tenant
func ListProviderBackups(c echo.Context) error {
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	backupSvc := getBackupService(c)
	if backupSvc == nil {
		return fail(c, http.StatusInternalServerError, "SERVICE_NOT_FOUND", "Backup service not initialized", nil)
	}

	backups, err := backupSvc.ListBackups(c.Request().Context(), tenantID)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "BACKUP_LIST_ERROR", "Failed to list backups", err)
	}

	return ok(c, backups)
}

// CreateProviderBackup creates a backup for current tenant
func CreateProviderBackup(c echo.Context) error {
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	backupSvc := getBackupService(c)
	if backupSvc == nil {
		return fail(c, http.StatusInternalServerError, "SERVICE_NOT_FOUND", "Backup service not initialized", nil)
	}

	record, err := backupSvc.CreateBackup(c.Request().Context(), tenantID, "manual")
	if err != nil {
		if err == backup.ErrBackupQuotaExceeded {
			return fail(c, http.StatusForbidden, "QUOTA_EXCEEDED", "Backup quota exceeded", nil)
		}
		return fail(c, http.StatusInternalServerError, "BACKUP_CREATE_ERROR", "Failed to create backup", err)
	}

	return ok(c, map[string]string{
		"message":  "Backup created successfully",
		"backup_id": fmt.Sprintf("%d", record.ID),
		"status":    record.Status,
	})
}

// RestoreProviderBackup restores a backup for current tenant
func RestoreProviderBackup(c echo.Context) error {
	backupID, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid backup ID", err)
	}

	tenantID, _ := tenant.FromContext(c.Request().Context())

	backupSvc := getBackupService(c)
	if backupSvc == nil {
		return fail(c, http.StatusInternalServerError, "SERVICE_NOT_FOUND", "Backup service not initialized", nil)
	}

	err = backupSvc.RestoreBackup(c.Request().Context(), tenantID, backupID)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "BACKUP_RESTORE_ERROR", "Failed to restore backup", err)
	}

	return ok(c, map[string]string{
		"message": "Backup restored successfully",
	})
}

// AdminCreateBackup creates a backup for any provider (admin only)
func AdminCreateBackup(c echo.Context) error {
	// Verify platform admin
	if !IsPlatformAdmin(c) {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Platform admin required", nil)
	}

	var req struct {
		TenantID int64  `json:"tenant_id" validate:"required"`
		Reason   string `json:"reason"`
	}

	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
	}

	backupSvc := getBackupService(c)
	if backupSvc == nil {
		return fail(c, http.StatusInternalServerError, "SERVICE_NOT_FOUND", "Backup service not initialized", nil)
	}

	record, err := backupSvc.AdminOverrideBackup(c.Request().Context(), req.TenantID, req.Reason)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "BACKUP_CREATE_ERROR", "Failed to create backup", err)
	}

	return ok(c, record)
}

// getBackupService returns the backup service from application context
func getBackupService(c echo.Context) *backup.BackupService {
	if svc, ok := c.Get("backupService").(*backup.BackupService); ok && svc != nil {
		return svc
	}
	return nil
}
