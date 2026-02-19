package acs

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"time"
)

// ParseSOAP parses a TR-069 CWMP SOAP message from an io.Reader.
// It returns a SOAPEnvelope containing the parsed message structure.
//
// The function handles the following CWMP methods:
//   - Inform: CPE status update
//   - InformResponse: ACS acknowledgment
//   - GetParameterValues: Request parameter values from CPE
//   - SetParameterValues: Write parameter values to CPE
//   - GetParameterNames: Discover available parameters
//   - Reboot: Reboot the device
//   - Fault: Error response
//
// Parameters:
//   - reader: An io.Reader containing the SOAP XML message
//
// Returns:
//   - *SOAPEnvelope: The parsed SOAP envelope
//   - error: Parsing error if the XML is malformed or invalid
//
// Example:
//
//	envelope, err := ParseSOAP(bytes.NewReader(xmlData))
//	if err != nil {
//	    return fmt.Errorf("failed to parse SOAP: %w", err)
//	}
//	if envelope.Body.Inform != nil {
//	    // Handle Inform message
//	}
func ParseSOAP(reader io.Reader) (*SOAPEnvelope, error) {
	var envelope SOAPEnvelope

	decoder := xml.NewDecoder(reader)
	decoder.Strict = false // Allow some flexibility in parsing

	if err := decoder.Decode(&envelope); err != nil {
		return nil, fmt.Errorf("failed to decode SOAP envelope: %w", err)
	}

	return &envelope, nil
}

// BuildSOAP generates a TR-069 CWMP SOAP message from a SOAPEnvelope.
// It returns the XML bytes ready to be sent over HTTP.
//
// The generated XML includes proper namespace declarations for:
//   - SOAP envelope (http://schemas.xmlsoap.org/soap/envelope/)
//   - CWMP protocol (urn:dslforum-org:cwmp-1-0)
//   - XML Schema instance (http://www.w3.org/2001/XMLSchema-instance)
//
// Parameters:
//   - envelope: The SOAPEnvelope to serialize
//
// Returns:
//   - []byte: The XML bytes
//   - error: Serialization error
//
// Example:
//
//	envelope := &SOAPEnvelope{
//	    Header: &SOAPHeader{ID: "123"},
//	    Body: SOAPBody{
//	        InformResponse: &InformResponse{MaxEnvelopes: 1},
//	    },
//	}
//	xmlBytes, err := BuildSOAP(envelope)
//	if err != nil {
//	    return err
//	}
//	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
//	w.Write(xmlBytes)
func BuildSOAP(envelope *SOAPEnvelope) ([]byte, error) {
	// Set XML name if not already set
	if envelope.XMLName.Local == "" {
		envelope.XMLName = xml.Name{
			Space: NamespaceSOAP,
			Local: "Envelope",
		}
	}

	xmlBytes, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SOAP envelope: %w", err)
	}

	// Add XML declaration
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteByte('\n')
	buf.Write(xmlBytes)

	return buf.Bytes(), nil
}

// ParseInform extracts an Inform message from a SOAP envelope.
// Returns nil if the envelope does not contain an Inform message.
//
// Parameters:
//   - envelope: The parsed SOAP envelope
//
// Returns:
//   - *Inform: The Inform message, or nil if not present
func ParseInform(envelope *SOAPEnvelope) *Inform {
	if envelope == nil {
		return nil
	}
	return envelope.Body.Inform
}

// BuildInformResponse creates a SOAP envelope containing an InformResponse.
// This is sent by the ACS to acknowledge receipt of an Inform message.
//
// Parameters:
//   - messageID: The ID from the Inform message (for correlation)
//   - maxEnvelopes: Maximum number of envelopes the ACS can accept (typically 1)
//
// Returns:
//   - []byte: The XML bytes ready to send
//   - error: Build error
//
// Example:
//
//	xmlBytes, err := BuildInformResponse("123456", 1)
//	if err != nil {
//	    return err
//	}
//	w.Write(xmlBytes)
func BuildInformResponse(messageID string, maxEnvelopes int) ([]byte, error) {
	envelope := &SOAPEnvelope{
		Header: &SOAPHeader{
			ID: messageID,
		},
		Body: SOAPBody{
			InformResponse: &InformResponse{
				MaxEnvelopes: maxEnvelopes,
			},
		},
	}
	return BuildSOAP(envelope)
}

// BuildGetParameterValues creates a SOAP envelope for GetParameterValues request.
// This requests the CPE to return the values of specified parameters.
//
// Parameters:
//   - messageID: Unique message ID for correlation
//   - parameterNames: List of parameter names to query
//
// Returns:
//   - []byte: The XML bytes ready to send
//   - error: Build error
//
// Example:
//
//	xmlBytes, err := BuildGetParameterValues("req123", []string{
//	    "Device.DeviceInfo.SerialNumber",
//	    "Device.DeviceInfo.SoftwareVersion",
//	})
func BuildGetParameterValues(messageID string, parameterNames []string) ([]byte, error) {
	envelope := &SOAPEnvelope{
		Header: &SOAPHeader{
			ID: messageID,
		},
		Body: SOAPBody{
			GetParameterValues: &GetParameterValues{
				ParameterNames: parameterNames,
			},
		},
	}
	return BuildSOAP(envelope)
}

// BuildSetParameterValues creates a SOAP envelope for SetParameterValues request.
// This instructs the CPE to set the specified parameter values.
//
// Parameters:
//   - messageID: Unique message ID for correlation
//   - parameters: List of parameter name-value pairs to set
//   - parameterKey: Optional key for tracking the configuration change
//
// Returns:
//   - []byte: The XML bytes ready to send
//   - error: Build error
//
// Example:
//
//	xmlBytes, err := BuildSetParameterValues("set123", []ParameterValueStruct{
//	    {Name: "Device.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Username",
//	     Value: "user@example.com", Type: "xsd:string"},
//	    {Name: "Device.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Password",
//	     Value: "secret123", Type: "xsd:string"},
//	}, "config_001")
func BuildSetParameterValues(messageID string, parameters []ParameterValueStruct, parameterKey string) ([]byte, error) {
	envelope := &SOAPEnvelope{
		Header: &SOAPHeader{
			ID: messageID,
		},
		Body: SOAPBody{
			SetParameterValues: &SetParameterValues{
				ParameterList: parameters,
				ParameterKey:  parameterKey,
			},
		},
	}
	return BuildSOAP(envelope)
}

// BuildReboot creates a SOAP envelope for a Reboot request.
// This instructs the CPE to reboot.
//
// Parameters:
//   - messageID: Unique message ID for correlation
//   - commandKey: Optional key for tracking the reboot command
//
// Returns:
//   - []byte: The XML bytes ready to send
//   - error: Build error
func BuildReboot(messageID string, commandKey string) ([]byte, error) {
	envelope := &SOAPEnvelope{
		Header: &SOAPHeader{
			ID: messageID,
		},
		Body: SOAPBody{
			Reboot: &Reboot{
				CommandKey: commandKey,
			},
		},
	}
	return BuildSOAP(envelope)
}

// BuildFault creates a SOAP envelope for a Fault response.
// This is sent when an error occurs processing a request.
//
// Parameters:
//   - messageID: The ID from the request message
//   - faultCode: SOAP fault code (e.g., "Client", "Server")
//   - faultString: Human-readable fault description
//   - cwmpFaultCode: CWMP-specific fault code
//   - cwmpFaultString: CWMP-specific fault description
//
// Returns:
//   - []byte: The XML bytes ready to send
//   - error: Build error
func BuildFault(messageID string, faultCode string, faultString string, cwmpFaultCode int, cwmpFaultString string) ([]byte, error) {
	envelope := &SOAPEnvelope{
		Header: &SOAPHeader{
			ID: messageID,
		},
		Body: SOAPBody{
			Fault: &Fault{
				FaultCode:   faultCode,
				FaultString: faultString,
			},
		},
	}

	// Set CWMP fault details
	envelope.Body.Fault.Detail.Fault.FaultCode = cwmpFaultCode
	envelope.Body.Fault.Detail.Fault.FaultString = cwmpFaultString

	return BuildSOAP(envelope)
}

// ExtractDeviceInfo extracts key device information from an Inform message.
// Returns a map of common device parameters.
//
// Parameters:
//   - inform: The Inform message to extract from
//
// Returns:
//   - map[string]string: Device information key-value pairs
func ExtractDeviceInfo(inform *Inform) map[string]string {
	info := make(map[string]string)

	if inform == nil {
		return info
	}

	// Extract DeviceId
	info["manufacturer"] = inform.DeviceId.Manufacturer
	info["oui"] = inform.DeviceId.OUI
	info["product_class"] = inform.DeviceId.ProductClass
	info["serial_number"] = inform.DeviceId.SerialNumber

	// Extract common parameters from ParameterList
	for _, param := range inform.ParameterList {
		switch param.Name {
		case "Device.DeviceInfo.SerialNumber", "InternetGatewayDevice.DeviceInfo.SerialNumber":
			info["serial_number"] = param.Value
		case "Device.DeviceInfo.SoftwareVersion", "InternetGatewayDevice.DeviceInfo.SoftwareVersion":
			info["software_version"] = param.Value
		case "Device.DeviceInfo.HardwareVersion", "InternetGatewayDevice.DeviceInfo.HardwareVersion":
			info["hardware_version"] = param.Value
		case "Device.DeviceInfo.ModelName", "InternetGatewayDevice.DeviceInfo.ModelName":
			info["model_name"] = param.Value
		case "Device.WANDevice.1.WANConnectionDevice.1.ExternalIPAddress",
			"InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.ExternalIPAddress":
			info["wan_ip"] = param.Value
		}
	}

	return info
}

// GetEventCodes returns a list of event codes from an Inform message.
//
// Parameters:
//   - inform: The Inform message to extract from
//
// Returns:
//   - []EventCode: List of event codes
func GetEventCodes(inform *Inform) []EventCode {
	if inform == nil {
		return nil
	}

	codes := make([]EventCode, len(inform.Event))
	for i, event := range inform.Event {
		codes[i] = event.EventCode
	}
	return codes
}

// HasEvent checks if an Inform message contains a specific event code.
//
// Parameters:
//   - inform: The Inform message to check
//   - code: The event code to look for
//
// Returns:
//   - bool: true if the event code is present
func HasEvent(inform *Inform, code EventCode) bool {
	if inform == nil {
		return false
	}

	for _, event := range inform.Event {
		if event.EventCode == code {
			return true
		}
	}
	return false
}

// IsBootEvent returns true if the Inform contains a boot event.
func IsBootEvent(inform *Inform) bool {
	return HasEvent(inform, EventBoot) || HasEvent(inform, EventBootStrap)
}

// IsPeriodicEvent returns true if the Inform is a periodic inform.
func IsPeriodicEvent(inform *Inform) bool {
	return HasEvent(inform, EventPeriodic)
}

// IsValueChangeEvent returns true if the Inform contains a value change event.
func IsValueChangeEvent(inform *Inform) bool {
	return HasEvent(inform, EventValueChange)
}

// GenerateMessageID generates a unique message ID for CWMP messages.
// The ID is used to correlate requests and responses.
//
// Returns:
//   - string: A unique message ID
func GenerateMessageID() string {
	// Use a simple counter-based ID for now
	// In production, this could use UUID or timestamp-based IDs
	return fmt.Sprintf("cwmp-%d", time.Now().UnixNano())
}

// GetParameterValue retrieves a parameter value by name from an Inform's parameter list.
//
// Parameters:
//   - inform: The Inform message to search
//   - name: The parameter name to find
//
// Returns:
//   - string: The parameter value, or empty string if not found
func GetParameterValue(inform *Inform, name string) string {
	if inform == nil {
		return ""
	}

	for _, param := range inform.ParameterList {
		if param.Name == name {
			return param.Value
		}
	}
	return ""
}

// NewSetParameterValuesForPPPoE creates parameter values for PPPoE configuration.
// This is a convenience function for the common use case of configuring PPPoE credentials.
//
// Parameters:
//   - username: PPPoE username
//   - password: PPPoE password
//   - usernamePath: Parameter path for username (optional, uses default)
//   - passwordPath: Parameter path for password (optional, uses default)
//
// Returns:
//   - []ParameterValueStruct: Parameter values for SetParameterValues
func NewSetParameterValuesForPPPoE(username, password, usernamePath, passwordPath string) []ParameterValueStruct {
	if usernamePath == "" {
		usernamePath = "InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Username"
	}
	if passwordPath == "" {
		passwordPath = "InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Password"
	}

	return []ParameterValueStruct{
		{
			Name:  usernamePath,
			Value: username,
			Type:  "xsd:string",
		},
		{
			Name:  passwordPath,
			Value: password,
			Type:  "xsd:string",
		},
	}
}

// NewSetParameterValuesForWiFi creates parameter values for Wi-Fi configuration.
// This is a convenience function for configuring Wi-Fi settings.
//
// Parameters:
//   - ssid: Wi-Fi network name
//   - password: Wi-Fi password (pre-shared key)
//   - wlanIndex: WLAN configuration index (default: 1)
//
// Returns:
//   - []ParameterValueStruct: Parameter values for SetParameterValues
func NewSetParameterValuesForWiFi(ssid, password string, wlanIndex int) []ParameterValueStruct {
	if wlanIndex < 1 {
		wlanIndex = 1
	}

	basePath := fmt.Sprintf("InternetGatewayDevice.LANDevice.1.WLANConfiguration.%d", wlanIndex)

	return []ParameterValueStruct{
		{
			Name:  fmt.Sprintf("%s.SSID", basePath),
			Value: ssid,
			Type:  "xsd:string",
		},
		{
			Name:  fmt.Sprintf("%s.PreSharedKey.1.PreSharedKey", basePath),
			Value: password,
			Type:  "xsd:string",
		},
	}
}

// CWMPFaultCode represents CWMP-specific fault codes.
type CWMPFaultCode int

const (
	FaultCodeRequestDenied              CWMPFaultCode = 9000
	FaultCodeRequestDeniedNoReason      CWMPFaultCode = 9001
	FaultCodeInternalError              CWMPFaultCode = 9002
	FaultCodeInvalidArguments           CWMPFaultCode = 9003
	FaultCodeResourcesExceeded          CWMPFaultCode = 9004
	FaultCodeInvalidParameterName       CWMPFaultCode = 9005
	FaultCodeInvalidParameterValue      CWMPFaultCode = 9006
	FaultCodeReadOnlyParameter          CWMPFaultCode = 9007
	FaultCodeNotificationRejected       CWMPFaultCode = 9008
	FaultCodeDownloadFailure            CWMPFaultCode = 9009
	FaultCodeUploadFailure              CWMPFaultCode = 9010
	FaultCodeFileTransferAuthFailure    CWMPFaultCode = 9011
	FaultCodeFileTransferUnsupported    CWMPFaultCode = 9012
	FaultCodeFileTransferFailure        CWMPFaultCode = 9013
	FaultCodeFileTransferCRCMismatch    CWMPFaultCode = 9014
	FaultCodeFileTransferFailed        CWMPFaultCode = 9015
	FaultCodeUnsupportedRPCMethod       CWMPFaultCode = 9016
)
