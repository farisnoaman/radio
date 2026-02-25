// Package acs implements the TR-069 CPE WAN Management Protocol (CWMP) Auto Configuration Server.
//
// TR-069 (Technical Report 069) is a technical specification from the Broadband Forum that
// defines the CPE WAN Management Protocol (CWMP). It enables remote management of customer
// premises equipment (CPE) such as routers, modems, and gateways.
//
// Key Features:
//   - Parse and generate TR-069 SOAP/XML messages
//   - Auto-provisioning of CPE devices with PPPoE credentials
//   - Integration with RADIUS user management
//
// Protocol Details:
//   - Transport: HTTP/HTTPS with SOAP/XML messages
//   - Default Port: TCP 7547 (ACS server)
//   - Authentication: HTTP Basic or Digest authentication
//
// Supported CWMP Methods:
//   - Inform: CPE status update (periodic or event-driven)
//   - InformResponse: Acknowledge Inform
//   - GetParameterValues: Read parameter values from CPE
//   - SetParameterValues: Write parameter values to CPE
//   - GetParameterNames: Discover available parameters
//   - Reboot: Reboot the device
//
// References:
//   - TR-069 Amendment 6: https://www.broadband-forum.org/technical/download/TR-069_Amendment-6.pdf
//   - TR-181 Data Model: https://www.broadband-forum.org/technical/download/TR-181_Issue-2_Amendment-13.pdf
package acs

import (
	"encoding/xml"
	"time"
)

// CWMP XML namespace constants used in SOAP messages.
// These namespaces are defined by the Broadband Forum for TR-069.
const (
	// NamespaceSOAP is the SOAP envelope namespace.
	NamespaceSOAP = "http://schemas.xmlsoap.org/soap/envelope/"
	// NamespaceCWMP is the CWMP protocol namespace (version 1.0).
	NamespaceCWMP = "urn:dslforum-org:cwmp-1-0"
	// NamespaceXSD is the XML Schema namespace for data types.
	NamespaceXSD = "http://www.w3.org/2001/XMLSchema-instance"
)

// CWMPMethod represents the type of CWMP method being processed.
type CWMPMethod string

// CWMP method constants for identifying message types.
const (
	MethodInform              CWMPMethod = "Inform"
	MethodInformResponse      CWMPMethod = "InformResponse"
	MethodGetParameterValues  CWMPMethod = "GetParameterValues"
	MethodSetParameterValues  CWMPMethod = "SetParameterValues"
	MethodGetParameterNames   CWMPMethod = "GetParameterNames"
	MethodReboot              CWMPMethod = "Reboot"
	MethodFactoryReset        CWMPMethod = "FactoryReset"
	MethodTransferComplete    CWMPMethod = "TransferComplete"
	MethodScheduleDownload    CWMPMethod = "ScheduleDownload"
	MethodDownload            CWMPMethod = "Download"
	MethodGetRPCMethods       CWMPMethod = "GetRPCMethods"
	MethodGetRPCMethodsResponse CWMPMethod = "GetRPCMethodsResponse"
)

// DeviceStatus represents the current status of a CPE device.
type DeviceStatus string

// Device status constants for tracking CPE state.
const (
	DeviceStatusPending      DeviceStatus = "pending"      // Device discovered, not yet provisioned
	DeviceStatusProvisioning DeviceStatus = "provisioning" // Provisioning in progress
	DeviceStatusProvisioned  DeviceStatus = "provisioned"  // Successfully provisioned
	DeviceStatusFailed       DeviceStatus = "failed"       // Provisioning failed
	DeviceStatusDisabled     DeviceStatus = "disabled"     // Device disabled by admin
)

// EventCode represents TR-069 event codes that trigger Inform messages.
// These codes indicate why the CPE is sending an Inform.
type EventCode string

// Event code constants as defined in TR-069 specification.
const (
	EventBoot                 EventCode = "0 BOOT"                  // Device boot
	EventBootStrap            EventCode = "1 BOOTSTRAP"             // First boot after factory reset
	EventPeriodic             EventCode = "2 PERIODIC"              // Periodic inform
	EventScheduled            EventCode = "3 SCHEDULED"             // Scheduled inform
	EventValueChange          EventCode = "4 VALUE CHANGE"          // Parameter value changed
	EventKicked               EventCode = "5 KICKED"                // Kicked by ACS
	EventConnectionRequest    EventCode = "6 CONNECTION REQUEST"    // ACS connection request
	EventTransferComplete     EventCode = "7 TRANSFER COMPLETE"     // File transfer complete
	EventDiagnosticsComplete  EventCode = "8 DIAGNOSTICS COMPLETE" // Diagnostics complete
	EventRequestDownload      EventCode = "9 REQUEST DOWNLOAD"     // Download requested
	EventAutonomousTransferComplete EventCode = "10 AUTONOMOUS TRANSFER COMPLETE"
	EventMReboot              EventCode = "M Reboot"               // Reboot command
	EventMDownload            EventCode = "M Download"             // Download command
	EventMScheduleInform      EventCode = "M ScheduleInform"       // Schedule inform command
)

// SOAPEnvelope represents the outer SOAP envelope structure.
// All CWMP messages are wrapped in a SOAP envelope.
//
// XML Structure:
//
//	<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
//	               xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
//	  <soap:Header>
//	    <!-- CWMP headers -->
//	  </soap:Header>
//	  <soap:Body>
//	    <!-- CWMP method -->
//	  </soap:Body>
//	</soap:Envelope>
type SOAPEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Header  *SOAPHeader `xml:"http://schemas.xmlsoap.org/soap/envelope/ Header,omitempty"`
	Body    SOAPBody    `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

// SOAPHeader contains CWMP-specific headers.
// The ID header is used to correlate requests and responses.
type SOAPHeader struct {
	ID     string `xml:"ID,omitempty"`     // Unique message ID for correlation
	NoResp int    `xml:"NoResp,omitempty"` // 1 = no response expected
}

// SOAPBody contains the CWMP method payload.
type SOAPBody struct {
	Fault                       *Fault                       `xml:"Fault,omitempty"`
	Inform                      *Inform                      `xml:"Inform,omitempty"`
	InformResponse              *InformResponse              `xml:"InformResponse,omitempty"`
	GetParameterValues          *GetParameterValues          `xml:"GetParameterValues,omitempty"`
	GetParameterValuesResponse  *GetParameterValuesResponse  `xml:"GetParameterValuesResponse,omitempty"`
	SetParameterValues          *SetParameterValues          `xml:"SetParameterValues,omitempty"`
	SetParameterValuesResponse  *SetParameterValuesResponse  `xml:"SetParameterValuesResponse,omitempty"`
	GetParameterNames           *GetParameterNames           `xml:"GetParameterNames,omitempty"`
	GetParameterNamesResponse   *GetParameterNamesResponse   `xml:"GetParameterNamesResponse,omitempty"`
	Reboot                      *Reboot                      `xml:"Reboot,omitempty"`
	RebootResponse              *RebootResponse              `xml:"RebootResponse,omitempty"`
}

// Fault represents a SOAP fault error response.
// Returned when the ACS cannot process a request.
type Fault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
	Detail      struct {
		Fault struct {
			FaultCode        int    `xml:"FaultCode"`
			FaultString      string `xml:"FaultString"`
			ParameterName    string `xml:"ParameterName,omitempty"`
			FaultStruct      string `xml:"FaultStruct,omitempty"`
			SetParamValues   struct {
				FaultCode   int    `xml:"FaultCode"`
				FaultString string `xml:"FaultString"`
			} `xml:"SetParameterValuesFault,omitempty"`
		} `xml:"cwmp:Fault"`
	} `xml:"detail"`
}

// Inform is the primary message sent by CPE to ACS.
// It contains device identification, events, and current parameter values.
//
// The CPE sends Inform messages:
//   - On boot (EventBoot)
//   - Periodically (EventPeriodic)
//   - On parameter value changes (EventValueChange)
//   - When requested by ACS (EventConnectionRequest)
//
// Example XML:
//
//	<cwmp:Inform>
//	  <DeviceId>
//	    <Manufacturer>Mikrotik</Manufacturer>
//	    <OUI>00:0C:42</OUI>
//	    <ProductClass>RouterOS</ProductClass>
//	    <SerialNumber>ABC123456</SerialNumber>
//	  </DeviceId>
//	  <Event soap:arrayType="cwmp:EventStruct[1]">
//	    <EventStruct>
//	      <EventCode>2 PERIODIC</EventCode>
//	      <CommandKey></CommandKey>
//	    </EventStruct>
//	  </Event>
//	  <ParameterList soap:arrayType="cwmp:ParameterValueStruct[2]">
//	    <ParameterValueStruct>
//	      <Name>InternetGatewayDevice.DeviceSummary</Name>
//	      <Value xsi:type="xsd:string">RouterOS v7.0</Value>
//	    </ParameterValueStruct>
//	  </ParameterList>
//	</cwmp:Inform>
type Inform struct {
	DeviceId      DeviceIdStruct           `xml:"DeviceId"`
	Event         []EventStruct            `xml:"Event>EventStruct"`
	ParameterList []ParameterValueStruct   `xml:"ParameterList>ParameterValueStruct"`
	MaxEnvelopes  int                      `xml:"MaxEnvelopes,omitempty"`
	CurrentTime   time.Time                `xml:"CurrentTime,omitempty"`
	RetryCount    int                      `xml:"RetryCount,omitempty"`
}

// DeviceIdStruct uniquely identifies a CPE device.
// The combination of OUI + ProductClass + SerialNumber should be globally unique.
//
// Fields:
//   - Manufacturer: Human-readable manufacturer name (e.g., "Mikrotik", "Cisco")
//   - OUI: Organization Unique Identifier (first 3 octets of MAC, e.g., "00:0C:42")
//   - ProductClass: Product model/class (e.g., "RouterOS", "C7200")
//   - SerialNumber: Device serial number
type DeviceIdStruct struct {
	Manufacturer string `xml:"Manufacturer"`
	OUI          string `xml:"OUI"`
	ProductClass string `xml:"ProductClass"`
	SerialNumber string `xml:"SerialNumber"`
}

// EventStruct represents an event in an Inform message.
// Multiple events can be included in a single Inform.
type EventStruct struct {
	EventCode  EventCode `xml:"EventCode"`
	CommandKey string    `xml:"CommandKey,omitempty"`
}

// ParameterValueStruct represents a parameter name-value pair.
// Used in Inform, GetParameterValuesResponse, and other methods.
//
// Type attribute values (xsi:type):
//   - xsd:string - String value
//   - xsd:int - Integer value
//   - xsd:boolean - Boolean value (true/false, 1/0)
//   - xsd:dateTime - ISO 8601 datetime
//   - xsd:base64 - Base64-encoded binary
type ParameterValueStruct struct {
	Name  string `xml:"Name"`
	Value string `xml:"Value"`
	Type  string `xml:"type,attr,omitempty"`
}

// InformResponse is sent by ACS to acknowledge an Inform message.
// After receiving InformResponse, the CPE may send more requests or close the connection.
type InformResponse struct {
	MaxEnvelopes int `xml:"MaxEnvelopes"`
}

// GetParameterValues requests parameter values from the CPE.
// The CPE responds with GetParameterValuesResponse containing the values.
//
// Example:
//
//	<cwmp:GetParameterValues>
//	  <ParameterNames soap:arrayType="xsd:string[2]">
//	    <string>InternetGatewayDevice.DeviceInfo.SerialNumber</string>
//	    <string>InternetGatewayDevice.DeviceInfo.SoftwareVersion</string>
//	  </ParameterNames>
//	</cwmp:GetParameterValues>
type GetParameterValues struct {
	ParameterNames []string `xml:"ParameterNames>string"`
}

// GetParameterValuesResponse contains the requested parameter values.
type GetParameterValuesResponse struct {
	ParameterList []ParameterValueStruct `xml:"ParameterList>ParameterValueStruct"`
}

// SetParameterValues writes parameter values to the CPE.
// Used for configuration provisioning (PPPoE credentials, Wi-Fi settings, etc.).
//
// Example for PPPoE configuration:
//
//	<cwmp:SetParameterValues>
//	  <ParameterList soap:arrayType="cwmp:ParameterValueStruct[2]">
//	    <ParameterValueStruct>
//	      <Name>InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Username</Name>
//	      <Value xsi:type="xsd:string">user@example.com</Value>
//	    </ParameterValueStruct>
//	    <ParameterValueStruct>
//	      <Name>InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Password</Name>
//	      <Value xsi:type="xsd:string">secret123</Value>
//	    </ParameterValueStruct>
//	  </ParameterList>
//	</cwmp:SetParameterValues>
type SetParameterValues struct {
	ParameterList  []ParameterValueStruct `xml:"ParameterList>ParameterValueStruct"`
	ParameterKey   string                 `xml:"ParameterKey,omitempty"`
}

// SetParameterValuesResponse confirms successful parameter write.
type SetParameterValuesResponse struct {
	Status       int    `xml:"Status"`                  // 0 = changes applied, 1 = changes pending
	ParameterKey string `xml:"ParameterKey,omitempty"`  // Echo of ParameterKey from request
}

// GetParameterNames discovers available parameters on the CPE.
// Used to explore the device data model.
type GetParameterNames struct {
	ParameterPath   string `xml:"ParameterPath"`
	NextLevel       bool   `xml:"NextLevel"`
}

// GetParameterNamesResponse contains discovered parameter names.
type GetParameterNamesResponse struct {
	ParameterList []ParameterInfoStruct `xml:"ParameterList>ParameterInfoStruct"`
}

// ParameterInfoStruct describes a parameter in the data model.
type ParameterInfoStruct struct {
	Name        string `xml:"Name"`
	Writable    bool   `xml:"Writable"`
}

// Reboot instructs the CPE to reboot.
type Reboot struct {
	CommandKey string `xml:"CommandKey,omitempty"`
}

// RebootResponse confirms reboot command acceptance.
type RebootResponse struct {
	Status int `xml:"Status"`
}

// CPEDevice represents a TR-069 managed device in the database.
// This struct maps to the "cpe_device" table and stores device information,
// provisioning status, and RADIUS integration details.
//
// Database table: cpe_device
// GORM features: Auto-migration, soft delete, timestamps
//
// Lifecycle:
//  1. Created when a new CPE sends its first Inform
//  2. Auto-provisioned with PPPoE credentials
//  3. Linked to RadiusUser for authentication
//  4. Monitored for periodic Inform messages
type CPEDevice struct {
	ID           int64      `json:"id" gorm:"primaryKey"`
	SerialNumber string     `json:"serial_number" gorm:"uniqueIndex;size:64;not null"`
	OUI          string     `json:"oui" gorm:"size:17;index"`
	Manufacturer string     `json:"manufacturer" gorm:"size:128"`
	ProductClass string     `json:"product_class" gorm:"size:128"`

	// Provisioning status
	Status        DeviceStatus `json:"status" gorm:"size:20;default:'pending';index"`
	ProvisionedAt *time.Time   `json:"provisioned_at"`
	LastError     string       `json:"last_error" gorm:"size:512"`

	// RADIUS integration
	RadiusUserID  *int64  `json:"radius_user_id" gorm:"index"`
	PPPoEUsername string  `json:"pppoe_username" gorm:"size:128;index"`
	PPPoEPassword string  `json:"-" gorm:"size:128"` // Never expose in JSON

	// Connection info
	LastInform    *time.Time `json:"last_inform"`
	LastIP        string     `json:"last_ip" gorm:"size:45"`
	ConnectionURL string     `json:"connection_url" gorm:"size:256"` // For ACS-initiated connections
	SoftwareVersion string   `json:"software_version" gorm:"size:64"`

	// Configuration
	AutoProvision bool   `json:"auto_provision" gorm:"default:true"`
	ProfileID     *int64 `json:"profile_id" gorm:"index"` // RADIUS profile for auto-provisioning

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the database table name for CPEDevice.
func (CPEDevice) TableName() string {
	return "cpe_device"
}

// IsProvisioned returns true if the device has been successfully provisioned.
func (d *CPEDevice) IsProvisioned() bool {
	return d.Status == DeviceStatusProvisioned
}

// NeedsProvisioning returns true if the device needs to be provisioned.
func (d *CPEDevice) NeedsProvisioning() bool {
	return d.Status == DeviceStatusPending && d.AutoProvision
}

// ProvisioningConfig contains configuration for the provisioning engine.
type ProvisioningConfig struct {
	// AutoProvision enables automatic provisioning of new devices
	AutoProvision bool `json:"auto_provision" yaml:"auto_provision"`

	// DefaultProfileID is the RADIUS profile assigned to auto-provisioned users
	DefaultProfileID int64 `json:"default_profile_id" yaml:"default_profile_id"`

	// UsernamePrefix is prepended to generated PPPoE usernames
	UsernamePrefix string `json:"username_prefix" yaml:"username_prefix"`

	// PasswordLength is the length of generated passwords
	PasswordLength int `json:"password_length" yaml:"password_length"`

	// DefaultNodeID is the node ID for created RADIUS users
	DefaultNodeID int64 `json:"default_node_id" yaml:"default_node_id"`

	// PPPoE parameter paths for different vendors
	PPPoEUsernamePath string `json:"pppoe_username_path" yaml:"pppoe_username_path"`
	PPPoEPasswordPath string `json:"pppoe_password_path" yaml:"pppoe_password_path"`
}

// DefaultProvisioningConfig returns a default provisioning configuration.
func DefaultProvisioningConfig() ProvisioningConfig {
	return ProvisioningConfig{
		AutoProvision:      true,
		UsernamePrefix:     "cpe_",
		PasswordLength:     12,
		PPPoEUsernamePath:  "InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Username",
		PPPoEPasswordPath:  "InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Password",
	}
}

// ACSConfig contains configuration for the ACS server.
type ACSConfig struct {
	// Listen address (default: ":7547")
	Listen string `json:"listen" yaml:"listen"`

	// HTTP Basic authentication credentials
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`

	// Connection request settings
	ConnectionRequestUsername string `json:"connection_request_username" yaml:"connection_request_username"`
	ConnectionRequestPassword string `json:"connection_request_password" yaml:"connection_request_password"`

	// Session timeout for CWMP sessions
	SessionTimeout time.Duration `json:"session_timeout" yaml:"session_timeout"`

	// Provisioning configuration
	Provisioning ProvisioningConfig `json:"provisioning" yaml:"provisioning"`
}

// DefaultACSConfig returns a default ACS configuration.
func DefaultACSConfig() ACSConfig {
	return ACSConfig{
		Listen:         ":7547",
		SessionTimeout: 30 * time.Second,
		Provisioning:   DefaultProvisioningConfig(),
	}
}
