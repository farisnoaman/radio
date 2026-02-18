package adminapi

import (
	"net/http"

	"path/filepath"



	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

func registerBackupRoutes() {
	webserver.ApiGET("/system/backup", ListBackups)
	webserver.ApiPOST("/system/backup", CreateBackup)
	webserver.ApiDELETE("/system/backup/:id", DeleteBackup)
	webserver.ApiPOST("/system/backup/:id/restore", RestoreBackup)
	webserver.ApiGET("/system/backup/:id/download", DownloadBackup)
}

func ListBackups(c echo.Context) error {
	backups, err := GetAppContext(c).BackupMgr().ListBackups()
	if err != nil {
		return fail(c, http.StatusInternalServerError, "BACKUP_LIST_ERROR", "Failed to list backups", err.Error())
	}
	return ok(c, backups)
}

func CreateBackup(c echo.Context) error {
	filename, err := GetAppContext(c).BackupMgr().CreateBackup()
	if err != nil {
		return fail(c, http.StatusInternalServerError, "BACKUP_CREATE_ERROR", "Failed to create backup", err.Error())
	}
	return ok(c, map[string]string{
		"filename": filename,
		"message":  "Backup created successfully",
	})
}

func DeleteBackup(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Backup ID is required", nil)
	}

	err := GetAppContext(c).BackupMgr().DeleteBackup(id)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "BACKUP_DELETE_ERROR", "Failed to delete backup", err.Error())
	}
	return ok(c, map[string]string{
		"message": "Backup deleted successfully",
	})
}

func RestoreBackup(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Backup ID is required", nil)
	}

	// This is a dangerous operation, should probably require extra confirmation or specific privileges
	// For now, assuming admin access is enough as per current auth model
	
	// Note: Restore might disrupt current connections.
	err := GetAppContext(c).BackupMgr().RestoreBackup(id)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "BACKUP_RESTORE_ERROR", "Failed to restore backup", err.Error())
	}
	
	return ok(c, map[string]string{
		"message": "Database restored successfully. Please restart the application if needed.",
	})
}

func DownloadBackup(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Backup ID is required", nil)
	}

	path, err := GetAppContext(c).BackupMgr().GetBackup(id)
	if err != nil {
		return fail(c, http.StatusNotFound, "BACKUP_NOT_FOUND", "Backup not found", err.Error())
	}

	return c.Attachment(path, filepath.Base(path))
}
