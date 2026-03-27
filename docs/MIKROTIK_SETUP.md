# Mikrotik RouterOS RADIUS Configuration Guide

This guide covers configuring Mikrotik RouterOS to work with ToughRADIUS server.

## Prerequisites

- Mikrotik RouterOS device (v6.0+ recommended)
- Network connectivity between Mikrotik and ToughRADIUS server
- ToughRADIUS server running and accessible
- Admin access to Mikrotik device

## Part 1: Configure RADIUS Client on Mikrotik

### Step 1: Access Your Mikrotik Device

```bash
# Via WinBox (GUI)
# - Download WinBox from Mikrotik website
# - Connect to your Mikrotik IP address

# Via SSH
ssh admin@<mikrotik-ip-address>

# Via API (for programmatic access)
# We'll create a test script for this
```

### Step 2: Add RADIUS Server to Mikrotik

Replace `<toughradius-ip>` with your actual ToughRADIUS server IP:

```mikrotik
# Access via CLI or WinBox Terminal
/radius
add address=<toughradius-ip> \
    secret=yourSecret123 \
    service=ppp,hotspot,dhcp \
    src-address=0.0.0.0 \
    authentication-port=1812 \
    accounting-port=1813 \
    comment=ToughRADIUS

# Example with actual IP:
/radius add address=192.168.1.100 secret=radiusSecret123 service=ppp,hotspot

# Verify configuration
/radius print
```

**★ Insight ─────────────────────────────────────**
- **Secret Key**: This is a shared secret between NAS and RADIUS server. Keep it secure!
- **Service Types**: `ppp` for VPN/PPPoE, `hotspot` for captive portal, `dhcp` for DHCP authentication
- **Ports**: Standard RADIUS ports 1812 (auth) and 1813 (accounting) - ensure firewall allows these
`─────────────────────────────────────────────────`

### Step 3: Configure PPP Secrets (for PPPoE testing)

```mikrotik
# Add a test user in Mikrotik (local authentication fallback)
/ppp secret
add name=testuser password=testpass123 profile=default-encryption \
    comment="Test user for RADIUS authentication"

# Or configure to use RADIUS authentication only
/ppp aaa
set use-radius=yes accounting=yes
```

### Step 4: Configure Hotspot (Optional - for captive portal)

```mikrotik
# Setup hotspot to use RADIUS
/ip hotspot profile
set default use-radius=yes

/ip hotspot
add name=hotspot1 interface=bridge-local address-pool=hs-pool1

# Add RADIUS to hotspot
/radius
add address=<toughradius-ip> secret=yourSecret123 service=hotspot
```

### Step 5: Enable RADIUS Debugging (for testing)

```mikrotik
# Enable RADIUS debugging
/radius
add address=<toughradius-ip> secret=yourSecret123 service=ppp,hotspot

# Monitor RADIUS requests
/tool monitor
# Look for radius packets

# Check logs
/log print
# Filter for RADIUS messages
/log print where topics~radius
```

## Part 2: Firewall Configuration

### On Mikrotik (Allow RADIUS traffic)

```mikrotik
# Allow RADIUS authentication (port 1812)
/ip firewall filter
add chain=input protocol=udp dst-port=1812 action=accept \
    comment="Allow RADIUS Auth"

# Allow RADIUS accounting (port 1813)
add chain=input protocol=udp dst-port=1813 action=accept \
    comment="Allow RADIUS Acct"

# Allow input from trusted network (your ToughRADIUS server)
add chain=input src-address=<toughradius-ip>/32 action=accept
```

### On ToughRADIUS Server

```bash
# Check if RADIUS ports are open
sudo netstat -tuln | grep -E '1812|1813'

# If using ufw (Ubuntu)
sudo ufw allow 1812/udp
sudo ufw allow 1813/udp
sudo ufw status

# If using firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-port=1812/udp
sudo firewall-cmd --permanent --add-port=1813/udp
sudo firewall-cmd --reload

# Test connectivity from Mikrotik to ToughRADIUS
# Run this on Mikrotik:
/tool ping <toughradius-ip>
```

## Part 3: Configure NAS in ToughRADIUS

### Add Mikrotik as NAS Device

You can add the Mikrotik device via ToughRADIUS web UI or API:

```bash
# Via API (replace with actual values)
curl -X POST http://localhost:1816/api/v1/nas \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{
    "ipaddr": "192.168.1.1",
    "name": "Mikrotik-PPPoE-Server",
    "secret": "yourSecret123",
    "coaport": 3799,
    "type": "Mikrotik",
    "status": "enabled"
  }'
```

**★ Insight ─────────────────────────────────────**
- **NAS Secret**: Must match exactly what's configured in Mikrotik `/radius add secret=`
- **COA Port**: 3799 is standard for CoA (Change of Authorization) - allows dynamic session control
- **Type Field**: Helps identify vendor for vendor-specific attributes (Mikrotik has custom RADIUS attributes)
`─────────────────────────────────────────────────`

## Part 4: Test Authentication Flow

### Test 1: Basic Connectivity

```bash
# From Mikrotik CLI, ping ToughRADIUS server
/ping <toughradius-ip>

# Check if ports are reachable
/tool torch
# Set filter: protocol=udp, dst-port=1812
# Then try to authenticate - you should see packets
```

### Test 2: PPPoE Authentication Test

1. **Create a user in ToughRADIUS** (via web UI or API):
```bash
curl -X POST http://localhost:1816/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "status": "enabled"
  }'
```

2. **Configure PPPoE client** (on a test machine):
```bash
# Linux
pon mikrotik-pppoe

# Windows
# Network Settings -> PPPoE -> Connect with username: testuser
```

3. **Monitor authentication on Mikrotik**:
```mikrotik
# Watch for RADIUS requests
/caps-man manager
# Or check logs
/log print follow
```

### Test 3: View RADIUS Debug on Mikrotik

```mikrotik
# Enable detailed RADIUS logging
/system logging
add topics=radius action=memory

# View RADIUS packets
/ping <toughradius-ip>

# Check active PPPoE sessions
/ppp active print
/interface pppoe-server print
```

## Part 5: Common Issues & Solutions

### Issue 1: "Shared Secret Incorrect"

**Symptoms**: Authentication fails with "Shared secret incorrect" error

**Solution**:
- Verify secret matches exactly on both sides
- Check for trailing spaces in secret
- Restart Mikrotik RADIUS client after changing secret:
```mikrotik
/radius
remove [find]
add address=<toughradius-ip> secret=yourSecret123
```

### Issue 2: "No Response from RADIUS Server"

**Symptoms**: Timeout waiting for RADIUS response

**Solutions**:
1. Check network connectivity:
```mikrotik
/ping <toughradius-ip>
```

2. Verify firewall allows UDP 1812/1813:
```bash
# On ToughRADIUS server
sudo tcpdump -i any -n udp port 1812 or port 1813
```

3. Check ToughRADIUS is listening:
```bash
sudo lsof -i :1812 -i :1813
```

### Issue 3: Authentication Works But No Accounting

**Solution**:
- Verify accounting port 1813 is open
- Check Mikrotik has accounting enabled:
```mikrotik
/ppp aaa
set accounting=yes
```

### Issue 4: VLAN or IP Assignment Not Working

**Solution**:
- Check if ToughRADIUS is returning proper attributes
- Verify Mikrotik accepts vendor-specific attributes:
```mikrotik
/radius incoming
set accept-coa=yes
```

## Part 6: Advanced Configuration

### CoA (Change of Authorization)

Enable dynamic session control:

```mikrotik
/radius incoming
set port=3799
set accept-coa=yes
```

This allows ToughRADIUS to:
- Disconnect users
- Change session parameters
- Update bandwidth limits dynamically

### Multiple RADIUS Servers (Failover)

```mikrotik
/radius
add address=192.168.1.100 secret=secret1 service=ppp
add address=192.168.1.101 secret=secret2 service=ppp
# Mikrotik will failover to secondary if primary fails
```

### Vendor-Specific Attributes

ToughRADIUS can send Mikrotik-specific attributes:

```mikrotik
# View received attributes
/radius print detail

# Common Mikrotik attributes:
# - Mikrotik-Recv-Limit (rate limit)
# - Mikrotik-Xmit-Limit (rate limit)
# - Mikrotik-Group (user group assignment)
```

## Quick Reference Commands

```bash
# Mikrotik: View RADIUS configuration
/radius print

# Mikrotik: Test connectivity to RADIUS server
/ping <toughradius-ip>

# Mikrotik: View active PPPoE sessions
/ppp active print

# Mikrotik: Enable RADIUS debugging
/system logging add topics=radius

# Linux: Test RADIUS port connectivity
nc -uzv <toughradius-ip> 1812
nc -uzv <toughradius-ip> 1813

# Linux: Capture RADIUS packets
sudo tcpdump -i any -n port 1812 or port 1813 -vvv

# ToughRADIUS: Check if RADIUS server is running
ps aux | grep radius
netstat -tuln | grep -E '1812|1813'
```

## Next Steps

1. ✅ Configure Mikrotik RADIUS client
2. ✅ Add Mikrotik as NAS in ToughRADIUS
3. ✅ Create test user in ToughRADIUS
4. ✅ Test PPPoE/Hotspot authentication
5. ✅ Monitor logs and troubleshoot

See `test_mikrotik_connectivity.sh` for automated testing.
--
## Propmt:
the issues tested manually and exist: and need to be fixed:
1- During batch creation: if "voucher Batch Exipry" field is left empty, the default value"31/12/2999 12:00AM is not set as expity date.
2- if  the user enter date time in "voucher Batch Exipry" field the expiry time of vouchers is not set as user input it. it set as : Expiry time = now + allocated time from product. which is wrong. 
example: i set product of 15 hours timelimit. , i created voucher at : "3/25/2026, 9:53:26 PM" and the expiry date set to" 3/26/2026, 12:53:51 PM" although i entered "31/12/2026 12:30Am" when i created the batch . This completely wrong. not as agreed and required logic. please modify the logic to set the batch expiry as user input or as default to agreed default: 31/12/2999 12:00AM.
the field "Expiry Time" that appear in http://localhost:3000/vouchers?filter=%7B%22batch_id%22%3A%2212%22%7D&order=DESC&page=1&perPage=50&sort=id ,  is the field that propgated with wrong date time as user input or default. 
--------
please Fix this simple issue and then will test the other time qouta enforcement , will test remaining time in status page. 
------
---
correction fro your understanding:
1a: correct
1b:
SHOULD BE:
  1. Batch creation: ValidityDays = 48 [ the window to use qouta]
  2. Voucher creation: TimeQuota = 86400 (24hrs from product)
  3. Activation:
     - Calculate window end = from first login time of the voucher in the batch + 48 hours (from ValidityDays)
     - User/Voucher  expires in 48 hours
     - User has TimeQuota = 24 hours to consume within that 48hr window
  4. Result: User expires at Frist login time  + 48 hours ✅
Note: Expiry time can be set to default date normally where the voucher will be valid until first login then the time window start to count for 48 hours then the voucher will expire.

2a: 
SHOULD BE:
  1. Parse "30/12/2028 12:12pm" correctly- parse the user data correctly -i mean  the user inpute in the field "voucher expiry date" not hardcooed this example date.
  2. Batch.PrintExpireTime = 30/12/2028 12:12pm - the user inpute in the field "voucher expiry date"
  3. Voucher.ExpireTime = 30/12/2028 12:12pm -  the user inpute in the field "voucher expiry date" 
  4. User.ExpireTime = 30/12/2028 12:12pm ✅  the user inpute in the field "voucher expiry date"
  5. Result: User valid until 30/12/2028 or until quota consumed ,
2b:
SHOULD BE:
  1. Parse "30/12/2028" correctly- parse: the user inpute in the field "voucher expiry date" : the voucher will remain valid untile this user defined date 
  2. Vouchercan be activaited any time when the batch owner activates it : 
  3. Activation:
once the batch activated , it will be valid until the user defined date in "voucher expiry date" field when batch created .
     - Calculate window end = from first login time of the voucher in the batch + 48 hours (from ValidityDays)
     - User/Voucher  expires in 48 hours window even if still has allocation. if allocation finish voucher expires,
     - User has TimeQuota = 24 hours to consume within that 48hr window
  4. Result: User expires at Frist login time  + 48 hours ✅ 

-------
now , tell me how do you understand  the code should work. in all the 4 scenarios:
given that : in batch creation: 
Product : has 4.9GB data qouta and 24 hrs time qouta,
4.9GB measn the user should consume 4.9GB of data either in one sessions or multiple sessions, 
24 Hrs measn the user should consume 24 hrs online in the  network either in one session or multiple sessions
Voucher Expiry date: define up to what date this batch is valid from activation.  be default to 31-Dec-2999 01:00AM.
- advance options: Expiration Type:
1- Fix : will end as defined or set in "voucher expiry date" field 
2- First-use: we give time window for the user to use his product within this window , voucher will valid untill "Voucher expiry date"
The batch owner can activate his batch anytime. then user can log with vouchers from this batch ,
the valuse set here are for refrence example only not hardcoded user can set his own values as he want. 
--- 
Batch Creation:
- user input the amount of vouchers, length, type of voucher number letters or numbers or mixed, 
- batch expiry time : date time field.
-advance options :
1- to make / create PIN for the voucher , 
2- Expiry type: to set time window to use the voucher not the batch, this time window starts count from the first login time of the voucher not from the activation time of the batch or the creation time of the batch. it is very differnt of it.
Expiry type: two options - 
a-Fixed : it measn the vouchers in the batch will stay valid up to the batch expiry date either set by the batch owner or defaulted to default date when bacth created.
b- first-use (login) : to enter value (1- xxx) with minutes/hours/days options : if this is set for instance xx hours/days , the voucher will be valid (if no batch expired) for this specified xx hours/days from the first login : the time is counted without stoping even if the user is not active from first login until the specified window xxhours/days get 0, then the voucher will expire even if the user still have data qouta or time qouta in this voucher , the allocated window finish --> voucher expire.


## Scenario 1a: Fixed, NO Voucher expiry date input 

Input:
  - Voucher expiry field: EMPTY
  - Type: fixed
  - Product: 4.9GB, 24 hours

Code behavior:
  1. Batch creation: Sets default 2999-12-31
  2. Voucher.ExpireTime = 2999-12-31
  3. User.ExpireTime = voucher.ExpireTime = 2999-12-31
  4. User  gets TimeQuota = 24 hours (86400 seconds) from first login time

Result:
 - User must activate within 2999-12-31 (forever)
  - User valid until 2999-12-31 OR until 4.9GB/24hrs online time in the network consumed ✅

## Scenario 1b: First-use: 48 hours window , NO Voucher expiry date  input 

Input:
  - Voucher expiry field: EMPTY
  - expiration Type: first_use
  - ValidityDays: 48 (HOURS - window duration)
  - Product: 4.9GB, 24 hours

Code should behave:
  1. Voucher.ExpireTime = 2999-12-31 (default placeholder)
  2. Voucher.TimeQuota = 86400 (24hrs from product)
  3. Activation at first_login_time:
     - User.ExpireTime = first_login_time + 48 hours or set default if this voucher is not signed in yet
     - User.TimeQuota = 86400 (24hrs to consume within 48hr window)
  4. Voucher status: Shows ExpireTime = 2999-12-31 (placeholder) 

Result:
  - User must activate within 2999-12-31 (forever)
  - Once activated, user expires in 48 hours from first login time  ✅
  - Even if 24hrs quota remaining, voucher expires after 48hr window

## Scenario 2a: Fixed, WITH expiry (user input date for Voucher Expiry date ) ✅

Input:
  - Voucher expiry field: USER INPUT (e.g., "30/12/2028 12:12pm")
  - expiration Type: fixed
  - Product: 4.9GB, 24 hours

Code behavior:
  1. Parse user input date  (DD/MM/YYYY support)
  2. Batch.PrintExpireTime = parsed_date
  3. Voucher.ExpireTime = parsed_date
  4. Activation: User.ExpireTime = Voucher.ExpireTime = parsed_date 
  5. User also gets TimeQuota = 24 hours from first login time

Result:
  - Voucher must be activated BY 30/12/2028 -user inpute date 
  - User valid until 2999-12-31 OR until 4.9GB/24hrs online time in the network consumed ✅
 

## Scenario 2b: First-use, 50 hours, WITH expiry date Voucher Expiry date✅

Input:
  - Voucher expiry field: USER INPUT (e.g., "30/12/2028")
  - Type: first_use
  - ValidityDays: 50 (HOURS - window duration)
  - Product: 4.9GB, 24 hours

Code behavior:
  1. Parse user input date correctly
  2. Batch.PrintExpireTime = parsed_date (30/12/2028)
  3. Voucher.ExpireTime = parsed_date (activation deadline)
  4. Voucher.TimeQuota = 86400 (24hrs from product)
  5. Activation at first_login_time:
     - User.ExpireTime = first_login_time + 50 hours
     - User.TimeQuota = 86400 (24hrs to consume within 50hr window)

Result:
  - Voucher must be activated BY 30/12/2028 -user inpute date 
  - Once activated (first login), user expires in 50 hours from first login time ✅ 
  - Even if 24hrs quota remaining, voucher expires after 50hr  window 

-------------
it measn we have two indpendent time limits:
- product usage time : defined by priducts which is the online time in active sessions to use to consume the data qouta. either in one sessions or in many discontinuity sessions.
- validity from first use: only if set during the batch creation: is the time window or the period of time allowed to user to consume his product qouta and time and if any exhausted before the window time the voucher expires.

-------------
full one prompt:
now our main issue was not this , the issue was in vouchers , how they created, exactly how batches was created:
i want the voucher to works like this:
Given that : in batch creation: 
Product : has 4.9GB data qouta and 24 hrs time qouta,
4.9GB measn the user should consume 4.9GB of data either in one sessions or multiple sessions, 
24 Hrs measn the user should consume 24 hrs online in the  network either in one session or multiple sessions
Voucher Expiry date: define up to what date this batch is valid from activation.  be default to 31-Dec-2999 01:00AM.
- advance options: Expiration Type:
1- Fix : will end as defined or set in "voucher expiry date" field 
2- First-use: we give time window for the user to use his product within this window , voucher will valid untill "Voucher expiry date"
The batch owner can activate his batch anytime. then user can log with vouchers from this batch ,
the valuse set here are for refrence example only not hardcoded user can set his own values as he want. 
--- 
Batch Creation:
- user input the amount of vouchers, length, type of voucher number letters or numbers or mixed, 
- batch expiry time : date time field.
-advance options :
1- to make / create PIN for the voucher , 
2- Expiry type: to set time window to use the voucher not the batch, this time window starts count from the first login time of the voucher not from the activation time of the batch or the creation time of the batch. it is very differnt of it.
Expiry type: two options - 
a-Fixed : it measn the vouchers in the batch will stay valid up to the batch expiry date either set by the batch owner or defaulted to default date when bacth created.
b- first-use (login) : to enter value (1- xxx) with minutes/hours/days options : if this is set for instance xx hours/days , the voucher will be valid (if no batch expired) for this specified xx hours/days from the first login : the time is counted without stoping even if the user is not active from first login until the specified window xxhours/days get 0, then the voucher will expire even if the user still have data qouta or time qouta in this voucher , the allocated window finish --> voucher expire.


## Scenario 1a: Fixed, NO Voucher expiry date input 

Input:
  - Voucher expiry field: EMPTY
  - Type: fixed
  - Product: 4.9GB, 24 hours

Code behavior:
  1. Batch creation: Sets default 2999-12-31
  2. Voucher.ExpireTime = 2999-12-31
  3. User.ExpireTime = voucher.ExpireTime = 2999-12-31
  4. User  gets TimeQuota = 24 hours (86400 seconds) from first login time

Result:
 - User must activate within 2999-12-31 (forever)
  - User valid until 2999-12-31 OR until 4.9GB/24hrs online time in the network consumed ✅
 - the time that should be shown in user status page: 24 hours time qouta , 4.9 GB data qouta ,Time qouta activated /start conunting from first_login_time. and only count the active user sessions

## Scenario 1b: First-use: 48 hours window , NO Voucher expiry date  input 

Input:
  - Voucher expiry field: EMPTY
  - expiration Type: first_use
  - ValidityDays: 48 (HOURS - window duration)
  - Product: 4.9GB, 24 hours

Code should behave:
  1. Voucher.ExpireTime = 2999-12-31 (default placeholder)
  2. Voucher.TimeQuota = 86400 (24hrs from product)
  3. Activation at first_login_time:
     - User.ExpireTime = first_login_time + 48 hours or set default if this voucher is not signed in yet
     - User.TimeQuota = 86400 (24hrs to consume within 48hr window)
  4. Voucher status: Shows ExpireTime = 2999-12-31 (placeholder) 

Result:
  - User must activate within 2999-12-31 (forever)
  - Once activated, user expires in 48 hours from first login time  ✅
  - Even if 24hrs quota remaining, voucher expires after 48hr window
   - the time that should be shown in user status page: 24 hours time qouta , 4.9 GB data qouta ,Time qouta activated /start conunting from first_login_time. and only count the active user sessions
## Scenario 2a: Fixed, WITH expiry (user input date for Voucher Expiry date ) ✅

Input:
  - Voucher expiry field: USER INPUT (e.g., "30/12/2028 12:12pm")
  - expiration Type: fixed
  - Product: 4.9GB, 24 hours

Code behavior:
  1. Parse user input date  (DD/MM/YYYY support)
  2. Batch.PrintExpireTime = parsed_date
  3. Voucher.ExpireTime = parsed_date
  4. Activation: User.ExpireTime = Voucher.ExpireTime = parsed_date 
  5. User also gets TimeQuota = 24 hours from first login time
  

Result:
  - Voucher must be activated BY 30/12/2028 -user inpute date 
  - User valid until 2999-12-31 OR until 4.9GB/24hrs online time in the network consumed ✅
  - the time that should be shown in user status page: 24 hours time qouta , 4.9 GB data qouta ,Time qouta activated /start conunting from first_login_time. and only count the active user sessions

## Scenario 2b: First-use, 50 hours, WITH expiry date Voucher Expiry date✅

Input:
  - Voucher expiry field: USER INPUT (e.g., "30/12/2028")
  - Type: first_use
  - ValidityDays: 50 (HOURS - window duration)
  - Product: 4.9GB, 24 hours

Code behavior:
  1. Parse user input date correctly
  2. Batch.PrintExpireTime = parsed_date (30/12/2028)
  3. Voucher.ExpireTime = parsed_date (activation deadline)
  4. Voucher.TimeQuota = 86400 (24hrs from product)
  5. Time activated /start conunting from first_login_time:
     - User.ExpireTime = first_login_time + 50 hours
     - User.TimeQuota = 86400 (24hrs to consume within 50hr window)

Result:
  - Voucher must be activated BY 30/12/2028 -user inpute date 
  - Once activated (first login), user expires in 50 hours from first login time ✅ 
  - Even if 24hrs quota remaining, voucher expires after 50hr  window 
 - the time that should be shown in user status page: 24 hours time qouta , 4.9 GB data qouta ,Time qouta activated /start conunting from first_login_time. and only count the active user sessions
-------------
it measn we have two indpendent time limits:
- product usage time : defined by priducts which is the online time in active sessions to use to consume the data qouta. either in one sessions or in many discontinuity sessions.
- validity from first use: only if set during the batch creation: is the time window or the period of time allowed to user to consume his product qouta and time and if any exhausted before the window time the voucher expires.
check if current implementaion satisfy this : strictly apply this or not, 
i gave you earlier detailed example of how time should be computed, check your history, 
How time qouta calculate: Lets assume the user not consumed all his data qouta and 
lets says the user has 24 hours time qouta from the product, and  5 days time window:
-First day user loged in for 5 active minutes, then inactive, then logged in for 5 active hours, then inactive , then 45 minutes active then in active, so
time qouta remaining = 24 hours - 5 minutes - 5hrs - 45 minutes = 18 hours 10 minutes
-second day: user logged for 2 hours active then inactive, 
time qouta remaining =  18 hours 10 minutes - 2 hours = 16 hours 10 minutes

third day: user logged in for 2 hours active then inactive, 
time qouta remaining =  16 hours 10 minutes - 2 hours = 14 hours 10 minutes
forth day: user logged in for 12 hours active then inactive, 
time qouta remaining =  14 hours 10 minutes - 12 hours = 2 hours 10 minutes
fith day: user logged in for 1 hour active then in active, upto the end of the day.
time qouta remaining  = 2 hours 10 minutes - 1 hour = 1 hour 10 minutes. not used in this period windo of 5 days , so this voucher should expire at the end of the 5th day.
and should not be able to continue even if his time qouta still there and data qouta still have some MBs.
---
