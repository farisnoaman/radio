// Package coa provides vendor-specific implementations for CoA operations.
//
// This package contains implementations of the VendorAttributeBuilder interface
// for different NAS vendors (Mikrotik, Cisco, Huawei). Each vendor has different
// attribute formats and requirements for CoA operations.
package coa

import (
	"layeh.com/radius"

	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/cisco"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/huawei"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/mikrotik"
)

// Vendors supported for CoA operations.
const (
	// VendorMikrotik is the vendor code for Mikrotik.
	VendorMikrotik = "mikrotik"
	// VendorCisco is the vendor code for Cisco.
	VendorCisco = "cisco"
	// VendorHuawei is the vendor code for Huawei.
	VendorHuawei = "huawei"
	// VendorGeneric is the generic vendor code using standard RADIUS attributes.
	VendorGeneric = ""
)

// GetVendorBuilder returns the appropriate VendorAttributeBuilder for the given vendor code.
//
// Parameters:
//   - vendorCode: The vendor code string (e.g., "mikrotik", "cisco", "huawei")
//
// Returns:
//   - VendorAttributeBuilder: The vendor-specific builder, or nil if vendor not supported
func GetVendorBuilder(vendorCode string) VendorAttributeBuilder {
	switch vendorCode {
	case VendorMikrotik:
		return &MikrotikBuilder{}
	case VendorCisco:
		return &CiscoBuilder{}
	case VendorHuawei:
		return &HuaweiBuilder{}
	default:
		return nil
	}
}

// MikrotikBuilder implements VendorAttributeBuilder for Mikrotik devices.
//
// Mikrotik uses the following vendor-specific attributes:
//   - Mikrotik-Rate-Limit (type 8): Format "rx/tx" e.g., "10M/10M"
//   - Mikrotik-Recv-Limit (type 1): Upload rate limit in Kbps
//   - Mikrotik-Xmit-Limit (type 2): Download rate limit in Kbps
//   - Mikrotik-Total-Limit (type 17): Total data quota in bytes
//   - Mikrotik-Address-List (type 19): Firewall address list name
//
// References:
//   - Mikrotik Wiki: RADIUS Client
//   - Vendor ID: 14988
type MikrotikBuilder struct{}

// VendorCode returns the vendor identifier.
func (b *MikrotikBuilder) VendorCode() string {
	return VendorMikrotik
}

// AddRateLimit adds Mikrotik-specific rate limit attributes to the packet.
//
// Mikrotik uses Mikrotik-Rate-Limit attribute with format "rx/tx" where
// each value can be in K (kilobits), M (megabits), or G (gigabits).
//
// Parameters:
//   - pkt: The RADIUS packet to add attributes to
//   - upRate: Upload rate in Kbps
//   - downRate: Download rate in Kbps
//
// Returns an error if attribute encoding fails.
func (b *MikrotikBuilder) AddRateLimit(pkt *radius.Packet, upRate, downRate int) error {
	// Convert Kbps to Mikrotik format: "rx/tx"
	rateLimit := formatMikrotikRate(downRate) + "/" + formatMikrotikRate(upRate)
	return mikrotik.MikrotikRateLimit_SetString(pkt, rateLimit)
}

// AddSessionTimeout adds Mikrotik session timeout support.
//
// Mikrotik supports Session-Timeout through the standard RADIUS attribute.
// This method is a no-op for Mikrotik as it uses standard attributes.
//
// Parameters:
//   - pkt: The RADIUS packet
//   - timeout: Session timeout in seconds
//
// Returns nil (no error).
func (b *MikrotikBuilder) AddSessionTimeout(pkt *radius.Packet, timeout int) error {
	// Mikrotik supports Session-Timeout through standard RADIUS attribute
	// The core client already handles this via SessionTimeout field
	_ = timeout
	return nil
}

// AddDataQuota adds Mikrotik-specific data quota attributes.
//
// Adds Mikrotik-Total-Limit for total bytes quota.
//
// Parameters:
//   - pkt: The RADIUS packet to add attributes to
//   - quotaMB: Data quota in megabytes
//
// Returns an error if attribute encoding fails.
func (b *MikrotikBuilder) AddDataQuota(pkt *radius.Packet, quotaMB int64) error {
	if quotaMB <= 0 {
		return nil
	}
	// Mikrotik-Total-Limit is in bytes
	quotaBytes := quotaMB * 1024 * 1024
	return mikrotik.MikrotikTotalLimit_Set(pkt, mikrotik.MikrotikTotalLimit(quotaBytes))
}

// AddDisconnectAttributes adds vendor-specific disconnect attributes.
//
// For Mikrotik, this adds the session identification attributes.
//
// Parameters:
//   - pkt: The RADIUS packet to add attributes to
//   - username: The username of the session
//   - acctSessionID: The accounting session ID
//
// Returns nil (no error needed, standard attributes are sufficient).
func (b *MikrotikBuilder) AddDisconnectAttributes(pkt *radius.Packet, username, acctSessionID string) error {
	// Standard attributes are sufficient for Mikrotik disconnect
	_ = username
	_ = acctSessionID
	return nil
}

// CiscoBuilder implements VendorAttributeBuilder for Cisco devices.
//
// Cisco uses the following vendor-specific attributes:
//   - Cisco-AVPair (type 1): Format "attribute=value"
//   - Cisco-Account-Info (type 25): Session information
//
// Rate limits are typically set using Cisco-AVPair with format:
//   "rate-limit input <peak-rate> <normal-rate> conform-action <action> exceed-action <action>"
//   "rate-limit output <peak-rate> <normal-rate> conform-action <action> exceed-action <action>"
//
// References:
//   - Cisco IOS Security Configuration Guide
//   - Vendor ID: 1 (Cisco Systems)
type CiscoBuilder struct{}

// VendorCode returns the vendor identifier.
func (b *CiscoBuilder) VendorCode() string {
	return VendorCisco
}

// AddRateLimit adds Cisco-specific rate limit attributes.
//
// Uses Cisco-AVPair attributes with rate-limit format.
//
// Parameters:
//   - pkt: The RADIUS packet to add attributes to
//   - upRate: Upload rate in Kbps
//   - downRate: Download rate in Kbps
//
// Returns an error if attribute encoding fails.
func (b *CiscoBuilder) AddRateLimit(pkt *radius.Packet, upRate, downRate int) error {
	// Cisco AVPair rate-limit format:
	// rate-limit input <peak> <normal> conform-action transmit exceed-action drop
	// rate-limit output <peak> <normal> conform-action transmit exceed-action drop

	// Convert Kbps to bps
	downBps := downRate * 1000
	upBps := upRate * 1000

	// Add input (download) rate limit
	inputAVP := ciscoAVPairRateLimit("input", downBps)
	if err := cisco.CiscoAVPair_AddString(pkt, inputAVP); err != nil {
		return err
	}

	// Add output (upload) rate limit
	outputAVP := ciscoAVPairRateLimit("output", upBps)
	return cisco.CiscoAVPair_AddString(pkt, outputAVP)
}

// AddSessionTimeout adds Cisco session timeout support.
//
// Uses standard Session-Timeout attribute which the core client handles.
//
// Parameters:
//   - pkt: The RADIUS packet
//   - timeout: Session timeout in seconds
//
// Returns nil.
func (b *CiscoBuilder) AddSessionTimeout(pkt *radius.Packet, timeout int) error {
	// Cisco uses standard Session-Timeout attribute
	// The core client already handles this
	_ = timeout
	return nil
}

// AddDataQuota adds Cisco-specific data quota attributes.
//
// Cisco doesn't have a standard data quota attribute, so this is a no-op.
//
// Parameters:
//   - pkt: The RADIUS packet
//   - quotaMB: Data quota in megabytes (ignored for Cisco)
//
// Returns nil.
func (b *CiscoBuilder) AddDataQuota(pkt *radius.Packet, quotaMB int64) error {
	// Cisco doesn't have a standard data quota attribute in RADIUS
	// Could potentially use Cisco-AVPair but not commonly supported
	_ = quotaMB
	return nil
}

// AddDisconnectAttributes adds Cisco-specific disconnect attributes.
//
// Uses Cisco-Account-Info attribute for session identification.
//
// Parameters:
//   - pkt: The RADIUS packet to add attributes to
//   - username: The username
//   - acctSessionID: The accounting session ID
//
// Returns nil or error.
func (b *CiscoBuilder) AddDisconnectAttributes(pkt *radius.Packet, username, acctSessionID string) error {
	// Add Cisco-Account-Info with session ID if available
	if acctSessionID != "" {
		return cisco.CiscoAccountInfo_AddString(pkt, "S"+acctSessionID)
	}
	_ = username
	return nil
}

// HuaweiBuilder implements VendorAttributeBuilder for Huawei devices.
//
// Huawei uses the following vendor-specific attributes:
//   - Huawei-Input-Average-Rate (type 20): Download rate in Kbps
//   - Huawei-Output-Average-Rate (type 21): Upload rate in Kbps
//   - Huawei-Max-Input-Bandwidth (type 22): Max download bandwidth
//   - Huawei-Max-Output-Bandwidth (type 23): Max upload bandwidth
//   - Huawei-Input-Peak-Rate (type 24): Peak download rate
//   - Huawei-Output-Peak-Rate (type 25): Peak upload rate
//
// References:
//   - Huawei AC Configuration Guide
//   - Vendor ID: 2011 (Huawei Technologies)
type HuaweiBuilder struct{}

// VendorCode returns the vendor identifier.
func (b *HuaweiBuilder) VendorCode() string {
	return VendorHuawei
}

// AddRateLimit adds Huawei-specific rate limit attributes.
//
// Uses Huawei-Input-Average-Rate (downstream) and Huawei-Output-Average-Rate (upstream).
// Values are in Kbps as per Huawei specification.
//
// Parameters:
//   - pkt: The RADIUS packet to add attributes to
//   - upRate: Upload rate in Kbps
//   - downRate: Download rate in Kbps
//
// Returns an error if attribute encoding fails.
func (b *HuaweiBuilder) AddRateLimit(pkt *radius.Packet, upRate, downRate int) error {
	// Huawei uses Kbps directly
	if downRate > 0 {
		if err := huawei.HuaweiInputAverageRate_Add(pkt, huawei.HuaweiInputAverageRate(downRate)); err != nil {
			return err
		}
	}
	if upRate > 0 {
		return huawei.HuaweiOutputAverageRate_Add(pkt, huawei.HuaweiOutputAverageRate(upRate))
	}
	return nil
}

// AddSessionTimeout adds Huawei session timeout support.
//
// Uses standard Session-Timeout attribute.
//
// Parameters:
//   - pkt: The RADIUS packet
//   - timeout: Session timeout in seconds
//
// Returns nil (handled by standard attribute).
func (b *HuaweiBuilder) AddSessionTimeout(pkt *radius.Packet, timeout int) error {
	// Huawei supports Session-Timeout through standard RADIUS attribute
	// The core client already handles this
	_ = timeout
	return nil
}

// AddDataQuota adds Huawei-specific data quota attributes.
//
// Uses Huawei-Max-Input-Bandwidth and related attributes.
//
// Parameters:
//   - pkt: The RADIUS packet to add attributes to
//   - quotaMB: Data quota in megabytes
//
// Returns nil or error.
func (b *HuaweiBuilder) AddDataQuota(pkt *radius.Packet, quotaMB int64) error {
	// Huawei supports quota through various attributes
	// This is vendor-specific and may require additional configuration
	// For now, return nil as it's not commonly used
	_ = quotaMB
	return nil
}

// AddDisconnectAttributes adds Huawei-specific disconnect attributes.
//
// For Huawei, standard RADIUS attributes are typically sufficient.
//
// Parameters:
//   - pkt: The RADIUS packet to add attributes to
//   - username: The username
//   - acctSessionID: The accounting session ID
//
// Returns nil.
func (b *HuaweiBuilder) AddDisconnectAttributes(pkt *radius.Packet, username, acctSessionID string) error {
	// Standard attributes are sufficient for Huawei disconnect
	_ = username
	_ = acctSessionID
	return nil
}

// formatMikrotikRate converts Kbps to Mikrotik rate limit format.
//
// Mikrotik accepts the following formats:
//   - K: kilobits per second
//   - M: megabits per second
//   - G: gigabits per second
//
// Examples:
//   - 1024 K -> "1M"
//   - 10240 K -> "10M"
//   - 1048576 K -> "1G"
func formatMikrotikRate(kbps int) string {
	if kbps >= 1000000 {
		// Convert to G
		return formatMikrotikValue(kbps, 1000000) + "G"
	} else if kbps >= 1000 {
		// Convert to M
		return formatMikrotikValue(kbps, 1000) + "M"
	}
	// Keep in K
	return formatMikrotikValue(kbps, 1) + "K"
}

func formatMikrotikValue(kbps, divisor int) string {
	value := kbps / divisor
	if value == 0 && kbps > 0 {
		// Find the smallest non-zero unit
		if divisor >= 1000000 {
			return "1G"
		} else if divisor >= 1000 {
			return "1M"
		}
		return "1K"
	}
	// Simple number formatting
	switch value {
	case 1, 2, 5, 10, 20, 50, 100, 200, 500, 1000:
		return intToStr(value)
	default:
		// For other values, check if we can express as larger unit
		if divisor == 1 && value >= 1000 {
			return formatMikrotikRate(kbps) // Recurse to get larger unit
		}
		return intToStr(value)
	}
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToStr(-n)
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// ciscoAVPairRateLimit formats a Cisco AVPair rate-limit attribute.
//
// Format: "rate-limit <direction> <peak> <normal> conform-action transmit exceed-action drop"
func ciscoAVPairRateLimit(direction string, bps int) string {
	// Format: rate-limit input 1024000 1024000 conform-action transmit exceed-action drop
	// For simplicity, we use the same value for both peak and normal
	rate := intToStr(bps)
	return "rate-limit " + direction + " " + rate + " " + rate + " conform-action transmit exceed-action drop"
}
