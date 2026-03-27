#!/bin/bash
################################################################################
# Test Mikrotik RouterOS Connectivity
#
# This script tests connectivity between ToughRADIUS and Mikrotik RouterOS
# including network reachability, RADIUS ports, and API access.
################################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration - update these values
MIKROTIK_IP="${MIKROTIK_IP:-192.168.1.1}"
MIKROTIK_USER="${MIKROTIK_USER:-admin}"
MIKROTIK_PASSWORD="${MIKROTIK_PASSWORD:-}"
RADIUS_SECRET="${RADIUS_SECRET:-yourSecret123}"
TOUGHRADIUS_IP="${TOUGHRADIUS_IP:-$(hostname -I | awk '{print $1}')}"

# Functions
print_header() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Test 1: Network Connectivity
test_network_connectivity() {
    print_header "TEST 1: Network Connectivity"

    print_info "Mikrotik IP: $MIKROTIK_IP"
    print_info "ToughRADIUS IP: $TOUGHRADIUS_IP"

    # Ping test
    if ping -c 3 -W 2 $MIKROTIK_IP &> /dev/null; then
        print_success "Mikrotik is reachable via ping"
    else
        print_error "Mikrotik is NOT reachable via ping"
        return 1
    fi

    # Check if host is up
    if nc -zv -w 2 $MIKROTIK_IP 22 2>&1 | grep -q succeeded; then
        print_success "Mikrotik SSH port (22) is open"
    else
        print_warning "Mikrotik SSH port (22) not accessible (may be disabled)"
    fi

    return 0
}

# Test 2: RADIUS Ports
test_radius_ports() {
    print_header "TEST 2: RADIUS Ports (1812/1813)"

    # Test authentication port
    if timeout 2 bash -c "echo > /dev/udp/$MIKROTIK_IP/1812" 2>/dev/null; then
        print_success "RADIUS Authentication port (1812) is reachable"
    else
        print_warning "Cannot verify RADIUS Auth port (may be UDP - requires actual packet)"
    fi

    # Test accounting port
    if timeout 2 bash -c "echo > /dev/udp/$MIKROTIK_IP/1813" 2>/dev/null; then
        print_success "RADIUS Accounting port (1813) is reachable"
    else
        print_warning "Cannot verify RADIUS Acct port (may be UDP - requires actual packet)"
    fi

    # Check local RADIUS ports (if ToughRADIUS is running locally)
    print_info "Checking local RADIUS server..."

    if netstat -tuln 2>/dev/null | grep -q ":1812 "; then
        print_success "Local RADIUS Auth server is listening on port 1812"
    else
        print_warning "Local RADIUS Auth server not found on port 1812"
    fi

    if netstat -tuln 2>/dev/null | grep -q ":1813 "; then
        print_success "Local RADIUS Acct server is listening on port 1813"
    else
        print_warning "Local RADIUS Acct server not found on port 1813"
    fi
}

# Test 3: Mikrotik API Port
test_api_port() {
    print_header "TEST 3: Mikrotik API Port (8728)"

    if nc -zv -w 2 $MIKROTIK_IP 8728 2>&1 | grep -q succeeded; then
        print_success "Mikrotik API port (8728) is open"
        return 0
    else
        print_error "Mikrotik API port (8728) is NOT accessible"
        print_info "Enable API: /ip service set api enabled=yes"
        return 1
    fi
}

# Test 4: SSH Access
test_ssh_access() {
    print_header "TEST 4: SSH Access to Mikrotik"

    if [ -z "$MIKROTIK_PASSWORD" ]; then
        print_warning "MIKROTIK_PASSWORD not set - skipping SSH test"
        print_info "Set password: export MIKROTIK_PASSWORD=yourpassword"
        return 0
    fi

    print_info "Testing SSH connection to $MIKROTIK_USER@$MIKROTIK_IP..."

    if sshpass -p "$MIKROTIK_PASSWORD" ssh -o StrictHostKeyChecking=no \
        -o ConnectTimeout=5 "$MIKROTIK_USER@$MIKROTIK_IP" \
        "/system resource print" &> /dev/null; then
        print_success "SSH access to Mikrotik works"
        return 0
    else
        print_error "SSH access failed"
        print_info "Check: Username, password, SSH service enabled"
        return 1
    fi
}

# Test 5: Mikrotik Configuration Check
test_mikrotik_config() {
    print_header "TEST 5: Mikrotik RADIUS Configuration"

    if [ -z "$MIKROTIK_PASSWORD" ]; then
        print_warning "MIKROTIK_PASSWORD not set - skipping config check"
        return 0
    fi

    print_info "Checking if RADIUS is configured on Mikrotik..."

    # Get RADIUS configuration via SSH
    RADIUS_OUTPUT=$(sshpass -p "$MIKROTIK_PASSWORD" ssh -o StrictHostKeyChecking=no \
        -o ConnectTimeout=5 "$MIKROTIK_USER@$MIKROTIK_IP" \
        "/radius print" 2>/dev/null || echo "")

    if [ -n "$RADIUS_OUTPUT" ]; then
        print_success "Retrieved RADIUS configuration"

        # Check if ToughRADIUS IP is configured
        if echo "$RADIUS_OUTPUT" | grep -q "$TOUGHRADIUS_IP"; then
            print_success "ToughRADIUS server ($TOUGHRADIUS_IP) is configured in Mikrotik"
        else
            print_warning "ToughRADIUS server NOT found in Mikrotik RADIUS config"
            print_info "Run: /radius add address=$TOUGHRADIUS_IP secret=$RADIUS_SECRET"
        fi

        # Show configuration
        echo -e "\n${BLUE}Current Mikrotik RADIUS Config:${NC}"
        echo "$RADIUS_OUTPUT"
    else
        print_error "Could not retrieve RADIUS configuration"
    fi
}

# Test 6: Packet Capture Helper
test_packet_capture() {
    print_header "TEST 6: Packet Capture Helper"

    print_info "To capture RADIUS packets for debugging:"
    echo ""
    echo "  sudo tcpdump -i any -n 'port 1812 or port 1813' -vvv -s 0"
    echo ""
    print_info "Then trigger a RADIUS request (PPP login, etc.)"
    print_info "Press Ctrl+C to stop capture"
}

# Main execution
main() {
    clear
    echo -e "${BLUE}"
    cat << "EOF"
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║   Mikrotik RouterOS Connectivity Test                         ║
║   ToughRADIUS Integration Tool                                 ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"

    print_info "Configuration:"
    echo "  MIKROTIK_IP: $MIKROTIK_IP"
    echo "  TOUGHRADIUS_IP: $TOUGHRADIUS_IP"
    echo "  MIKROTIK_USER: $MIKROTIK_USER"
    echo ""

    read -p "Press Enter to start tests or Ctrl+C to cancel..."

    # Run tests
    test_network_connectivity || true
    test_radius_ports || true
    test_api_port || true
    test_ssh_access || true
    test_mikrotik_config || true
    test_packet_capture

    # Summary
    print_header "Test Summary"
    print_success "Connectivity tests completed!"
    echo ""
    print_info "Next Steps:"
    echo "  1. Ensure Mikrotik has RADIUS configured (see docs/MIKROTIK_SETUP.md)"
    echo "  2. Add Mikrotik as NAS in ToughRADIUS web UI"
    echo "  3. Create test user in ToughRADIUS"
    echo "  4. Test PPPoE/Hotspot authentication"
    echo ""
}

# Check dependencies
check_dependencies() {
    local deps=("nc" "ping" "netstat" "sshpass")

    for dep in "${deps[@]}"; do
        if ! command -v $dep &> /dev/null; then
            print_warning "Dependency '$dep' not found. Install: sudo apt install $dep"
        fi
    done
}

# Check dependencies first
check_dependencies

# Run main function
main "$@"
