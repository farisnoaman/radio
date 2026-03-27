#!/bin/bash

echo "=== Debugging Profile Tenant Isolation ==="
echo ""
echo "1. Checking backend logs for profile creation..."
tail -100 /home/faris/Documents/lamees/radio/backend.log | grep -i "createprofile\|listprofiles" | tail -20

echo ""
echo "2. Checking for tenant context errors..."
tail -100 /home/faris/Documents/lamees/radio/backend.log | grep -i "tenant.*context\|tenant_id" | tail -20

echo ""
echo "3. Checking for any database errors..."
tail -100 /home/faris/Documents/lamees/radio/backend.log | grep -i "error\|failed" | tail -20

echo ""
echo "=== Instructions ==="
echo "1. Restart the backend: ./stop_dev.sh && ./start_dev.sh"
echo "2. Open browser console (F12) -> Network tab"
echo "3. Create a new profile"
echo "4. Check the Network tab for POST /api/v1/radius-profiles"
echo "   - Look at the Response - what tenant_id does it have?"
echo "5. Refresh the profiles list"
echo "6. Check the Network tab for GET /api/v1/radius-profiles"
echo "   - Look at the Response - are profiles returned?"
echo "7. Share the backend log output here"
