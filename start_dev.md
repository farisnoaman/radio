# Starting ToughRadius in Development

This guide outlines the steps to start the ToughRadius application in a development environment.

## Prerequisites

- **Go**: Version 1.24.0 or higher.
- **Node.js**: Version 18 or higher.
- **Configuration**: A `toughradius.dev.yml` file in the root directory.

## Quick Start (Script)

We have provided a script to automate the startup process.

```bash
chmod +x start_dev.sh
./start_dev.sh
```

This script will:
1.  Check for Go and Node.js. // go installed here: /home/faris/go/go/bin/go 
2.  Install frontend dependencies.
3.  Initialize the database.
4.  Start the backend server (on port 1816) and frontend dev server (on port 3000) concurrently.

## Manual Startup

If you prefer to run steps manually:

### 1. Build Frontend Assets
The backend requires frontend assets to be present in `web/dist` for embedding.

```bash
cd web
npm install
npm run build
cd ..
```

### 2. Initialize Database
Initialize the SQLite database (created in `rundata/data/toughradius.db`).

```bash
go run main.go -initdb -c toughradius.dev.yml
```

### 3. Start Backend Server
Run the Go backend.

```bash
go run main.go -c toughradius.dev.yml
```

The backend API will be available at `http://localhost:1816`.

### 4. Start Frontend Dev Server
Open a new terminal and run the Vite dev server.

```bash
cd web
npm run dev
```

The frontend will be available at `http://localhost:3000`.

## Default Credentials

- **Username**: `admin`
- **Password**: `toughradius`


## Notes

- Ensure the `toughradius.dev.yml` file is properly configured for your development environment.
- The script assumes a Unix-like environment. Adjust paths or commands if using Windows.

## Fetch Remote chanes:
```bash
git status && git remote -v
git fetch --all
git merge upstream/main
-------------------------------------
How to set mikrotik and tough radius:
Thinking: The user wants a complete, organized script/steps to configure their MikroTik router to work with ToughRADIUS. They want all the commands organized in a clear, copy-pasteable format.
Based on our conversation history, here's what we've done:
1. Set MikroTik IP to 192.168.1.20/24 on ether2
2. Set gateway to 192.168.1.1
3. Added RADIUS server at 192.168.1.10 (the ToughRADIUS server)
4. Enabled RADIUS for PPPoE
5. Configured hotspot (but it was disabled)
The user wants a complete organized script. I should provide:
1. Full MikroTik terminal script
2. Full organized steps
Let me create this comprehensive guide based on what we've done.
---
# Complete MikroTik Configuration Script for ToughRADIUS
Here's the full organized script and steps:
---
Part 1: MikroTik Terminal Commands (Copy & Paste)
Run these commands in sequence on your MikroTik:
# ============================================
# STEP 1: Configure IP Address on ether2
# ============================================
# Set IP address for MikroTik on ether2 (connected to laptop)
/ip address add address=192.168.1.20/24 interface=ether2
# ============================================
# STEP 2: Set Default Gateway
# ============================================
/ip route add dst-address=0.0.0.0/0 gateway=192.168.1.1
# ============================================
# STEP 3: Configure RADIUS Server
# ============================================
/radius add address=192.168.1.10 secret=toughradius service=ppp,hotspot authentication-port=1812 accounting-port=1813
# ============================================
# STEP 4: Enable RADIUS for PPP and Hotspot
# ============================================
/ppp aaa set use-radius=yes accounting=yes
# ============================================
# STEP 5: Enable Hotspot Server
# ============================================
/ip hotspot enable numbers=hotspot1
# OR use: /ip hotspot enable 0
---
Part 2: Verify Configuration
Run these to verify everything is working:
# Check IP addresses
/ip address print
# Check routes
/ip route print
# Check RADIUS servers
/radius print detail
# Check Hotspot status
/ip hotspot print
# Test connection to ToughRADIUS
ping 192.168.1.10
---
## Part 3: ToughRADIUS Web UI Configuration
### Step 1: Add MikroTik to ToughRADIUS
1. Open: **http://localhost:3000/admin**
2. Login with admin credentials
3. Go to **Network → NAS**
4. Click **Add Router**
5. Fill in:
   - **Name**: mikrotik-router
   - **IP Address**: 192.168.1.20
   - **Vendor Code**: mikrotik
   - **Secret**: toughradius
   - **COA Port**: 3799
### Step 2: Create Test User
1. Go to **Users** in ToughRADIUS
2. Click **Add User**
3. Fill in:
   - **Username**: testuser
   - **Password**: testpass123
### Step 3: Create Vouchers (Optional)
1. Go to **Vouchers**
2. Click **Generate**
3. Set quantity, length, validity
---
Part 4: Test the Setup
1. Connect a device to your network (WiFi or ethernet)
2. Open browser and try to access any website
3. Hotspot login page should appear
4. Login with test user credentials or voucher code
5. Check ToughRADIUS → Sessions to see active session
---
Network Diagram
┌──────────────┐                    ┌─────────────────┐
│   Laptop     │    ether2           │    MikroTik      │
│ 192.168.1.10 │◄──────────────────►│   192.168.1.20   │
│              │    (cable)          │                 │
│ ToughRADIUS  │                    │    ether5        │
│  localhost   │                    │        │         │
│   :3000     │                    │        ▼         │
└──────────────┘                    │  WiFi Router     │
                                     │  (customers)    │
                                     └─────────────────┘
                                              │
                                              ▼
                                    ┌─────────────────┐
                                    │  Customer       │
                                    │  Devices        │
                                    │  (get internet) │
                                    └─────────────────┘
---
Quick Reference
Item
MikroTik IP
Gateway
ToughRADIUS Server
RADIUS Secret
RADIUS Auth Port
RADIUS Acct Port
---
Commands Summary (One-Liner)
/ip address add address=192.168.1.20/24 interface=ether2 && /ip route add dst-address=0.0.0.0/0 gateway=192.168.1.1 && /radius add address=192.168.1.10 secret=toughradius service=ppp,hotspot authentication-port=1812 accounting-port=1813 && /ppp aaa set use-radius=yes accounting=yes
---
