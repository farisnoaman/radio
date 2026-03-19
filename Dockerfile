# Multi-Stage Dockerfile for Multi-Provider RADIO
# Optimized for: 100 providers × 5000 users × 1500 concurrent per provider

# 1. Base stage to load all source code once
FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS source
COPY . /src
WORKDIR /src

# 2. Frontend builder
FROM --platform=$BUILDPLATFORM node:20-bookworm AS frontend-builder
COPY --from=source /src/web /web
WORKDIR /web
RUN npm ci
RUN npm run build && \
    echo "Frontend build completed:" && \
    ls -lah dist/

# 3. Backend builder
FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS builder
ARG TARGETOS
ARG TARGETARCH
ARG BUILD_VERSION=unknown

COPY --from=source /src /src
WORKDIR /src

# Copy built frontend from frontend-builder stage
COPY --from=frontend-builder /web/dist /src/web/dist

# Verify frontend is present
RUN test -f /src/web/dist/index.html || (echo "ERROR: Frontend not found!" && exit 1)

# Build for target platform with version info
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -ldflags " \
     -s -w \
     -extldflags '-static' \
     -X main.BuildVersion=${BUILD_VERSION} \
     -X main.BuildTime=$(date -u) \
     " -o /toughradius main.go

# 4. Final production image
FROM alpine:latest

RUN apk add --no-cache curl ca-certificates tzdata

# Create directory structure
RUN mkdir -p /var/toughradius/data \
             /var/toughradius/logs \
             /var/toughradius/backup \
             /var/toughradius/share

# Copy binary from builder
COPY --from=builder /toughradius /usr/local/bin/toughradius

# Copy RADIUS dictionaries and shared assets
COPY --from=source /src/share/ /var/toughradius/share/

# Copy entrypoint script
COPY --from=source /src/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Production Environment Defaults
ENV TOUGHRADIUS_SYSTEM_WORKER_DIR=/var/toughradius \
    TOUGHRADIUS_SYSTEM_APPID=RADIO \
    TOUGHRADIUS_LOGGER_MODE=production \
    TOUGHRADIUS_SYSTEM_DEBUG=false \
    \
    # Database (override in docker-compose) \
    TOUGHRADIUS_DB_TYPE=sqlite \
    TOUGHRADIUS_DB_NAME=/var/toughradius/data/toughradius.db \
    \
    # Logging \
    TOUGHRADIUS_LOGGER_FILE_ENABLE=true \
    TOUGHRADIUS_LOGGER_FILENAME=/var/toughradius/logs/toughradius.log \
    \
    # Multi-Tenant (enabled by default) \
    TOUGHRADIUS_MULTITENANT_ENABLED=true \
    \
    # Cache Configuration (L1 + L2) \
    TOUGHRADIUS_CACHE_L1_ENABLED=true \
    TOUGHRADIUS_CACHE_L1_MAX_ENTRIES=10000 \
    TOUGHRADIUS_CACHE_L2_ENABLED=true \
    TOUGHRADIUS_CACHE_USER_TTL=30 \
    TOUGHRADIUS_CACHE_NAS_TTL=300 \
    TOUGHRADIUS_CACHE_SESSION_TTL=5 \
    \
    # Redis (override with actual host in docker-compose) \
    TOUGHRADIUS_REDIS_HOST=redis \
    TOUGHRADIUS_REDIS_PORT=6379

# Expose ports:
# 1816 - Web/Admin API (HTTP)
# 1812 - RADIUS Authentication (UDP)
# 1813 - RADIUS Accounting (UDP)
# 2083 - RadSec (RADIUS over TLS)
EXPOSE 1816/tcp 1812/udp 1813/udp 2083/tcp

WORKDIR /var/toughradius

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:1816/ready || exit 1

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
