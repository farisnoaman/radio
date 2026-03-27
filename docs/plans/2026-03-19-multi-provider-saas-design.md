# Multi-Provider SaaS Platform Design

**Date:** 2026-03-19
**Status:** Validated & Approved
**Version:** 1.0

## Overview

Transform the RADIUS management system into a scalable multi-provider SaaS platform supporting 100+ providers with 5000 users each and 1500 concurrent sessions per provider.

## Architecture Principles

### Core Design Decisions

1. **Schema Separation**: Each provider gets isolated PostgreSQL schema
   - Strong data isolation
   - Easy per-provider backups
   - Migration path to dedicated databases for large providers

2. **Shared Services**: Single RADIUS service with tenant-aware routing
   - Efficient resource utilization
   - Lower operational complexity
   - Centralized monitoring and management

3. **Admin-Moderated Registration**: Providers require platform admin approval
   - Quality control
   - Fraud prevention
   - Resource planning

4. **Resource Quotas**: Enforce limits on users, sessions, devices, storage
   - Prevent noisy neighbor problem
   - Fair resource distribution
   - Predictable performance

5. **Hybrid Billing**: Base fee + usage-based overages
   - Predictable revenue
   - Flexible for different provider sizes
   - Automated invoicing

6. **Tenant-Isolated Monitoring**: Providers see only their metrics
   - Security: No cross-tenant data leakage
   - Platform admin gets aggregated view
   - Prometheus-based metrics collection

7. **Provider-Managed Backups**: With admin override
   - Provider autonomy
   - Platform control when needed
   - Encryption at rest

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Platform Admin Portal                     │
│  (Provider management, monitoring, billing, global settings)  │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                   Multi-Tenant Middleware Layer                │
│  (Tenant context, authentication, rate limiting, quotas)       │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                   Application Services Layer                    │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐  │
│  │ RADIUS   │  User    │ Monitor  │ Backup   │ Billing  │  │
│  │ Service  │  Mgmt    │  Engine  │  Engine  │  Engine  │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                    PostgreSQL Schema Layer                     │
│  ┌──────────────┬──────────────┬──────────────┬─────────┐  │
│  │   provider_1 │   provider_2 │   provider_3 │  ...    │  │
│  │   (schema)   │   (schema)   │   (schema)   │(schemas)│  │
│  └──────────────┴──────────────┴──────────────┴─────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Key Features

### 1. Provider Registration & Onboarding
- Public signup form with company details
- Admin approval workflow
- Automated schema provisioning
- Welcome email with credentials

### 2. Resource Management
- User quotas: 100-5000 per provider
- Session limits: 1500 concurrent per provider
- Device/NAS quotas: Configurable per provider
- Storage limits: With automated cleanup

### 3. Tenant-Isolated Monitoring
- Device health: CPU, memory, uptime
- Network performance: Latency, packet loss, bandwidth
- RADIUS service health: Auth rate, errors
- User session analytics: Real-time tracking

### 4. Billing & Subscriptions
- Billing plans: Basic, Pro, Enterprise
- Hybrid pricing: Base fee + overage charges
- Automated invoice generation
- Usage tracking and reporting

### 5. Backup & Disaster Recovery
- Provider-configurable backups
- Automated scheduling
- Encryption at rest
- Admin override capability

### 6. White-Label Branding
- Custom domains: provider.yourplatform.com
- Visual customization: Logo, colors, themes
- Branded communications: Emails, SMS
- White-label UI: Provider-specific features

## Performance & Scalability

### High-Performance Patterns
- **Connection Pooling**: Reuse MikroTik API connections
- **Parallel Monitoring**: Check 50 devices concurrently
- **Rate Limiting**: Per-tenant request throttling
- **Circuit Breakers**: Prevent cascading failures
- **Graceful Degradation**: Emergency modes under load

### Database Optimization
- **Schema Separation**: Isolate provider data
- **Strategic Indexes**: Optimize common queries
- **Query Optimization**: Tenant-aware filtering
- **Connection Pooling**: Efficient database connections

## Security & Compliance

### Tenant Isolation
- Query-level enforcement: WHERE tenant_id = ?
- API-level validation: Cross-tenant request blocking
- Schema-level separation: Physical data isolation
- Audit logging: All access tracked

### Data Protection
- Encryption at rest: Backup encryption
- Secure secrets: API key generation
- Password hashing: bcrypt
- SQL injection prevention: Parameterized queries

## Implementation Phases

### Phase 1: Foundation (Weeks 1-4)
- Database schema & migration
- Multi-tenant middleware
- Provider CRUD APIs

### Phase 2: Provider Management (Weeks 5-8)
- Registration & onboarding
- Branding & customization

### Phase 3: Resource Management (Weeks 9-12)
- Quota system
- Rate limiting & circuit breakers

### Phase 4: Monitoring & Analytics (Weeks 13-16)
- Metrics collection
- Dashboard & alerts

### Phase 5: Billing & Backups (Weeks 17-20)
- Billing engine
- Backup system

### Phase 6: Testing & Launch (Weeks 21-24)
- Load testing
- Beta launch

## Success Criteria

- ✅ Support 100+ providers
- ✅ 5000 users per provider (500,000 total)
- ✅ 1500 concurrent sessions per provider
- ✅ Sub-second RADIUS authentication
- ✅ 99.9% uptime per provider
- ✅ Complete data isolation
- ✅ Automated billing & invoicing
- ✅ Real-time monitoring per tenant

## Dependencies

- PostgreSQL 16+ (schema support)
- Redis 7+ (caching & rate limiting)
- Prometheus (metrics collection)
- Existing RADIUS service (to be enhanced)

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Database performance at scale | High | Schema separation, connection pooling, read replicas |
| Cross-tenant data leakage | Critical | Multi-layer isolation enforcement, audit logging |
| Single provider affecting others | High | Quotas, rate limiting, circuit breakers |
| Backup failures | High | Automated testing, admin override, multiple storage locations |
| Payment integration issues | Medium | Multiple payment providers, manual billing fallback |

## Next Steps

1. Create detailed implementation plans for each phase
2. Set up git worktrees for isolated development
3. Begin Phase 1: Foundation implementation

---

**Approved by:** [User]
**Date:** 2026-03-19
