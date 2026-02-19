package acs

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Server implements the TR-069 ACS HTTP server.
// It handles CWMP requests from CPE devices and provides provisioning functionality.
type Server struct {
	httpServer  *http.Server
	provisioner *Provisioner
	repository  *Repository
	config      ACSConfig
	mux         *http.ServeMux
}

// NewServer creates a new ACS server instance.
//
// Parameters:
//   - config: ACS configuration
//   - provisioner: The provisioning engine
//   - repository: Database repository
//
// Returns:
//   - *Server: The server instance
func NewServer(config ACSConfig, provisioner *Provisioner, repository *Repository) *Server {
	s := &Server{
		provisioner: provisioner,
		repository:  repository,
		config:      config,
		mux:         http.NewServeMux(),
	}

	s.registerRoutes()
	return s
}

// registerRoutes registers all HTTP routes.
func (s *Server) registerRoutes() {
	// CWMP endpoint (the main TR-069 endpoint)
	s.mux.HandleFunc("/cpe", s.handleCPE)

	// Health check
	s.mux.HandleFunc("/health", s.handleHealth)

	// Admin API endpoints
	s.mux.HandleFunc("/api/v1/devices", s.listDevices)
	s.mux.HandleFunc("/api/v1/devices/", s.handleDevice)
	s.mux.HandleFunc("/api/v1/stats", s.getStats)
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Start starts the ACS HTTP server.
// This is a blocking call that runs until an error occurs or Shutdown is called.
func (s *Server) Start() error {
	addr := s.config.Listen
	if addr == "" {
		addr = ":7547"
	}

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

// StartTLS starts the ACS HTTPS server.
func (s *Server) StartTLS(certFile, keyFile string) error {
	addr := s.config.Listen
	if addr == "" {
		addr = ":7547"
	}

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s.httpServer.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// handleCPE is the main TR-069 CWMP endpoint handler.
// It processes Inform messages from CPE devices and returns appropriate responses.
func (s *Server) handleCPE(w http.ResponseWriter, r *http.Request) {
	// Check authentication if configured
	if s.config.Username != "" {
		if !s.authenticateRequest(r) {
			w.Header().Set("WWW-Authenticate", `Basic realm="TR-069 ACS"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendFault(w, "9002", "Internal error: failed to read request")
		return
	}

	// Parse SOAP request
	envelope, err := ParseSOAP(bytes.NewReader(body))
	if err != nil {
		s.sendFault(w, "9003", "Invalid request: "+err.Error())
		return
	}

	// Handle based on method type
	if envelope.Body.Inform != nil {
		s.handleInform(w, r, envelope.Body.Inform)
	} else if envelope.Body.GetParameterValues != nil {
		s.handleGetParameterValues(w, envelope)
	} else if envelope.Body.SetParameterValues != nil {
		s.handleSetParameterValues(w, envelope)
	} else if envelope.Body.Reboot != nil {
		s.handleReboot(w, envelope)
	} else {
		// Unknown method - return fault
		s.sendFault(w, "9016", "Unsupported RPC method")
	}
}

// authenticateRequest validates HTTP Basic authentication.
func (s *Server) authenticateRequest(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}

	// Parse Basic auth
	if !strings.HasPrefix(auth, "Basic ") {
		return false
	}

	encoded := strings.TrimPrefix(auth, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}

	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return false
	}

	username, password := credentials[0], credentials[1]
	return username == s.config.Username && password == s.config.Password
}

// handleInform processes an Inform message from a CPE.
func (s *Server) handleInform(w http.ResponseWriter, r *http.Request, inform *Inform) {
	// Get source IP
	sourceIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		sourceIP = strings.Split(forwarded, ",")[0]
	}

	// Process the Inform through provisioner
	device, provisionResult, err := s.provisioner.HandleInform(inform, sourceIP)
	if err != nil {
		s.sendFault(w, "9002", "Internal error: "+err.Error())
		return
	}

	_ = device // Device info logged for future use

	// Get message ID
	messageID := ""
	if r.Header.Get("X-CWMP-ID") != "" {
		messageID = r.Header.Get("X-CWMP-ID")
	} else {
		messageID = GenerateMessageID()
	}

	// Build response
	var responseXML []byte

	if provisionResult != nil && provisionResult.Configs != nil && len(provisionResult.Configs) > 0 {
		// Need to send configuration
		setParamXML, err := BuildSetParameterValues(messageID, provisionResult.Configs, "")
		if err != nil {
			s.sendFault(w, "9002", "Internal error: failed to build response")
			return
		}
		responseXML = setParamXML
	} else {
		// Just acknowledge the Inform
		responseXML, err = BuildInformResponse(messageID, 1)
		if err != nil {
			s.sendFault(w, "9002", "Internal error: failed to build response")
			return
		}
	}

	// Send response
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Write(responseXML)
}

// handleGetParameterValues handles GetParameterValues requests.
func (s *Server) handleGetParameterValues(w http.ResponseWriter, envelope *SOAPEnvelope) {
	responseXML, err := BuildSOAP(&SOAPEnvelope{
		Header: &SOAPHeader{ID: envelope.Header.ID},
		Body: SOAPBody{
			GetParameterValuesResponse: &GetParameterValuesResponse{
				ParameterList: []ParameterValueStruct{},
			},
		},
	})

	if err != nil {
		s.sendFault(w, "9002", "Internal error")
		return
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Write(responseXML)
}

// handleSetParameterValues handles SetParameterValues requests.
func (s *Server) handleSetParameterValues(w http.ResponseWriter, envelope *SOAPEnvelope) {
	responseXML, err := BuildSOAP(&SOAPEnvelope{
		Header: &SOAPHeader{ID: envelope.Header.ID},
		Body: SOAPBody{
			SetParameterValuesResponse: &SetParameterValuesResponse{
				Status: 0,
			},
		},
	})

	if err != nil {
		s.sendFault(w, "9002", "Internal error")
		return
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Write(responseXML)
}

// handleReboot handles Reboot requests.
func (s *Server) handleReboot(w http.ResponseWriter, envelope *SOAPEnvelope) {
	responseXML, err := BuildSOAP(&SOAPEnvelope{
		Header: &SOAPHeader{ID: envelope.Header.ID},
		Body: SOAPBody{
			RebootResponse: &RebootResponse{Status: 0},
		},
	})

	if err != nil {
		s.sendFault(w, "9002", "Internal error")
		return
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Write(responseXML)
}

// sendFault sends a SOAP fault response.
func (s *Server) sendFault(w http.ResponseWriter, faultCode, faultString string) {
	responseXML, _ := BuildFault(GenerateMessageID(), "Client", faultString, 9000, faultString)
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Write(responseXML)
}

// handleHealth returns server health status.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

// listDevices returns a list of all CPE devices.
func (s *Server) listDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	devices, err := s.repository.ListDevices("", 100, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"devices": devices,
		"total":   len(devices),
	})
}

// handleDevice handles device-specific API endpoints.
func (s *Server) handleDevice(w http.ResponseWriter, r *http.Request) {
	// Extract device ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/devices/")
	if path == "" {
		http.Error(w, "Device ID required", http.StatusBadRequest)
		return
	}

	var deviceID int64
	fmt.Sscanf(path, "%d", &deviceID)

	switch r.Method {
	case http.MethodGet:
		device, err := s.repository.GetDeviceByID(deviceID)
		if err != nil {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(device)

	case http.MethodDelete:
		if err := s.repository.DeleteDevice(deviceID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodPost:
		// Check for provision or reprovision action
		action := r.URL.Query().Get("action")
		if action == "provision" {
			s.provisionDevice(deviceID, w, r)
		} else if action == "reprovision" {
			s.reprovisionDevice(deviceID, w, r)
		} else {
			http.Error(w, "Unknown action", http.StatusBadRequest)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// provisionDevice manually provisions a device.
func (s *Server) provisionDevice(deviceID int64, w http.ResponseWriter, r *http.Request) {
	result, err := s.provisioner.ProvisionDeviceManual(deviceID, "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username": result.PPPoEUsername,
		"status":   "provisioned",
	})
}

// reprovisionDevice reprovisions a device with new credentials.
func (s *Server) reprovisionDevice(deviceID int64, w http.ResponseWriter, r *http.Request) {
	result, err := s.provisioner.ReprovisionDevice(deviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username": result.PPPoEUsername,
		"status":   "reprovisioned",
	})
}

// getStats returns ACS statistics.
func (s *Server) getStats(w http.ResponseWriter, r *http.Request) {
	pending, _ := s.repository.CountDevicesByStatus(DeviceStatusPending)
	provisioned, _ := s.repository.CountDevicesByStatus(DeviceStatusProvisioned)
	failed, _ := s.repository.CountDevicesByStatus(DeviceStatusFailed)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{
		"pending":     pending,
		"provisioned": provisioned,
		"failed":      failed,
	})
}
