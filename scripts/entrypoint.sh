#!/bin/sh
set -e

# Default workdir if not set
WORK_DIR=${TOUGHRADIUS_SYSTEM_WORKER_DIR:-/var/toughradius}
DB_FILE=${TOUGHRADIUS_DB_NAME:-$WORK_DIR/data/toughradius.db}

echo "--------------------------------------------------"
echo "🚀 Starting RADIO (ISP Billing & Management)"
echo "--------------------------------------------------"
echo "Workdir: $WORK_DIR"
echo "Database: $DB_FILE"
echo "Logger Mode: ${TOUGHRADIUS_LOGGER_MODE:-production}"
echo "--------------------------------------------------"

# Ensure critical directories exist
mkdir -p "$WORK_DIR/data"
mkdir -p "$WORK_DIR/logs"
mkdir -p "$WORK_DIR/backup"
mkdir -p "$WORK_DIR/private"

# Note: Database migration is automatically handled by the application 
# on startup via internal/app/app.go. No need for manual -initdb.

# Start the application
# We pass through all arguments to the binary
exec /usr/local/bin/toughradius "$@"

