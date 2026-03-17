#!/bin/sh
set -e

# Default workdir if not set
WORK_DIR=${TOUGHRADIUS_SYSTEM_WORKER_DIR:-/var/toughradius}
DB_FILE=${TOUGHRADIUS_DB_NAME:-$WORK_DIR/data/toughradius.db}

echo "Starting ToughRadius Entrypoint..."
echo "Workdir: $WORK_DIR"
echo "Database: $DB_FILE"

# Ensure directories exist
mkdir -p "$WORK_DIR/data"
mkdir -p "$WORK_DIR/logs"

# Initialize database if it doesn't exist
if [ ! -f "$DB_FILE" ]; then
    echo "Database file not found. Initializing database..."
    /usr/local/bin/toughradius -initdb
else
    echo "Database file exists. Skipping initialization."
fi

# Start the application
exec /usr/local/bin/toughradius "$@"
