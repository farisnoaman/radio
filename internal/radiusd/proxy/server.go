// Package proxy implements RADIUS proxy functionality for forwarding
// authentication and accounting requests to upstream RADIUS servers.
//
// The proxy supports:
//   - Realm-based routing (forward based on username suffix)
//   - Load balancing across multiple upstream servers
//   - Automatic failover on server unavailability
//   - Request logging for auditing
//
// Example:
//
//	proxy := proxy.NewProxyServer(proxy.Config{
//	    ListenAddr: ":1814",
//	    Timeout:    5 * time.Second,
//	})
//	proxy.Start(ctx)
package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/repository"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProxyConfig holds the configuration for the RADIUS proxy server.
type ProxyConfig struct {
	// ListenAddr is the address to listen on for proxy requests.
	// Format: "host:port" (e.g., ":1814" for all interfaces, port 1814)
	ListenAddr string

	// Timeout is the maximum time to wait for upstream server response.
	Timeout time.Duration

	// MaxConcurrentRequests is the maximum number of concurrent proxy requests.
	MaxConcurrentRequests int

	// EnableLogging enables request/response logging.
	EnableLogging bool
}

// ProxyServer handles RADIUS proxy operations.
type ProxyServer struct {
	config   ProxyConfig
	db       *gorm.DB
	repo     *repository.ProxyRepository
	client   *ProxyClient
	conn     *net.UDPConn
	packetCh chan *proxyRequest
	shutdown chan struct{}
	wg       sync.WaitGroup
}

// proxyRequest represents an incoming proxy request.
type proxyRequest struct {
	pkt     *radius.Packet
	srcAddr *net.UDPAddr
	ctx     context.Context
}

// NewProxyServer creates a new RADIUS proxy server.
func NewProxyServer(db *gorm.DB, config ProxyConfig) *ProxyServer {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.MaxConcurrentRequests == 0 {
		config.MaxConcurrentRequests = 100
	}

	return &ProxyServer{
		config:   config,
		db:       db,
		repo:     repository.NewProxyRepository(db),
		client:   NewProxyClient(config.Timeout),
		packetCh: make(chan *proxyRequest, 100),
		shutdown: make(chan struct{}),
	}
}

// Start starts the proxy server.
func (s *ProxyServer) Start(ctx context.Context) error {
	// Create UDP listener
	conn, err := net.ListenPacket("udp", s.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.config.ListenAddr, err)
	}
	s.conn = conn.(*net.UDPConn)

	zap.S().Info("RADIUS proxy server started",
		zap.String("addr", s.config.ListenAddr))

	// Start request handler workers
	for i := 0; i < 10; i++ {
		s.wg.Add(1)
		go s.requestWorker()
	}

	// Start packet receiver
	s.wg.Add(1)
	go s.packetReceiver(ctx)

	// Wait for shutdown
	<-ctx.Done()
	return s.Shutdown()
}

// Shutdown gracefully shuts down the proxy server.
func (s *ProxyServer) Shutdown() error {
	close(s.shutdown)
	s.wg.Wait()

	if s.conn != nil {
		s.conn.Close()
	}

	zap.S().Info("RADIUS proxy server stopped")
	return nil
}

// packetReceiver receives RADIUS packets from the network.
func (s *ProxyServer) packetReceiver(ctx context.Context) {
	defer s.wg.Done()

	buf := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.shutdown:
			return
		default:
			// Set read deadline
			s.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			n, srcAddr, err := s.conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Normal timeout, continue
				}
				zap.S().Error("Packet receive error", zap.Error(err))
				continue
			}

			// Parse RADIUS packet
			pkt, err := radius.Parse(buf[:n], nil)
			if err != nil {
				zap.S().Warn("Failed to parse RADIUS packet",
					zap.String("src", srcAddr.String()),
					zap.Error(err))
				continue
			}

			// Queue request for processing
			req := &proxyRequest{
				pkt:     pkt,
				srcAddr: srcAddr,
				ctx:     ctx,
			}

			select {
			case s.packetCh <- req:
			default:
				zap.S().Warn("Request queue full, dropping packet",
					zap.String("src", srcAddr.String()))
			}
		}
	}
}

// requestWorker processes proxy requests.
func (s *ProxyServer) requestWorker() {
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdown:
			return
		case req := <-s.packetCh:
			s.handleRequest(req)
		}
	}
}

// handleRequest processes a single proxy request.
func (s *ProxyServer) handleRequest(req *proxyRequest) {
	start := time.Now()

	// Extract username from packet
	username := rfc2865.UserName_GetString(req.pkt)
	if username == "" {
		zap.S().Warn("Proxy request without username")
		return
	}

	// Find matching realm
	realm, err := s.repo.FindRealmForUsername(req.ctx, username)
	if err != nil {
		zap.S().Error("Failed to find realm",
			zap.String("username", username),
			zap.Error(err))
		s.sendError(req.pkt, req.srcAddr)
		return
	}

	if realm == nil {
		// No matching realm, send NAK
		zap.S().Debug("No matching realm for username",
			zap.String("username", username))
		s.sendNAK(req.pkt, req.srcAddr)
		return
	}

	// Get proxy servers for realm
	var servers []*domain.RadiusProxyServer
	for _, serverID := range realm.ProxyServers {
		server, err := s.repo.GetServerByID(req.ctx, serverID)
		if err != nil {
			zap.S().Error("Failed to get proxy server",
				zap.Int64("server_id", serverID),
				zap.Error(err))
			continue
		}
		if server != nil && server.Status == "enabled" {
			servers = append(servers, server)
		}
	}

	if len(servers) == 0 {
		zap.S().Warn("No available proxy servers for realm",
			zap.String("realm", realm.Realm))
		s.sendNAK(req.pkt, req.srcAddr)
		return
	}

	// Try servers in order (failover)
	var lastErr error
	for _, server := range servers {
		// Forward request to upstream server
		resp, err := s.client.Forward(req.ctx, req.pkt, server)
		if err != nil {
			lastErr = err
			zap.S().Debug("Proxy server failed, trying next",
				zap.String("server", server.Name),
				zap.Error(err))
			continue
		}

		// Log successful proxy
		s.logRequest(username, realm.Realm, server.ID, req.pkt.Code, true, time.Since(start))

		// Send response back to client
		s.sendResponse(resp, req.srcAddr)
		return
	}

	// All servers failed
	zap.S().Error("All proxy servers failed",
		zap.String("realm", realm.Realm),
		zap.Error(lastErr))

	s.logRequest(username, realm.Realm, servers[0].ID, req.pkt.Code, false, time.Since(start))
	s.sendNAK(req.pkt, req.srcAddr)
}

// sendResponse sends a RADIUS response packet.
func (s *ProxyServer) sendResponse(pkt *radius.Packet, dst *net.UDPAddr) {
	encoded, err := pkt.Encode()
	if err != nil {
		zap.S().Error("Failed to encode packet",
			zap.String("dst", dst.String()),
			zap.Error(err))
		return
	}

	_, err = s.conn.WriteToUDP(encoded, dst)
	if err != nil {
		zap.S().Error("Failed to send response",
			zap.String("dst", dst.String()),
			zap.Error(err))
	}
}

// sendNAK sends a Reject/NAK response.
func (s *ProxyServer) sendNAK(pkt *radius.Packet, dst *net.UDPAddr) {
	var code radius.Code
	switch pkt.Code {
	case radius.CodeAccessRequest:
		code = radius.CodeAccessReject
	case radius.CodeAccountingRequest:
		code = radius.CodeAccountingResponse
	case radius.CodeCoARequest:
		code = radius.CodeCoANAK
	default:
		code = radius.CodeAccessReject
	}

	nak := radius.New(code, pkt.Secret)
	s.sendResponse(nak, dst)
}

// sendError sends an error response.
func (s *ProxyServer) sendError(pkt *radius.Packet, dst *net.UDPAddr) {
	nak := radius.New(radius.CodeAccessReject, pkt.Secret)
	s.sendResponse(nak, dst)
}

// logRequest logs a proxy request to the database.
func (s *ProxyServer) logRequest(
	username string,
	realm string,
	serverID int64,
	requestCode radius.Code,
	success bool,
	latency time.Duration,
) {
	if !s.config.EnableLogging {
		return
	}

	reqType := "auth"
	if requestCode == radius.CodeAccountingRequest {
		reqType = "acct"
	}

	log := &domain.ProxyRequestLog{
		Realm:       realm,
		Username:    username,
		ServerID:    serverID,
		RequestType: reqType,
		Success:     success,
		LatencyMs:   int(latency.Milliseconds()),
	}

	if err := s.repo.LogProxyRequest(context.Background(), log); err != nil {
		zap.S().Error("Failed to log proxy request", zap.Error(err))
	}
}
