#!/bin/bash

# Configuration
CONFIG_FILE="toughradius.prod.yml"
BINARY_NAME="toughradius"

# Colors
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}Preparing ToughRadius for Production...${NC}"

# Check/Set Go Path if needed
if ! command -v go &> /dev/null; then
    export PATH=$HOME/go/go/bin:$PATH
fi

# 1. Build Frontend
echo -e "${GREEN}Building Frontend...${NC}"
cd web
npm install
npm run build
cd ..

# 2. Build Backend
echo -e "${GREEN}Building Backend Binary...${NC}"
go build -o $BINARY_NAME main.go

# 3. Create Production Config if missing
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${GREEN}Creating default production config: $CONFIG_FILE${NC}"
    # copying dev config as base, ensuring it exists, or writing default
    cat > $CONFIG_FILE <<EOF
system:
  appid: ToughRADIUS
  location: Asia/Shanghai
  workdir: ./rundata

database:
  type: sqlite
  name: toughradius.db

radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  radsec_port: 2083

web:
  host: 0.0.0.0
  port: 1816
EOF
fi

# 4. Initialize Database
echo -e "${GREEN}Initializing Database...${NC}"
./$BINARY_NAME -initdb -c $CONFIG_FILE

# 5. Start Service
echo -e "${GREEN}Starting ToughRadius Production Service...${NC}"
echo "Use Ctrl+C to stop."
./$BINARY_NAME -c $CONFIG_FILE
