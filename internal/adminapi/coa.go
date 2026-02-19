package adminapi

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/coa"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"go.uber.org/zap"
)

// CoARequest represents a CoA request from the API.
type CoARequest struct {
	// SessionTimeout is the new session timeout in seconds.
	SessionTimeout int `json:"session_timeout"`
	// UpRate is the new upload rate in Kbps.
	UpRate int `json:"up_rate"`
	// DownRate is the new download rate in Kbps.
	DownRate int `json:"down_rate"`
	// DataQuota is the new data quota in MB.
	DataQuota int64 `json:"data_quota"`
	// Reason is the reason for the CoA operation.
	Reason string `json:"reason"`
}

// DisconnectRequest represents a disconnect request from the API.
type DisconnectRequest struct {
	// Reason is the reason for the disconnect.
	Reason string `json:"reason"`
	// NotifyUser indicates whether to notify the user.
	NotifyUser bool `json:"notify_user"`
}

// CoAResponse represents the response from a CoA operation.
type CoAResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	NASResp   string `json:"nas_response,omitempty"`
	Duration  int64  `json:"duration_ms,omitempty"`
	RetryCnt  int    `json:"retry_count,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

// SendCoA sends a CoA-Request to modify an active session.
//
// @Summary Send CoA-Request to modify session
// @Tags CoA
// @Accept json
// @Produce json
// @Param id path int true "Session ID"
// @Param request body CoARequest true "CoA Request"
// @Success 200 {object} CoAResponse
// @Router /api/v1/sessions/{id}/coa [post]
func SendCoA(c echo.Context) error {
	// Get session ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid session ID", nil)
	}

	// Parse request body
	var req CoARequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	// Fetch session from database
	var session domain.RadiusOnline
	if err := GetDB(c).First(&session, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Session not found", err)
	}

	// Fetch NAS info
	var nas domain.NetNas
	if err := GetDB(c).Where("ipaddr = ?", session.NasAddr).First(&nas).Error; err != nil {
		return fail(c, http.StatusNotFound, "NAS_NOT_FOUND", "NAS device not found", err)
	}

	// Determine vendor code
	vendorCode := nas.VendorCode
	if vendorCode == "" {
		vendorCode = coa.VendorGeneric
	}

	// Create CoA request
	coaReq := coa.CoARequest{
		NASIP:          nas.Ipaddr,
		NASPort:        nas.CoaPort,
		Secret:         nas.Secret,
		Username:       session.Username,
		AcctSessionID:  session.AcctSessionId,
		VendorCode:     vendorCode,
		SessionTimeout: req.SessionTimeout,
		UpRate:         req.UpRate,
		DownRate:       req.DownRate,
		DataQuota:      req.DataQuota,
		Reason:         req.Reason,
	}

	// Send CoA request asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		client := coa.NewClient(coa.Config{
			Timeout:    5 * time.Second,
			RetryCount: 2,
		})

		// Build packet with vendor attributes
		// Note: The client handles basic packet, vendor attributes would need to be added
		// through the vendor builder interface

		resp := client.SendCoA(ctx, coaReq)

		if resp.Success {
			zap.L().Info("CoA-Request successful",
				zap.String("nas_addr", nas.Ipaddr),
				zap.String("username", session.Username),
				zap.Int64("session_id", id),
				zap.String("namespace", "adminapi"))
		} else {
			zap.L().Error("CoA-Request failed",
				zap.Error(resp.Error),
				zap.String("nas_addr", nas.Ipaddr),
				zap.String("username", session.Username),
				zap.Int64("session_id", id),
				zap.String("namespace", "adminapi"))
		}
	}()

	return ok(c, CoAResponse{
		Success: true,
		Message: "CoA request sent successfully",
	})
}

// DisconnectSessionEnhanced disconnects a user session using the enhanced CoA client.
//
// @Summary Disconnect user session (enhanced)
// @Tags CoA
// @Accept json
// @Produce json
// @Param id path int true "Session ID"
// @Param request body DisconnectRequest false "Disconnect Request"
// @Success 200 {object} CoAResponse
// @Router /api/v1/sessions/{id}/disconnect [post]
func DisconnectSessionEnhanced(c echo.Context) error {
	// Get session ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid session ID", nil)
	}

	// Parse request body (optional)
	var req DisconnectRequest
	if err := c.Bind(&req); err != nil {
		// Use defaults if no body provided
		req = DisconnectRequest{
			Reason:    "admin_initiated",
			NotifyUser: false,
		}
	}

	// Fetch session from database
	var session domain.RadiusOnline
	if err := GetDB(c).First(&session, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Session not found", err)
	}

	// Delete session from database first (best effort)
	if err := GetDB(c).Delete(&domain.RadiusOnline{}, id).Error; err != nil {
		zap.L().Error("Failed to delete session from database",
			zap.Error(err),
			zap.Int64("session_id", id),
			zap.String("namespace", "adminapi"))
	}

	// Fetch NAS info
	var nas domain.NetNas
	if err := GetDB(c).Where("ipaddr = ?", session.NasAddr).First(&nas).Error; err != nil {
		zap.L().Warn("NAS not found for CoA, session deleted from database only",
			zap.String("nas_addr", session.NasAddr),
			zap.String("username", session.Username),
			zap.String("namespace", "adminapi"))
		return ok(c, CoAResponse{
			Success: true,
			Message: "Session removed from database (NAS not found)",
		})
	}

	// Determine vendor code
	vendorCode := nas.VendorCode
	if vendorCode == "" {
		vendorCode = coa.VendorGeneric
	}

	// Determine CoA port
	coaPort := nas.CoaPort
	if coaPort <= 0 {
		coaPort = 3799 // Default CoA port
	}

	// Create disconnect request
	discReq := coa.DisconnectRequest{
		NASIP:         nas.Ipaddr,
		NASPort:       coaPort,
		Secret:        nas.Secret,
		Username:       session.Username,
		AcctSessionID: session.AcctSessionId,
		VendorCode:    vendorCode,
		Reason:        req.Reason,
	}

	// Send disconnect request synchronously with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := coa.NewClient(coa.Config{
		Timeout:    5 * time.Second,
		RetryCount: 2,
	})

	resp := client.SendDisconnect(ctx, discReq)

	if resp.Success {
		zap.L().Info("Disconnect-Request ACK received",
			zap.String("nas_addr", nas.Ipaddr),
			zap.String("username", session.Username),
			zap.String("acct_session_id", session.AcctSessionId),
			zap.String("namespace", "adminapi"))

		return ok(c, CoAResponse{
			Success:  true,
			Message:  "User disconnected successfully",
			NASResp:  "ACK",
			Duration: resp.Duration.Milliseconds(),
			RetryCnt: resp.RetryCount,
		})
	}

	// Handle failure
	errorMsg := "Unknown error"
	if resp.Error != nil {
		errorMsg = resp.Error.Error()
	}

	zap.L().Warn("Disconnect-Request failed",
		zap.String("nas_addr", nas.Ipaddr),
		zap.String("username", session.Username),
		zap.String("error", errorMsg),
		zap.String("namespace", "adminapi"))

	// Even if NAS fails, session is removed from DB
	return ok(c, CoAResponse{
		Success:   true,
		Message:   "Session removed (NAS disconnect may have failed)",
		NASResp:   "NAK",
		ErrorMsg:  errorMsg,
		Duration:  resp.Duration.Milliseconds(),
		RetryCnt:  resp.RetryCount,
	})
}

// BulkDisconnect disconnects multiple user sessions.
//
// @Summary Bulk disconnect sessions
// @Tags CoA
// @Accept json
// @Produce json
// @Param request body struct{SessionIDs []int64 `json:"session_ids"`} true "Session IDs"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/sessions/bulk-disconnect [post]
func BulkDisconnect(c echo.Context) error {
	type BulkRequest struct {
		SessionIDs []int64 `json:"session_ids"`
	}

	var req BulkRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if len(req.SessionIDs) == 0 {
		return fail(c, http.StatusBadRequest, "NO_SESSIONS", "No session IDs provided", nil)
	}

	// Process each session
	successCount := 0
	failCount := 0

	for _, sessionID := range req.SessionIDs {
		// Fetch session
		var session domain.RadiusOnline
		if err := GetDB(c).First(&session, sessionID).Error; err != nil {
			failCount++
			continue
		}

		// Fetch NAS
		var nas domain.NetNas
		if err := GetDB(c).Where("ipaddr = ?", session.NasAddr).First(&nas).Error; err != nil {
			// Delete from DB anyway
			GetDB(c).Delete(&domain.RadiusOnline{}, sessionID)
			successCount++
			continue
		}

		// Determine CoA port
		coaPort := nas.CoaPort
		if coaPort <= 0 {
			coaPort = 3799
		}

		// Create disconnect request
		discReq := coa.DisconnectRequest{
			NASIP:         nas.Ipaddr,
			NASPort:       coaPort,
			Secret:        nas.Secret,
			Username:      session.Username,
			AcctSessionID: session.AcctSessionId,
		}

		// Send disconnect (fire and forget for bulk)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			client := coa.NewClient(coa.Config{
				Timeout:    3 * time.Second,
				RetryCount: 1,
			})

			resp := client.SendDisconnect(ctx, discReq)
			if resp.Success {
				zap.L().Info("Bulk disconnect successful",
					zap.String("username", session.Username),
					zap.String("namespace", "coa"))
			}
		}()

		// Delete from database
		GetDB(c).Delete(&domain.RadiusOnline{}, sessionID)
		successCount++
	}

	return ok(c, map[string]interface{}{
		"message":       "Bulk disconnect completed",
		"total":        len(req.SessionIDs),
		"success":      successCount,
		"failed":       failCount,
	})
}

// registerCoARoutes registers CoA-related routes.
func registerCoARoutes() {
	webserver.ApiPOST("/sessions/:id/coa", SendCoA)
	webserver.ApiPOST("/sessions/:id/disconnect", DisconnectSessionEnhanced)
	webserver.ApiPOST("/sessions/bulk-disconnect", BulkDisconnect)
}
