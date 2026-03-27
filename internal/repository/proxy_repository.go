package repository

import (
	"context"
	"errors"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"gorm.io/gorm"
)

// ProxyRepository handles database operations for RADIUS proxy configuration.
type ProxyRepository struct {
	db *gorm.DB
}

// NewProxyRepository creates a new proxy repository.
func NewProxyRepository(db *gorm.DB) *ProxyRepository {
	return &ProxyRepository{db: db}
}

// CreateServer creates a new proxy server with tenant isolation.
func (r *ProxyRepository) CreateServer(ctx context.Context, server *domain.RadiusProxyServer) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	server.TenantID = tenantID

	return r.db.Create(server).Error
}

// GetServerByID retrieves a proxy server by ID.
func (r *ProxyRepository) GetServerByID(ctx context.Context, id int64) (*domain.RadiusProxyServer, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var server domain.RadiusProxyServer
	err = r.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&server).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &server, nil
}

// ListServers retrieves all proxy servers for the current tenant.
func (r *ProxyRepository) ListServers(ctx context.Context) ([]*domain.RadiusProxyServer, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var servers []*domain.RadiusProxyServer
	err = r.db.Where("tenant_id = ? AND status = ?", tenantID, "enabled").
		Order("priority ASC, name ASC").
		Find(&servers).Error

	return servers, err
}

// CreateRealm creates a new proxy realm.
func (r *ProxyRepository) CreateRealm(ctx context.Context, realm *domain.RadiusProxyRealm) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	realm.TenantID = tenantID

	return r.db.Create(realm).Error
}

// ListRealms retrieves all proxy realms for the current tenant.
func (r *ProxyRepository) ListRealms(ctx context.Context) ([]*domain.RadiusProxyRealm, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var realms []*domain.RadiusProxyRealm
	err = r.db.Where("tenant_id = ? AND status = ?", tenantID, "enabled").
		Order("fallback_order ASC").
		Find(&realms).Error

	return realms, err
}

// FindRealmForUsername finds the matching realm for a given username.
// Returns the realm if username suffix matches a configured realm.
func (r *ProxyRepository) FindRealmForUsername(ctx context.Context, username string) (*domain.RadiusProxyRealm, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Extract realm from username (e.g., "user@example.com" -> "example.com")
	realm := extractRealm(username)
	if realm == "" {
		return nil, nil // No realm suffix
	}

	var proxyRealm domain.RadiusProxyRealm
	err = r.db.Where("tenant_id = ? AND realm = ? AND status = ?", tenantID, realm, "enabled").
		Order("fallback_order ASC").
		First(&proxyRealm).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &proxyRealm, nil
}

// extractRealm extracts the realm suffix from a username.
// Example: "user@example.com" -> "example.com"
func extractRealm(username string) string {
	for i := len(username) - 1; i >= 0; i-- {
		if username[i] == '@' {
			return username[i+1:]
		}
	}
	return ""
}

// LogProxyRequest logs a proxied request for auditing.
func (r *ProxyRepository) LogProxyRequest(ctx context.Context, log *domain.ProxyRequestLog) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	log.TenantID = tenantID

	return r.db.Create(log).Error
}

// UpdateServer updates an existing proxy server.
func (r *ProxyRepository) UpdateServer(ctx context.Context, server *domain.RadiusProxyServer) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}

	return r.db.Where("id = ? AND tenant_id = ?", server.ID, tenantID).
		Updates(server).Error
}

// DeleteServer deletes a proxy server.
func (r *ProxyRepository) DeleteServer(ctx context.Context, id int64) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}

	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&domain.RadiusProxyServer{}).Error
}
