# MikroTik Hotspot with ToughRADIUS Setup Guide

## Overview

This guide explains how to set up MikroTik router with ToughRADIUS to enable hotspot authentication using vouchers.

## Network Architecture

```
Internet
    │
    ▼
┌─────────────────┐      Ethernet       ┌─────────────┐
│  MikroTik Router│◄──────────────────►│   Laptop    │
│   192.168.1.20  │    192.168.1.0/24   │ 192.168.1.9 │
└─────────────────┘                     └─────────────┘
       │                                       │
       │ Hotspot (10.0.0.0/24)                │ RADIUS
       ▼                                       ▼ (UDP 1812/1813)
  Client Devices                           ToughRADIUS
```

## Prerequisites

- ToughRADIUS server running on laptop
- MikroTik router with RouterOS
- Network connectivity between MikroTik and laptop

## Step 1: ToughRADIUS Configuration

### 1.1 Set RADIUS Server IP

1. Log in to ToughRADIUS web interface
2. Navigate to: **Settings → System Config**
3. Set **RADIUS Server IP** = `192.168.1.9` (your laptop's ethernet IP)

### 1.2 Add MikroTik as NAS Device

1. Navigate to: **Network → NAS Devices**
2. Click **Create** to add new NAS
3. Fill in the following:

| Field | Value | Description |
|-------|-------|-------------|
| IP Address | `192.168.1.20` | MikroTik router's IP |
| Name | MikroTik Router | Descriptive name |
| Vendor | MikroTik (14988) | Router vendor code |
| Secret | `aliali6` | RADIUS shared secret |
| Status | Enabled | Enable this NAS |

4. Click **Save**

## Step 2: MikroTik Configuration

Connect to MikroTik via terminal (Winbox, SSH, or console) and run:

```bash
# Add RADIUS client configuration
/radius add address=192.168.1.9 secret=aliali6 service=hotspot,ppp authentication-port=1812 accounting-port=1813

# Verify configuration
/radius print

# Enable Hotspot (if not already enabled)
/ip hotspot enable
```

### Verify RADIUS Connection

Run this command to monitor RADIUS requests:

```
/radius monitor 0
```

Expected output when users login:
```
           pending: 0
          requests: 10
           accepts: 8
           rejects: 2
          resends: 0
       bad-replies: 0
  last-request-rtt: 5ms
```

- **accepts**: Successful authentications
- **rejects**: Failed authentications
- **timeouts**: No response from RADIUS server

## Step 3: Test Login

1. Connect a device to the MikroTik hotspot network (should get IP in 10.0.0.0/24 range)
2. Open a web browser and try to access any website
3. You should be redirected to the MikroTik login page
4. Enter voucher credentials from ToughRADIUS
5. Should authenticate and get internet access

## Troubleshooting

### Issue: "RADIUS server is not responding" in MikroTik logs

**Causes:**
1. Firewall blocking UDP ports 1812/1813
2. Wrong IP address configured
3. Wrong secret configured
4. NAS not added in ToughRADIUS

**Solutions:**

1. **Check firewall on laptop:**
```bash
sudo ufw allow from 192.168.1.20 to any port 1812,1813 proto udp
```

2. **Verify NAS exists in ToughRADIUS:**
- Go to **Network → NAS Devices**
- Ensure entry exists for 192.168.1.20

3. **Check RADIUS is listening:**
```bash
ss -tulnp | grep 1812
```

4. **Check ToughRADIUS logs:**
```bash
tail -f /path/to/backend.log | grep -i "radius\|auth"
```

### Issue: Login fails even with correct credentials

**Check:**
1. Voucher exists and is active in ToughRADIUS
2. Voucher hasn't expired
3. Voucher has remaining data/quota

### Issue: Authentication works but no internet access

**Check:**
1. Hotspot IP pool is configured correctly
2. NAT rules are in place
3. Firewall allows traffic from hotspot network

## Configuration Summary

| Setting | Value |
|---------|-------|
| MikroTik IP | 192.168.1.20 |
| Laptop (ToughRADIUS) IP | 192.168.1.9 |
| RADIUS Auth Port | 1812 |
| RADIUS Acct Port | 1813 |
| Shared Secret | aliali6 |
| Hotspot Network | 10.0.0.0/24 |

## Quick Reference Commands

**MikroTik:**
```bash
# View RADIUS configuration
/radius print

# Monitor RADIUS in real-time
/radius monitor 0

# Remove RADIUS entry
/radius remove 0

# Re-add RADIUS configuration
/radius add address=192.168.1.9 secret=aliali6 service=hotspot,ppp authentication-port=1812 accounting-port=1813

# Check hotspot status
/ip hotspot print

# View active hotspot sessions
/ip hotspot active print
```

**Linux (Firewall):**
```bash
# Allow RADIUS ports
sudo ufw allow from 192.168.1.20 to any port 1812,1813 proto udp

# Check listening ports
ss -tulnp | grep 1812
```

## Related Documentation

- ToughRADIUS Documentation: https://toughradius.org
- MikroTik Hotspot Guide: https://wiki.mikrotik.com/wiki/Manual:IP/Hotspot
- RADIUS Protocol: RFC 2865, RFC 2866