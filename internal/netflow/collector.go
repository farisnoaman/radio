// Package netflow implements NetFlow v9 collector for traffic analysis.
//
// NetFlow v9 (RFC 3954) provides network traffic flow information exported by
// routers and switches. This collector receives and processes these exports
// for traffic analysis, accounting, and security monitoring.
//
// Key features:
//   - Receive NetFlow v9 UDP packets
//   - Parse flow records from templates
//   - Store flows in database and time-series DB
//   - Aggregate traffic statistics
//   - Support for IPv6 flows
//
// Example:
//
//	collector := netflow.NewCollector(db, &netflow.CollectorConfig{
//	    ListenAddr: ":2056",
//	    BatchSize: 100,
//	})
//	go collector.Start(ctx)
package netflow

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CollectorConfig holds the configuration for the NetFlow collector.
type CollectorConfig struct {
	// ListenAddr is the UDP address to listen on (e.g., ":2056").
	ListenAddr string

	// BufferSize is the size of the receive buffer in bytes.
	BufferSize int

	// EnableLogging enables detailed logging of received flows.
	EnableLogging bool

	// BatchSize is the number of records to batch before database insert.
	BatchSize int

	// BatchTimeout is the maximum time to wait before flushing batch.
	BatchTimeout time.Duration
}

// Collector receives and processes NetFlow exports.
type Collector struct {
	config      *CollectorConfig
	db          *gorm.DB
	conn        *net.UDPConn
	packetCh    chan []byte
	batchCh     chan *domain.NetFlowRecord
	templates   map[uint32]map[uint16][]TemplateField // sourceID -> templateID -> fields
	templateMux sync.RWMutex
	shutdown    chan struct{}
	wg          sync.WaitGroup
}

// TemplateField represents a field in a NetFlow template.
type TemplateField struct {
	Type   uint16
	Length uint16
}

// NewCollector creates a new NetFlow collector.
func NewCollector(db *gorm.DB, config *CollectorConfig) *Collector {
	if config == nil {
		config = &CollectorConfig{}
	}
	if config.BufferSize == 0 {
		config.BufferSize = 9000 // Max NetFlow UDP packet size
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.BatchTimeout == 0 {
		config.BatchTimeout = 5 * time.Second
	}
	if config.ListenAddr == "" {
		config.ListenAddr = ":2056"
	}

	return &Collector{
		config:    config,
		db:        db,
		packetCh:  make(chan []byte, 100),
		batchCh:   make(chan *domain.NetFlowRecord, config.BatchSize*10),
		templates: make(map[uint32]map[uint16][]TemplateField),
		shutdown:  make(chan struct{}),
	}
}

// Start starts the NetFlow collector.
func (c *Collector) Start(ctx context.Context) error {
	// Create UDP listener
	addr, err := net.ResolveUDPAddr("udp", c.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", c.config.ListenAddr, err)
	}
	c.conn = conn

	zap.S().Info("NetFlow collector started",
		zap.String("addr", c.config.ListenAddr))

	// Start workers
	c.wg.Add(3)
	go c.packetReceiver(ctx)
	go c.packetProcessor(ctx)
	go c.batchWriter(ctx)

	// Wait for shutdown
	<-ctx.Done()
	return c.Shutdown()
}

// Shutdown gracefully shuts down the collector.
func (c *Collector) Shutdown() error {
	close(c.shutdown)
	c.wg.Wait()

	if c.conn != nil {
		c.conn.Close()
	}

	zap.S().Info("NetFlow collector stopped")
	return nil
}

// ParseHeader extracts NetFlow header fields.
func (c *Collector) ParseHeader(packet []byte) (version, count uint16, sourceID uint32) {
	if len(packet) < 24 {
		return 0, 0, 0
	}

	version = binary.BigEndian.Uint16(packet[0:2])
	count = binary.BigEndian.Uint16(packet[2:4])
	sourceID = binary.BigEndian.Uint32(packet[20:24])

	return
}

// ProcessPacket processes a single NetFlow packet.
func (c *Collector) ProcessPacket(packet []byte) (int, error) {
	// Parse NetFlow header
	if len(packet) < 24 {
		return 0, fmt.Errorf("packet too short: %d bytes", len(packet))
	}

	version := binary.BigEndian.Uint16(packet[0:2])
	if version != 9 {
		return 0, fmt.Errorf("unsupported version: %d", version)
	}

	// Extract header fields
	count := binary.BigEndian.Uint16(packet[2:4])
	sourceID := binary.BigEndian.Uint32(packet[20:24])
	unixSeconds := binary.BigEndian.Uint32(packet[8:12])
	unixNanoseconds := binary.BigEndian.Uint32(packet[12:16])

	timestamp := time.Unix(int64(unixSeconds), int64(unixNanoseconds))

	// Parse flow sets
	offset := 24
	recordsProcessed := 0

	for offset < len(packet) && recordsProcessed < int(count) {
		if offset+4 > len(packet) {
			break
		}

		// Get flow set header
		flowSetID := binary.BigEndian.Uint16(packet[offset : offset+2])
		flowSetLength := binary.BigEndian.Uint16(packet[offset+2 : offset+4])

		if offset+int(flowSetLength) > len(packet) {
			break
		}

		// Process flow set based on ID
		switch flowSetID {
		case 0: // Template Flow Set
			c.processTemplateFlowSet(packet[offset+4:offset+int(flowSetLength)], sourceID)

		case 1: // Options Template Flow Set
			// TODO: Implement options template

		case 256: // Data Flow Set
			records := c.decodeDataFlowSet(
				packet[offset+4:offset+int(flowSetLength)],
				sourceID,
				timestamp,
			)

			for _, record := range records {
				select {
				case c.batchCh <- record:
					recordsProcessed++
				default:
					zap.S().Warn("Batch channel full, dropping record")
				}
			}

		default:
			// Unknown flow set
			zap.S().Debug("Unknown flow set ID",
				zap.Uint16("flow_set_id", flowSetID))
		}

		offset += int(flowSetLength)
	}

	return recordsProcessed, nil
}

// processTemplateFlowSet processes a template flow set.
func (c *Collector) processTemplateFlowSet(data []byte, sourceID uint32) {
	offset := 0

	for offset < len(data) {
		if offset+4 > len(data) {
			break
		}

		templateID := binary.BigEndian.Uint16(data[offset : offset+2])
		fieldCount := binary.BigEndian.Uint16(data[offset+2 : offset+4])
		offset += 4

		fields := make([]TemplateField, 0, fieldCount)

		for i := 0; i < int(fieldCount) && offset+4 <= len(data); i++ {
			fieldType := binary.BigEndian.Uint16(data[offset : offset+2])
			fieldLength := binary.BigEndian.Uint16(data[offset+2 : offset+4])
			offset += 4

			fields = append(fields, TemplateField{
				Type:   fieldType,
				Length: fieldLength,
			})
		}

		// Store template
		c.templateMux.Lock()
		if c.templates[sourceID] == nil {
			c.templates[sourceID] = make(map[uint16][]TemplateField)
		}
		c.templates[sourceID][templateID] = fields
		c.templateMux.Unlock()
	}
}

// decodeTemplate decodes a NetFlow v9 template.
func (c *Collector) decodeTemplate(data []byte) map[uint16][]TemplateField {
	offset := 0
	templates := make(map[uint16][]TemplateField)

	for offset < len(data) {
		if offset+4 > len(data) {
			break
		}

		templateID := binary.BigEndian.Uint16(data[offset : offset+2])
		fieldCount := binary.BigEndian.Uint16(data[offset+2 : offset+4])
		offset += 4

		fields := make([]TemplateField, 0, fieldCount)

		for i := 0; i < int(fieldCount) && offset+4 <= len(data); i++ {
			fieldType := binary.BigEndian.Uint16(data[offset : offset+2])
			fieldLength := binary.BigEndian.Uint16(data[offset+2 : offset+4])
			offset += 4

			fields = append(fields, TemplateField{
				Type:   fieldType,
				Length: fieldLength,
			})
		}

		templates[templateID] = fields
	}

	return templates
}

// decodeDataFlowSet decodes a data flow set using templates.
func (c *Collector) decodeDataFlowSet(
	data []byte,
	sourceID uint32,
	timestamp time.Time,
) []*domain.NetFlowRecord {
	c.templateMux.RLock()
	templates, ok := c.templates[sourceID]
	c.templateMux.RUnlock()

	if !ok || len(templates) == 0 {
		// No template yet
		return nil
	}

	// Use first template (simplified)
	var fields []TemplateField
	for _, f := range templates {
		fields = f
		break
	}

	if fields == nil {
		return nil
	}

	records := make([]*domain.NetFlowRecord, 0)
	offset := 0

	// Decode records based on template
	for offset < len(data) {
		record := &domain.NetFlowRecord{
			RouterID:      fmt.Sprintf("router-%d", sourceID),
			FirstSwitched: timestamp,
			LastSwitched:  timestamp,
		}

		// Parse fields from template
		for _, field := range fields {
			if offset+int(field.Length) > len(data) {
				break
			}

			fieldData := data[offset : offset+int(field.Length)]
			offset += int(field.Length)

			c.parseField(record, field.Type, fieldData)
		}

		// Validate record
		if err := record.Validate(); err == nil {
			records = append(records, record)
		}
	}

	return records
}

// parseField parses a single field and updates the record.
func (c *Collector) parseField(record *domain.NetFlowRecord, fieldType uint16, data []byte) {
	switch fieldType {
	case 8: // IPV4_SRC_ADDR
		if len(data) == 4 {
			record.SourceAddr = fmt.Sprintf("%d.%d.%d.%d", data[0], data[1], data[2], data[3])
		}
	case 12: // IPV4_DST_ADDR
		if len(data) == 4 {
			record.DestAddr = fmt.Sprintf("%d.%d.%d.%d", data[0], data[1], data[2], data[3])
		}
	case 27: // IPV4_SRC_MASK
	case 28: // IPV4_DST_MASK
	case 4: // PROTOCOL
		if len(data) == 1 {
			record.Protocol = data[0]
		}
	case 5: // TOS
		if len(data) == 1 {
			record.Tos = data[0]
		}
	case 6: // TCP_FLAGS
		if len(data) == 1 {
			record.TcpFlags = data[0]
		}
	case 1: // IN_BYTES
		if len(data) == 4 {
			record.Bytes = uint64(binary.BigEndian.Uint32(data))
		} else if len(data) == 8 {
			record.Bytes = binary.BigEndian.Uint64(data)
		}
	case 2: // IN_PKTS
		if len(data) == 4 {
			record.Packets = uint64(binary.BigEndian.Uint32(data))
		} else if len(data) == 8 {
			record.Packets = binary.BigEndian.Uint64(data)
		}
	case 7: // L4_SRC_PORT
		if len(data) == 2 {
			record.SourcePort = binary.BigEndian.Uint16(data)
		}
	case 10: // L4_DST_PORT (corrected field type)
		if len(data) == 2 {
			record.DestPort = binary.BigEndian.Uint16(data)
		}
	case 152: // FLOW_DURATION_MILLISECONDS
		if len(data) == 4 {
			record.FlowDuration = binary.BigEndian.Uint32(data)
		}
	// Add more field types as needed
	}
}

// packetReceiver receives NetFlow UDP packets.
func (c *Collector) packetReceiver(ctx context.Context) {
	defer c.wg.Done()

	buf := make([]byte, c.config.BufferSize)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.shutdown:
			return
		default:
			// Set read deadline
			c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			n, srcAddr, err := c.conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Normal timeout
				}
				zap.S().Error("Packet receive error", zap.Error(err))
				continue
			}

			// Copy packet data
			packet := make([]byte, n)
			copy(packet, buf[:n])

			if c.config.EnableLogging {
				zap.S().Debug("NetFlow packet received",
					zap.String("src", srcAddr.String()),
					zap.Int("size", n))
			}

			select {
			case c.packetCh <- packet:
			default:
				zap.S().Warn("Packet channel full, dropping packet")
			}
		}
	}
}

// packetProcessor processes NetFlow packets.
func (c *Collector) packetProcessor(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.shutdown:
			return
		case packet := <-c.packetCh:
			c.ProcessPacket(packet)
		}
	}
}

// batchWriter batches records for database insertion.
func (c *Collector) batchWriter(ctx context.Context) {
	defer c.wg.Done()

	batch := make([]*domain.NetFlowRecord, 0, c.config.BatchSize)
	ticker := time.NewTicker(c.config.BatchTimeout)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if c.db != nil {
			if err := c.db.Create(&batch).Error; err != nil {
				zap.S().Error("Failed to insert NetFlow records",
					zap.Int("count", len(batch)),
					zap.Error(err))
			} else {
				zap.S().Debug("NetFlow records inserted",
					zap.Int("count", len(batch)))
			}
		}

		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case <-c.shutdown:
			flush()
			return
		case record := <-c.batchCh:
			batch = append(batch, record)
			if len(batch) >= c.config.BatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

// GetTrafficStats retrieves aggregated traffic statistics.
func (c *Collector) GetTrafficStats(
	ctx context.Context,
	tenantID int64,
	timeRange string,
) ([]*domain.TrafficSummary, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	var summaries []*domain.TrafficSummary

	// Query based on time range
	var timeFilter string
	switch timeRange {
	case "hour":
		timeFilter = "created_at >= NOW() - INTERVAL '1 hour'"
	case "day":
		timeFilter = "created_at >= NOW() - INTERVAL '1 day'"
	case "week":
		timeFilter = "created_at >= NOW() - INTERVAL '1 week'"
	case "month":
		timeFilter = "created_at >= NOW() - INTERVAL '1 month'"
	default:
		timeFilter = "created_at >= NOW() - INTERVAL '1 day'"
	}

	err := c.db.Raw(`
		SELECT
			user_id,
			session_id,
			application_name,
			protocol,
			SUM(bytes) as total_bytes,
			SUM(packets) as total_packets,
			COUNT(*) as total_flows,
			EXTRACT(EPOCH FROM (MAX(last_switched) - MIN(first_switched))) as duration_sec,
			MIN(first_switched) as first_seen,
			MAX(last_switched) as last_seen
		FROM netflow_record
		WHERE tenant_id = ? AND `+timeFilter+`
		GROUP BY user_id, session_id, application_name, protocol
		ORDER BY total_bytes DESC
	`, tenantID).Scan(&summaries).Error

	return summaries, err
}
