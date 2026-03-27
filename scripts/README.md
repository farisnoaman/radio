# Mikrotik Testing Scripts

Quick reference for testing Mikrotik RouterOS connectivity with ToughRADIUS.

## Quick Start

### 1. Set Your Mikrotik Credentials

```bash
export MIKROTIK_IP="192.168.1.1"
export MIKROTIK_USER="admin"
export MIKROTIK_PASSWORD="yourpassword"
export RADIUS_SECRET="yourSecret123"
```

### 2. Run Network Connectivity Test

```bash
cd /home/faris/Documents/lamees/radio
./scripts/test_mikrotik_connectivity.sh
```

This tests:
- Network connectivity (ping)
- RADIUS ports (1812/1813)
- Mikrotik API port (8728)
- SSH access
- Retrieves RADIUS config from Mikrotik

### 3. Run API Test (Python)

```bash
# Install dependencies first
pip3 install routeros-api

# Run API test
./scripts/test_mikrotik_api.py
```

This tests:
- API TCP connection
- API login authentication
- Retrieves system info
- Shows RADIUS configuration

## Common Mikrotik Commands

### Enable API

```mikrotik
/ip service
set api enabled=yes port=8728
set api-ssl enabled=yes port=8729
```

### Configure RADIUS Client

```mikrotik
/radius
add address=<toughradius-ip> secret=yourSecret123 service=ppp,hotspot
```

### Enable PPPoE Server

```mikrotik
/interface pppoe-server server
add interface=bridge-local authentication=chap
```

### Monitor RADIUS Requests

```mikrotik
/tool torch
# Set filter: protocol=udp, dst-port=1812

/log print follow
# Watch for RADIUS messages
```

## Test Authentication Flow

### 1. Create Test User in ToughRADIUS

```bash
# Via API
curl -X POST http://localhost:1816/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "status": "enabled"
  }'
```

### 2. Add Mikrotik as NAS

```bash
curl -X POST http://localhost:1816/api/v1/nas \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "ipaddr": "192.168.1.1",
    "name": "Mikrotik-PPPoE",
    "secret": "yourSecret123",
    "type": "Mikrotik",
    "status": "enabled"
  }'
```

### 3. Test PPPoE Connection

On a client machine:

```bash
# Linux
pon mikrotik-pppoe

# Windows
# Configure PPPoE connection with username: testuser
```

### 4. Monitor Authentication

```mikrotik
# Watch active sessions
/ppp active print

# Check logs
/log print where topics~radius,ppp
```

## Capture RADIUS Packets

```bash
# Start packet capture
sudo tcpdump -i any -n 'port 1812 or port 1813' -vvv -s 0 -w radius.pcap

# Trigger authentication (login attempt)

# Stop capture with Ctrl+C

# Analyze
tcpdump -r radius.pcap -vvv
```

**★ Insight ─────────────────────────────────────**
- **RADIUS Protocol**: Uses UDP (not TCP) - packets may not show in simple netstat
- **Authentication vs Accounting**: Port 1812 for auth, 1813 for accounting - both needed
- **Shared Secret**: Critical security - must match EXACTLY on both NAS and server
`─────────────────────────────────────────────────`

## Troubleshooting

### Issue: "Cannot connect to Mikrotik"

```bash
# Test basic connectivity
ping $MIKROTIK_IP

# Check if API port is open
nc -zv $MIKROTIK_IP 8728

# From Mikrotik, check if API is running
/ip service print
```

### Issue: "RADIUS authentication fails"

```bash
# Check if ToughRADIUS is listening
sudo netstat -tuln | grep -E '1812|1813'

# Test RADIUS port from Mikrotik
/tool ping <toughradius-ip>

# Check RADIUS logs
tail -f /path/to/toughradius/logs/radius.log
```

### Issue: "API connection timeout"

```bash
# Verify API service is enabled on Mikrotik
/ip service print where name~api

# Enable if disabled
/ip service set api disabled=no

# Check firewall
/ip firewall filter print
```

## File Structure

```
scripts/
├── test_mikrotik_connectivity.sh  # Network & RADIUS port tests
├── test_mikrotik_api.py            # API communication tests
└── README.md                        # This file

docs/
└── MIKROTIK_SETUP.md               # Complete Mikrotik setup guide
```

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| MIKROTIK_IP | Mikrotik device IP | 192.168.1.1 |
| MIKROTIK_USER | Admin username | admin |
| MIKROTIK_PASSWORD | Admin password | yourpassword |
| RADIUS_SECRET | RADIUS shared secret | mySecret123 |
| TOUGHRADIUS_IP | ToughRADIUS server IP | 192.168.1.100 |

## Next Steps

1. ✅ Run connectivity tests
2. ✅ Configure Mikrotik RADIUS client
3. ✅ Add Mikrotik as NAS in ToughRADIUS
4. ✅ Create test user
5. ✅ Test authentication flow
6. ✅ Monitor and troubleshoot

See [docs/MIKROTIK_SETUP.md](../docs/MIKROTIK_SETUP.md) for detailed configuration guide.
