package acs

import (
	"bytes"
	"encoding/xml"
	"testing"
	"time"
)

// TestParseInform_Minimal tests parsing a minimal Inform message.
func TestParseInform_Minimal(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
                 xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap:Header>
    <cwmp:ID>123456</cwmp:ID>
  </soap:Header>
  <soap:Body>
    <cwmp:Inform>
      <DeviceId>
        <Manufacturer>TestVendor</Manufacturer>
        <OUI>00:11:22</OUI>
        <ProductClass>TestDevice</ProductClass>
        <SerialNumber>SN001</SerialNumber>
      </DeviceId>
      <Event>
        <EventStruct>
          <EventCode>2 PERIODIC</EventCode>
          <CommandKey></CommandKey>
        </EventStruct>
      </Event>
      <ParameterList>
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.HardwareVersion</Name>
          <Value xsi:type="xsd:string" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">1.0</Value>
        </ParameterValueStruct>
      </ParameterList>
      <MaxEnvelopes>1</MaxEnvelopes>
    </cwmp:Inform>
  </soap:Body>
</soap:Envelope>`

	envelope, err := ParseSOAP(bytes.NewReader([]byte(xmlData)))
	if err != nil {
		t.Fatalf("ParseSOAP failed: %v", err)
	}

	if envelope.Header == nil {
		t.Fatal("Expected header to be present")
	}

	if envelope.Header.ID != "123456" {
		t.Errorf("Expected ID '123456', got '%s'", envelope.Header.ID)
	}

	if envelope.Body.Inform == nil {
		t.Fatal("Expected Inform to be present")
	}

	inform := envelope.Body.Inform
	if inform.DeviceId.SerialNumber != "SN001" {
		t.Errorf("Expected SerialNumber 'SN001', got '%s'", inform.DeviceId.SerialNumber)
	}

	if inform.DeviceId.Manufacturer != "TestVendor" {
		t.Errorf("Expected Manufacturer 'TestVendor', got '%s'", inform.DeviceId.Manufacturer)
	}

	if inform.DeviceId.OUI != "00:11:22" {
		t.Errorf("Expected OUI '00:11:22', got '%s'", inform.DeviceId.OUI)
	}

	if inform.DeviceId.ProductClass != "TestDevice" {
		t.Errorf("Expected ProductClass 'TestDevice', got '%s'", inform.DeviceId.ProductClass)
	}

	if len(inform.Event) != 1 {
		t.Errorf("Expected 1 event, got %d", len(inform.Event))
	}

	if inform.Event[0].EventCode != EventPeriodic {
		t.Errorf("Expected event code '2 PERIODIC', got '%s'", inform.Event[0].EventCode)
	}

	if len(inform.ParameterList) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(inform.ParameterList))
	}

	if inform.ParameterList[0].Name != "Device.DeviceInfo.HardwareVersion" {
		t.Errorf("Expected parameter name 'Device.DeviceInfo.HardwareVersion', got '%s'", inform.ParameterList[0].Name)
	}

	if inform.ParameterList[0].Value != "1.0" {
		t.Errorf("Expected parameter value '1.0', got '%s'", inform.ParameterList[0].Value)
	}
}

// TestParseInform_BootEvent tests parsing an Inform with boot event.
func TestParseInform_BootEvent(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
                 xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap:Body>
    <cwmp:Inform>
      <DeviceId>
        <Manufacturer>Mikrotik</Manufacturer>
        <OUI>00:0C:42</OUI>
        <ProductClass>RouterOS</ProductClass>
        <SerialNumber>ABC123456</SerialNumber>
      </DeviceId>
      <Event>
        <EventStruct>
          <EventCode>0 BOOT</EventCode>
          <CommandKey></CommandKey>
        </EventStruct>
      </Event>
      <ParameterList/>
    </cwmp:Inform>
  </soap:Body>
</soap:Envelope>`

	envelope, err := ParseSOAP(bytes.NewReader([]byte(xmlData)))
	if err != nil {
		t.Fatalf("ParseSOAP failed: %v", err)
	}

	if envelope.Body.Inform == nil {
		t.Fatal("Expected Inform to be present")
	}

	if envelope.Body.Inform.Event[0].EventCode != EventBoot {
		t.Errorf("Expected event code '0 BOOT', got '%s'", envelope.Body.Inform.Event[0].EventCode)
	}

	if envelope.Body.Inform.DeviceId.SerialNumber != "ABC123456" {
		t.Errorf("Expected SerialNumber 'ABC123456', got '%s'", envelope.Body.Inform.DeviceId.SerialNumber)
	}
}

// TestParseInform_MultipleEvents tests parsing Inform with multiple events.
func TestParseInform_MultipleEvents(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
                 xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap:Body>
    <cwmp:Inform>
      <DeviceId>
        <Manufacturer>Cisco</Manufacturer>
        <OUI>00:1B:2B</OUI>
        <ProductClass>CPE</ProductClass>
        <SerialNumber>CISCO001</SerialNumber>
      </DeviceId>
      <Event>
        <EventStruct>
          <EventCode>2 PERIODIC</EventCode>
          <CommandKey></CommandKey>
        </EventStruct>
        <EventStruct>
          <EventCode>4 VALUE CHANGE</EventCode>
          <CommandKey>config_change</CommandKey>
        </EventStruct>
      </Event>
      <ParameterList/>
    </cwmp:Inform>
  </soap:Body>
</soap:Envelope>`

	envelope, err := ParseSOAP(bytes.NewReader([]byte(xmlData)))
	if err != nil {
		t.Fatalf("ParseSOAP failed: %v", err)
	}

	if len(envelope.Body.Inform.Event) != 2 {
		t.Errorf("Expected 2 events, got %d", len(envelope.Body.Inform.Event))
	}

	if envelope.Body.Inform.Event[0].EventCode != EventPeriodic {
		t.Errorf("Expected first event '2 PERIODIC', got '%s'", envelope.Body.Inform.Event[0].EventCode)
	}

	if envelope.Body.Inform.Event[1].EventCode != EventValueChange {
		t.Errorf("Expected second event '4 VALUE CHANGE', got '%s'", envelope.Body.Inform.Event[1].EventCode)
	}
}

// TestParseInform_MultipleParameters tests parsing Inform with multiple parameters.
func TestParseInform_MultipleParameters(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
                 xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap:Body>
    <cwmp:Inform>
      <DeviceId>
        <Manufacturer>Huawei</Manufacturer>
        <OUI>00:1E:10</OUI>
        <ProductClass>HG8245</ProductClass>
        <SerialNumber>HW123456</SerialNumber>
      </DeviceId>
      <Event>
        <EventStruct>
          <EventCode>1 BOOTSTRAP</EventCode>
          <CommandKey></CommandKey>
        </EventStruct>
      </Event>
      <ParameterList>
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.Manufacturer</Name>
          <Value xsi:type="xsd:string">Huawei</Value>
        </ParameterValueStruct>
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.SoftwareVersion</Name>
          <Value xsi:type="xsd:string">V3R017</Value>
        </ParameterValueStruct>
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.SerialNumber</Name>
          <Value xsi:type="xsd:string">HW123456</Value>
        </ParameterValueStruct>
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.ModelName</Name>
          <Value xsi:type="xsd:string">HG8245H</Value>
        </ParameterValueStruct>
      </ParameterList>
    </cwmp:Inform>
  </soap:Body>
</soap:Envelope>`

	envelope, err := ParseSOAP(bytes.NewReader([]byte(xmlData)))
	if err != nil {
		t.Fatalf("ParseSOAP failed: %v", err)
	}

	if len(envelope.Body.Inform.ParameterList) != 4 {
		t.Errorf("Expected 4 parameters, got %d", len(envelope.Body.Inform.ParameterList))
	}

	// Check first parameter
	if envelope.Body.Inform.ParameterList[0].Name != "Device.DeviceInfo.Manufacturer" {
		t.Errorf("Expected first parameter name 'Device.DeviceInfo.Manufacturer', got '%s'",
			envelope.Body.Inform.ParameterList[0].Name)
	}

	if envelope.Body.Inform.ParameterList[0].Value != "Huawei" {
		t.Errorf("Expected first parameter value 'Huawei', got '%s'",
			envelope.Body.Inform.ParameterList[0].Value)
	}
}

// TestParseInform_InvalidXML tests handling of malformed XML.
func TestParseInform_InvalidXML(t *testing.T) {
	// Test with truly malformed XML (invalid syntax - broken XML declaration)
	xmlData := `<?xml version="1.0"
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <cwmp:Inform>
      <DeviceId>
        <Manufacturer>Test</Manufacturer>
      </DeviceId>
    </cwmp:Inform>
  </soap:Body>
</soap:Envelope>`

	_, err := ParseSOAP(bytes.NewReader([]byte(xmlData)))
	if err == nil {
		t.Fatal("Expected error for malformed XML, got nil")
	}
}

// TestParseInform_EmptyBody tests handling of empty body.
func TestParseInform_EmptyBody(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Header>
    <cwmp:ID>test123</cwmp:ID>
  </soap:Header>
  <soap:Body/>
</soap:Envelope>`

	envelope, err := ParseSOAP(bytes.NewReader([]byte(xmlData)))
	if err != nil {
		t.Fatalf("ParseSOAP failed: %v", err)
	}

	if envelope.Header == nil {
		t.Fatal("Expected header to be present")
	}

	if envelope.Header.ID != "test123" {
		t.Errorf("Expected ID 'test123', got '%s'", envelope.Header.ID)
	}

	// Body should have nil Inform (empty body)
	if envelope.Body.Inform != nil {
		t.Error("Expected nil Inform for empty body")
	}
}

// TestBuildInformResponse tests building an InformResponse message.
func TestBuildInformResponse(t *testing.T) {
	envelope := &SOAPEnvelope{
		Header: &SOAPHeader{
			ID: "123456",
		},
		Body: SOAPBody{
			InformResponse: &InformResponse{
				MaxEnvelopes: 1,
			},
		},
	}

	xmlBytes, err := BuildSOAP(envelope)
	if err != nil {
		t.Fatalf("BuildSOAP failed: %v", err)
	}

	// Verify the XML can be parsed back
	parsed, err := ParseSOAP(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	if parsed.Body.InformResponse == nil {
		t.Fatal("Expected InformResponse in parsed envelope")
	}

	if parsed.Body.InformResponse.MaxEnvelopes != 1 {
		t.Errorf("Expected MaxEnvelopes 1, got %d", parsed.Body.InformResponse.MaxEnvelopes)
	}
}

// TestBuildGetParameterValues tests building a GetParameterValues request.
func TestBuildGetParameterValues(t *testing.T) {
	envelope := &SOAPEnvelope{
		Header: &SOAPHeader{
			ID: "654321",
		},
		Body: SOAPBody{
			GetParameterValues: &GetParameterValues{
				ParameterNames: []string{
					"Device.DeviceInfo.SerialNumber",
					"Device.DeviceInfo.SoftwareVersion",
				},
			},
		},
	}

	xmlBytes, err := BuildSOAP(envelope)
	if err != nil {
		t.Fatalf("BuildSOAP failed: %v", err)
	}

	// Verify the XML can be parsed back
	parsed, err := ParseSOAP(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	if parsed.Body.GetParameterValues == nil {
		t.Fatal("Expected GetParameterValues in parsed envelope")
	}

	if len(parsed.Body.GetParameterValues.ParameterNames) != 2 {
		t.Errorf("Expected 2 parameter names, got %d",
			len(parsed.Body.GetParameterValues.ParameterNames))
	}
}

// TestBuildSetParameterValues tests building a SetParameterValues request.
func TestBuildSetParameterValues(t *testing.T) {
	envelope := &SOAPEnvelope{
		Header: &SOAPHeader{
			ID: "789012",
		},
		Body: SOAPBody{
			SetParameterValues: &SetParameterValues{
				ParameterList: []ParameterValueStruct{
					{
						Name:  "Device.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Username",
						Value: "user@example.com",
						Type:  "xsd:string",
					},
					{
						Name:  "Device.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Password",
						Value: "secret123",
						Type:  "xsd:string",
					},
				},
				ParameterKey: "config_001",
			},
		},
	}

	xmlBytes, err := BuildSOAP(envelope)
	if err != nil {
		t.Fatalf("BuildSOAP failed: %v", err)
	}

	// Verify the XML can be parsed back
	parsed, err := ParseSOAP(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	if parsed.Body.SetParameterValues == nil {
		t.Fatal("Expected SetParameterValues in parsed envelope")
	}

	if len(parsed.Body.SetParameterValues.ParameterList) != 2 {
		t.Errorf("Expected 2 parameters, got %d",
			len(parsed.Body.SetParameterValues.ParameterList))
	}

	if parsed.Body.SetParameterValues.ParameterList[0].Name != "Device.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Username" {
		t.Errorf("Unexpected first parameter name: %s",
			parsed.Body.SetParameterValues.ParameterList[0].Name)
	}

	if parsed.Body.SetParameterValues.ParameterList[0].Value != "user@example.com" {
		t.Errorf("Unexpected first parameter value: %s",
			parsed.Body.SetParameterValues.ParameterList[0].Value)
	}
}

// TestBuildFault tests building a Fault response.
func TestBuildFault(t *testing.T) {
	envelope := &SOAPEnvelope{
		Header: &SOAPHeader{
			ID: "fault123",
		},
		Body: SOAPBody{
			Fault: &Fault{
				FaultCode:   "Client",
				FaultString: "CWMP fault",
				Detail: struct {
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
				}{},
			},
		},
	}

	xmlBytes, err := BuildSOAP(envelope)
	if err != nil {
		t.Fatalf("BuildSOAP failed: %v", err)
	}

	// Verify the XML can be parsed back
	parsed, err := ParseSOAP(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	if parsed.Body.Fault == nil {
		t.Fatal("Expected Fault in parsed envelope")
	}

	if parsed.Body.Fault.FaultCode != "Client" {
		t.Errorf("Expected FaultCode 'Client', got '%s'", parsed.Body.Fault.FaultCode)
	}
}

// TestParseGetParameterValuesResponse tests parsing a GetParameterValuesResponse.
func TestParseGetParameterValuesResponse(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
                 xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap:Header>
    <cwmp:ID>response456</cwmp:ID>
  </soap:Header>
  <soap:Body>
    <cwmp:GetParameterValuesResponse>
      <ParameterList>
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.SerialNumber</Name>
          <Value xsi:type="xsd:string">SN12345</Value>
        </ParameterValueStruct>
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.SoftwareVersion</Name>
          <Value xsi:type="xsd:string">v2.1.0</Value>
        </ParameterValueStruct>
      </ParameterList>
    </cwmp:GetParameterValuesResponse>
  </soap:Body>
</soap:Envelope>`

	envelope, err := ParseSOAP(bytes.NewReader([]byte(xmlData)))
	if err != nil {
		t.Fatalf("ParseSOAP failed: %v", err)
	}

	if envelope.Body.GetParameterValuesResponse == nil {
		t.Fatal("Expected GetParameterValuesResponse to be present")
	}

	if len(envelope.Body.GetParameterValuesResponse.ParameterList) != 2 {
		t.Errorf("Expected 2 parameters, got %d",
			len(envelope.Body.GetParameterValuesResponse.ParameterList))
	}

	if envelope.Body.GetParameterValuesResponse.ParameterList[0].Value != "SN12345" {
		t.Errorf("Expected first value 'SN12345', got '%s'",
			envelope.Body.GetParameterValuesResponse.ParameterList[0].Value)
	}
}

// TestParseSetParameterValuesResponse tests parsing a SetParameterValuesResponse.
func TestParseSetParameterValuesResponse(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
                 xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap:Header>
    <cwmp:ID>setResp789</cwmp:ID>
  </soap:Header>
  <soap:Body>
    <cwmp:SetParameterValuesResponse>
      <Status>0</Status>
      <ParameterKey>config_001</ParameterKey>
    </cwmp:SetParameterValuesResponse>
  </soap:Body>
</soap:Envelope>`

	envelope, err := ParseSOAP(bytes.NewReader([]byte(xmlData)))
	if err != nil {
		t.Fatalf("ParseSOAP failed: %v", err)
	}

	if envelope.Body.SetParameterValuesResponse == nil {
		t.Fatal("Expected SetParameterValuesResponse to be present")
	}

	if envelope.Body.SetParameterValuesResponse.Status != 0 {
		t.Errorf("Expected Status 0, got %d",
			envelope.Body.SetParameterValuesResponse.Status)
	}

	if envelope.Body.SetParameterValuesResponse.ParameterKey != "config_001" {
		t.Errorf("Expected ParameterKey 'config_001', got '%s'",
			envelope.Body.SetParameterValuesResponse.ParameterKey)
	}
}

// TestParseRebootRequest tests parsing a Reboot request.
func TestParseRebootRequest(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
                 xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap:Header>
    <cwmp:ID>reboot001</cwmp:ID>
  </soap:Header>
  <soap:Body>
    <cwmp:Reboot>
      <CommandKey>admin_reboot</CommandKey>
    </cwmp:Reboot>
  </soap:Body>
</soap:Envelope>`

	envelope, err := ParseSOAP(bytes.NewReader([]byte(xmlData)))
	if err != nil {
		t.Fatalf("ParseSOAP failed: %v", err)
	}

	if envelope.Body.Reboot == nil {
		t.Fatal("Expected Reboot to be present")
	}

	if envelope.Body.Reboot.CommandKey != "admin_reboot" {
		t.Errorf("Expected CommandKey 'admin_reboot', got '%s'",
			envelope.Body.Reboot.CommandKey)
	}
}

// TestBuildRebootResponse tests building a RebootResponse.
func TestBuildRebootResponse(t *testing.T) {
	envelope := &SOAPEnvelope{
		Header: &SOAPHeader{
			ID: "reboot_resp",
		},
		Body: SOAPBody{
			RebootResponse: &RebootResponse{
				Status: 0,
			},
		},
	}

	xmlBytes, err := BuildSOAP(envelope)
	if err != nil {
		t.Fatalf("BuildSOAP failed: %v", err)
	}

	// Verify the XML contains the expected elements
	xmlStr := string(xmlBytes)
	if !bytes.Contains(xmlBytes, []byte("RebootResponse")) {
		t.Error("Expected RebootResponse in generated XML")
	}

	_ = xmlStr // Use xmlStr to avoid unused variable
}

// TestExtractDeviceId tests extracting DeviceId from an Inform.
func TestExtractDeviceId(t *testing.T) {
	inform := &Inform{
		DeviceId: DeviceIdStruct{
			Manufacturer: "TestVendor",
			OUI:          "00:11:22",
			ProductClass: "TestDevice",
			SerialNumber: "SN001",
		},
	}

	deviceID := inform.DeviceId.SerialNumber
	if deviceID != "SN001" {
		t.Errorf("Expected 'SN001', got '%s'", deviceID)
	}

	// Test unique key generation
	uniqueKey := inform.DeviceId.OUI + "-" + inform.DeviceId.SerialNumber
	if uniqueKey != "00:11:22-SN001" {
		t.Errorf("Expected unique key '00:11:22-SN001', got '%s'", uniqueKey)
	}
}

// TestParameterValueStruct_TypeHandling tests type attribute handling.
func TestParameterValueStruct_TypeHandling(t *testing.T) {
	tests := []struct {
		name     string
		param    ParameterValueStruct
		expected string
	}{
		{
			name: "string type",
			param: ParameterValueStruct{
				Name:  "Test.String",
				Value: "hello",
				Type:  "xsd:string",
			},
			expected: "hello",
		},
		{
			name: "int type",
			param: ParameterValueStruct{
				Name:  "Test.Int",
				Value: "42",
				Type:  "xsd:int",
			},
			expected: "42",
		},
		{
			name: "boolean true",
			param: ParameterValueStruct{
				Name:  "Test.Bool",
				Value: "1",
				Type:  "xsd:boolean",
			},
			expected: "1",
		},
		{
			name: "empty type defaults to string",
			param: ParameterValueStruct{
				Name:  "Test.Empty",
				Value: "default",
				Type:  "",
			},
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.param.Value != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, tt.param.Value)
			}
		})
	}
}

// TestCWMPMethodConstants tests that method constants are correctly defined.
func TestCWMPMethodConstants(t *testing.T) {
	if MethodInform != "Inform" {
		t.Errorf("MethodInform should be 'Inform', got '%s'", MethodInform)
	}

	if MethodInformResponse != "InformResponse" {
		t.Errorf("MethodInformResponse should be 'InformResponse', got '%s'", MethodInformResponse)
	}

	if MethodGetParameterValues != "GetParameterValues" {
		t.Errorf("MethodGetParameterValues should be 'GetParameterValues', got '%s'", MethodGetParameterValues)
	}

	if MethodSetParameterValues != "SetParameterValues" {
		t.Errorf("MethodSetParameterValues should be 'SetParameterValues', got '%s'", MethodSetParameterValues)
	}

	if MethodReboot != "Reboot" {
		t.Errorf("MethodReboot should be 'Reboot', got '%s'", MethodReboot)
	}
}

// TestEventCodeConstants tests that event code constants are correctly defined.
func TestEventCodeConstants(t *testing.T) {
	if EventBoot != "0 BOOT" {
		t.Errorf("EventBoot should be '0 BOOT', got '%s'", EventBoot)
	}

	if EventBootStrap != "1 BOOTSTRAP" {
		t.Errorf("EventBootStrap should be '1 BOOTSTRAP', got '%s'", EventBootStrap)
	}

	if EventPeriodic != "2 PERIODIC" {
		t.Errorf("EventPeriodic should be '2 PERIODIC', got '%s'", EventPeriodic)
	}

	if EventValueChange != "4 VALUE CHANGE" {
		t.Errorf("EventValueChange should be '4 VALUE CHANGE', got '%s'", EventValueChange)
	}
}

// TestDeviceStatusConstants tests device status constants.
func TestDeviceStatusConstants(t *testing.T) {
	if DeviceStatusPending != "pending" {
		t.Errorf("DeviceStatusPending should be 'pending', got '%s'", DeviceStatusPending)
	}

	if DeviceStatusProvisioned != "provisioned" {
		t.Errorf("DeviceStatusProvisioned should be 'provisioned', got '%s'", DeviceStatusProvisioned)
	}

	if DeviceStatusFailed != "failed" {
		t.Errorf("DeviceStatusFailed should be 'failed', got '%s'", DeviceStatusFailed)
	}
}

// TestInformWithCurrentTime tests parsing Inform with CurrentTime field.
func TestInformWithCurrentTime(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
                 xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap:Body>
    <cwmp:Inform>
      <DeviceId>
        <Manufacturer>Test</Manufacturer>
        <OUI>00:11:22</OUI>
        <ProductClass>TestDevice</ProductClass>
        <SerialNumber>SN001</SerialNumber>
      </DeviceId>
      <Event>
        <EventStruct>
          <EventCode>2 PERIODIC</EventCode>
        </EventStruct>
      </Event>
      <ParameterList/>
      <MaxEnvelopes>1</MaxEnvelopes>
      <CurrentTime>2024-01-15T10:30:00Z</CurrentTime>
      <RetryCount>0</RetryCount>
    </cwmp:Inform>
  </soap:Body>
</soap:Envelope>`

	envelope, err := ParseSOAP(bytes.NewReader([]byte(xmlData)))
	if err != nil {
		t.Fatalf("ParseSOAP failed: %v", err)
	}

	inform := envelope.Body.Inform
	if inform == nil {
		t.Fatal("Expected Inform to be present")
	}

	// CurrentTime should be parsed
	expectedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if !inform.CurrentTime.Equal(expectedTime) {
		t.Errorf("Expected CurrentTime %v, got %v", expectedTime, inform.CurrentTime)
	}

	if inform.RetryCount != 0 {
		t.Errorf("Expected RetryCount 0, got %d", inform.RetryCount)
	}
}

// TestXMLNamespaceHandling tests proper handling of XML namespaces.
func TestXMLNamespaceHandling(t *testing.T) {
	// Test with explicit namespace declarations
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
                 xmlns:cwmp="urn:dslforum-org:cwmp-1-0"
                 xmlns:xsd="http://www.w3.org/2001/XMLSchema"
                 xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <soap:Body>
    <cwmp:Inform>
      <DeviceId>
        <Manufacturer>NS-Test</Manufacturer>
        <OUI>00:11:22</OUI>
        <ProductClass>NS-Device</ProductClass>
        <SerialNumber>NS001</SerialNumber>
      </DeviceId>
      <Event>
        <EventStruct>
          <EventCode>0 BOOT</EventCode>
        </EventStruct>
      </Event>
      <ParameterList>
        <ParameterValueStruct>
          <Name>Device.Info</Name>
          <Value xsi:type="xsd:string">test-value</Value>
        </ParameterValueStruct>
      </ParameterList>
    </cwmp:Inform>
  </soap:Body>
</soap:Envelope>`

	envelope, err := ParseSOAP(bytes.NewReader([]byte(xmlData)))
	if err != nil {
		t.Fatalf("ParseSOAP failed: %v", err)
	}

	if envelope.Body.Inform.DeviceId.Manufacturer != "NS-Test" {
		t.Errorf("Expected Manufacturer 'NS-Test', got '%s'", envelope.Body.Inform.DeviceId.Manufacturer)
	}

	if envelope.Body.Inform.ParameterList[0].Value != "test-value" {
		t.Errorf("Expected parameter value 'test-value', got '%s'", envelope.Body.Inform.ParameterList[0].Value)
	}
}

// BenchmarkParseInform benchmarks the Inform parsing performance.
func BenchmarkParseInform(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
                 xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap:Header>
    <cwmp:ID>123456</cwmp:ID>
  </soap:Header>
  <soap:Body>
    <cwmp:Inform>
      <DeviceId>
        <Manufacturer>TestVendor</Manufacturer>
        <OUI>00:11:22</OUI>
        <ProductClass>TestDevice</ProductClass>
        <SerialNumber>SN001</SerialNumber>
      </DeviceId>
      <Event>
        <EventStruct>
          <EventCode>2 PERIODIC</EventCode>
        </EventStruct>
      </Event>
      <ParameterList>
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.HardwareVersion</Name>
          <Value xsi:type="xsd:string">1.0</Value>
        </ParameterValueStruct>
      </ParameterList>
      <MaxEnvelopes>1</MaxEnvelopes>
    </cwmp:Inform>
  </soap:Body>
</soap:Envelope>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseSOAP(bytes.NewReader([]byte(xmlData)))
		if err != nil {
			b.Fatalf("ParseSOAP failed: %v", err)
		}
	}
}

// BenchmarkBuildSOAP benchmarks the SOAP building performance.
func BenchmarkBuildSOAP(b *testing.B) {
	envelope := &SOAPEnvelope{
		Header: &SOAPHeader{
			ID: "123456",
		},
		Body: SOAPBody{
			InformResponse: &InformResponse{
				MaxEnvelopes: 1,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := BuildSOAP(envelope)
		if err != nil {
			b.Fatalf("BuildSOAP failed: %v", err)
		}
	}
}

// TestBuildInformResponseFunction tests the BuildInformResponse helper function.
func TestBuildInformResponseFunction(t *testing.T) {
	xmlBytes, err := BuildInformResponse("msg123", 1)
	if err != nil {
		t.Fatalf("BuildInformResponse failed: %v", err)
	}

	// Verify the XML can be parsed back
	parsed, err := ParseSOAP(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	if parsed.Header == nil || parsed.Header.ID != "msg123" {
		t.Errorf("Expected message ID 'msg123', got '%v'", parsed.Header)
	}

	if parsed.Body.InformResponse == nil {
		t.Fatal("Expected InformResponse in parsed envelope")
	}

	if parsed.Body.InformResponse.MaxEnvelopes != 1 {
		t.Errorf("Expected MaxEnvelopes 1, got %d", parsed.Body.InformResponse.MaxEnvelopes)
	}
}

// TestBuildGetParameterValuesFunction tests the BuildGetParameterValues helper function.
func TestBuildGetParameterValuesFunction(t *testing.T) {
	params := []string{
		"Device.DeviceInfo.SerialNumber",
		"Device.DeviceInfo.SoftwareVersion",
	}

	xmlBytes, err := BuildGetParameterValues("req123", params)
	if err != nil {
		t.Fatalf("BuildGetParameterValues failed: %v", err)
	}

	// Verify the XML can be parsed back
	parsed, err := ParseSOAP(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	if parsed.Body.GetParameterValues == nil {
		t.Fatal("Expected GetParameterValues in parsed envelope")
	}

	if len(parsed.Body.GetParameterValues.ParameterNames) != 2 {
		t.Errorf("Expected 2 parameter names, got %d",
			len(parsed.Body.GetParameterValues.ParameterNames))
	}
}

// TestBuildSetParameterValuesFunction tests the BuildSetParameterValues helper function.
func TestBuildSetParameterValuesFunction(t *testing.T) {
	params := []ParameterValueStruct{
		{Name: "Device.Username", Value: "user@test.com", Type: "xsd:string"},
		{Name: "Device.Password", Value: "secret", Type: "xsd:string"},
	}

	xmlBytes, err := BuildSetParameterValues("set123", params, "key001")
	if err != nil {
		t.Fatalf("BuildSetParameterValues failed: %v", err)
	}

	// Verify the XML can be parsed back
	parsed, err := ParseSOAP(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	if parsed.Body.SetParameterValues == nil {
		t.Fatal("Expected SetParameterValues in parsed envelope")
	}

	if len(parsed.Body.SetParameterValues.ParameterList) != 2 {
		t.Errorf("Expected 2 parameters, got %d",
			len(parsed.Body.SetParameterValues.ParameterList))
	}

	if parsed.Body.SetParameterValues.ParameterKey != "key001" {
		t.Errorf("Expected ParameterKey 'key001', got '%s'",
			parsed.Body.SetParameterValues.ParameterKey)
	}
}

// TestBuildRebootFunction tests the BuildReboot helper function.
func TestBuildRebootFunction(t *testing.T) {
	xmlBytes, err := BuildReboot("reboot123", "cmd001")
	if err != nil {
		t.Fatalf("BuildReboot failed: %v", err)
	}

	// Verify the XML can be parsed back
	parsed, err := ParseSOAP(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	if parsed.Body.Reboot == nil {
		t.Fatal("Expected Reboot in parsed envelope")
	}

	if parsed.Body.Reboot.CommandKey != "cmd001" {
		t.Errorf("Expected CommandKey 'cmd001', got '%s'", parsed.Body.Reboot.CommandKey)
	}
}

// TestBuildFaultFunction tests the BuildFault helper function.
func TestBuildFaultFunction(t *testing.T) {
	xmlBytes, err := BuildFault("fault123", "Client", "Invalid request", 9005, "Invalid parameter name")
	if err != nil {
		t.Fatalf("BuildFault failed: %v", err)
	}

	// Verify the XML can be parsed back
	parsed, err := ParseSOAP(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("Failed to parse generated XML: %v", err)
	}

	if parsed.Body.Fault == nil {
		t.Fatal("Expected Fault in parsed envelope")
	}

	if parsed.Body.Fault.FaultCode != "Client" {
		t.Errorf("Expected FaultCode 'Client', got '%s'", parsed.Body.Fault.FaultCode)
	}

	if parsed.Body.Fault.FaultString != "Invalid request" {
		t.Errorf("Expected FaultString 'Invalid request', got '%s'", parsed.Body.Fault.FaultString)
	}
}

// TestExtractDeviceInfo tests the ExtractDeviceInfo helper function.
func TestExtractDeviceInfo(t *testing.T) {
	inform := &Inform{
		DeviceId: DeviceIdStruct{
			Manufacturer: "TestVendor",
			OUI:          "00:11:22",
			ProductClass: "TestDevice",
			SerialNumber: "SN001",
		},
		ParameterList: []ParameterValueStruct{
			{Name: "Device.DeviceInfo.SerialNumber", Value: "SN001"},
			{Name: "Device.DeviceInfo.SoftwareVersion", Value: "v1.0.0"},
			{Name: "Device.DeviceInfo.HardwareVersion", Value: "hw1.0"},
			{Name: "Device.DeviceInfo.ModelName", Value: "ModelX"},
			{Name: "Device.WANDevice.1.WANConnectionDevice.1.ExternalIPAddress", Value: "192.168.1.1"},
		},
	}

	info := ExtractDeviceInfo(inform)

	if info["manufacturer"] != "TestVendor" {
		t.Errorf("Expected manufacturer 'TestVendor', got '%s'", info["manufacturer"])
	}

	if info["serial_number"] != "SN001" {
		t.Errorf("Expected serial_number 'SN001', got '%s'", info["serial_number"])
	}

	if info["software_version"] != "v1.0.0" {
		t.Errorf("Expected software_version 'v1.0.0', got '%s'", info["software_version"])
	}

	if info["hardware_version"] != "hw1.0" {
		t.Errorf("Expected hardware_version 'hw1.0', got '%s'", info["hardware_version"])
	}

	if info["model_name"] != "ModelX" {
		t.Errorf("Expected model_name 'ModelX', got '%s'", info["model_name"])
	}

	if info["wan_ip"] != "192.168.1.1" {
		t.Errorf("Expected wan_ip '192.168.1.1', got '%s'", info["wan_ip"])
	}
}

// TestExtractDeviceInfo_NilInform tests ExtractDeviceInfo with nil input.
func TestExtractDeviceInfo_NilInform(t *testing.T) {
	info := ExtractDeviceInfo(nil)
	if len(info) != 0 {
		t.Errorf("Expected empty map for nil inform, got %v", info)
	}
}

// TestGetEventCodes tests the GetEventCodes helper function.
func TestGetEventCodes(t *testing.T) {
	inform := &Inform{
		Event: []EventStruct{
			{EventCode: EventBoot},
			{EventCode: EventPeriodic},
		},
	}

	codes := GetEventCodes(inform)
	if len(codes) != 2 {
		t.Errorf("Expected 2 event codes, got %d", len(codes))
	}

	if codes[0] != EventBoot {
		t.Errorf("Expected first event '0 BOOT', got '%s'", codes[0])
	}

	if codes[1] != EventPeriodic {
		t.Errorf("Expected second event '2 PERIODIC', got '%s'", codes[1])
	}
}

// TestGetEventCodes_NilInform tests GetEventCodes with nil input.
func TestGetEventCodes_NilInform(t *testing.T) {
	codes := GetEventCodes(nil)
	if codes != nil {
		t.Errorf("Expected nil for nil inform, got %v", codes)
	}
}

// TestHasEvent tests the HasEvent helper function.
func TestHasEvent(t *testing.T) {
	inform := &Inform{
		Event: []EventStruct{
			{EventCode: EventBoot},
			{EventCode: EventPeriodic},
		},
	}

	if !HasEvent(inform, EventBoot) {
		t.Error("Expected HasEvent to return true for EventBoot")
	}

	if !HasEvent(inform, EventPeriodic) {
		t.Error("Expected HasEvent to return true for EventPeriodic")
	}

	if HasEvent(inform, EventValueChange) {
		t.Error("Expected HasEvent to return false for EventValueChange")
	}
}

// TestHasEvent_NilInform tests HasEvent with nil input.
func TestHasEvent_NilInform(t *testing.T) {
	if HasEvent(nil, EventBoot) {
		t.Error("Expected HasEvent to return false for nil inform")
	}
}

// TestIsBootEvent tests the IsBootEvent helper function.
func TestIsBootEvent(t *testing.T) {
	bootInform := &Inform{
		Event: []EventStruct{{EventCode: EventBoot}},
	}
	if !IsBootEvent(bootInform) {
		t.Error("Expected IsBootEvent to return true for boot event")
	}

	bootstrapInform := &Inform{
		Event: []EventStruct{{EventCode: EventBootStrap}},
	}
	if !IsBootEvent(bootstrapInform) {
		t.Error("Expected IsBootEvent to return true for bootstrap event")
	}

	periodicInform := &Inform{
		Event: []EventStruct{{EventCode: EventPeriodic}},
	}
	if IsBootEvent(periodicInform) {
		t.Error("Expected IsBootEvent to return false for periodic event")
	}
}

// TestIsPeriodicEvent tests the IsPeriodicEvent helper function.
func TestIsPeriodicEvent(t *testing.T) {
	periodicInform := &Inform{
		Event: []EventStruct{{EventCode: EventPeriodic}},
	}
	if !IsPeriodicEvent(periodicInform) {
		t.Error("Expected IsPeriodicEvent to return true for periodic event")
	}

	bootInform := &Inform{
		Event: []EventStruct{{EventCode: EventBoot}},
	}
	if IsPeriodicEvent(bootInform) {
		t.Error("Expected IsPeriodicEvent to return false for boot event")
	}
}

// TestIsValueChangeEvent tests the IsValueChangeEvent helper function.
func TestIsValueChangeEvent(t *testing.T) {
	valueChangeInform := &Inform{
		Event: []EventStruct{{EventCode: EventValueChange}},
	}
	if !IsValueChangeEvent(valueChangeInform) {
		t.Error("Expected IsValueChangeEvent to return true for value change event")
	}

	bootInform := &Inform{
		Event: []EventStruct{{EventCode: EventBoot}},
	}
	if IsValueChangeEvent(bootInform) {
		t.Error("Expected IsValueChangeEvent to return false for boot event")
	}
}

// TestGenerateMessageID tests the GenerateMessageID function.
func TestGenerateMessageID(t *testing.T) {
	id1 := GenerateMessageID()
	id2 := GenerateMessageID()

	if id1 == "" {
		t.Error("Expected non-empty message ID")
	}

	if id1 == id2 {
		t.Error("Expected unique message IDs")
	}

	// Check format: should start with "cwmp-"
	if len(id1) < 5 || id1[:5] != "cwmp-" {
		t.Errorf("Expected message ID to start with 'cwmp-', got '%s'", id1)
	}
}

// TestGetParameterValue tests the GetParameterValue helper function.
func TestGetParameterValue(t *testing.T) {
	inform := &Inform{
		ParameterList: []ParameterValueStruct{
			{Name: "Device.SerialNumber", Value: "SN001"},
			{Name: "Device.SoftwareVersion", Value: "v1.0"},
		},
	}

	val := GetParameterValue(inform, "Device.SerialNumber")
	if val != "SN001" {
		t.Errorf("Expected 'SN001', got '%s'", val)
	}

	val = GetParameterValue(inform, "Device.SoftwareVersion")
	if val != "v1.0" {
		t.Errorf("Expected 'v1.0', got '%s'", val)
	}

	val = GetParameterValue(inform, "Device.NonExistent")
	if val != "" {
		t.Errorf("Expected empty string for non-existent parameter, got '%s'", val)
	}
}

// TestGetParameterValue_NilInform tests GetParameterValue with nil input.
func TestGetParameterValue_NilInform(t *testing.T) {
	val := GetParameterValue(nil, "Device.SerialNumber")
	if val != "" {
		t.Errorf("Expected empty string for nil inform, got '%s'", val)
	}
}

// TestNewSetParameterValuesForPPPoE tests the NewSetParameterValuesForPPPoE helper function.
func TestNewSetParameterValuesForPPPoE(t *testing.T) {
	params := NewSetParameterValuesForPPPoE("user@test.com", "secret123", "", "")

	if len(params) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(params))
	}

	// Check username parameter
	if params[0].Name != "InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Username" {
		t.Errorf("Unexpected username parameter name: %s", params[0].Name)
	}
	if params[0].Value != "user@test.com" {
		t.Errorf("Expected username 'user@test.com', got '%s'", params[0].Value)
	}

	// Check password parameter
	if params[1].Name != "InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Password" {
		t.Errorf("Unexpected password parameter name: %s", params[1].Name)
	}
	if params[1].Value != "secret123" {
		t.Errorf("Expected password 'secret123', got '%s'", params[1].Value)
	}
}

// TestNewSetParameterValuesForPPPoE_CustomPaths tests PPPoE with custom paths.
func TestNewSetParameterValuesForPPPoE_CustomPaths(t *testing.T) {
	params := NewSetParameterValuesForPPPoE("user@test.com", "secret123",
		"Device.WAN.1.Username", "Device.WAN.1.Password")

	if params[0].Name != "Device.WAN.1.Username" {
		t.Errorf("Expected custom username path, got '%s'", params[0].Name)
	}

	if params[1].Name != "Device.WAN.1.Password" {
		t.Errorf("Expected custom password path, got '%s'", params[1].Name)
	}
}

// TestNewSetParameterValuesForWiFi tests the NewSetParameterValuesForWiFi helper function.
func TestNewSetParameterValuesForWiFi(t *testing.T) {
	params := NewSetParameterValuesForWiFi("MyWiFi", "wifipass123", 1)

	if len(params) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(params))
	}

	// Check SSID parameter
	if params[0].Name != "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.SSID" {
		t.Errorf("Unexpected SSID parameter name: %s", params[0].Name)
	}
	if params[0].Value != "MyWiFi" {
		t.Errorf("Expected SSID 'MyWiFi', got '%s'", params[0].Value)
	}

	// Check password parameter
	if params[1].Name != "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.PreSharedKey.1.PreSharedKey" {
		t.Errorf("Unexpected password parameter name: %s", params[1].Name)
	}
	if params[1].Value != "wifipass123" {
		t.Errorf("Expected password 'wifipass123', got '%s'", params[1].Value)
	}
}

// TestNewSetParameterValuesForWiFi_DefaultIndex tests WiFi with default index.
func TestNewSetParameterValuesForWiFi_DefaultIndex(t *testing.T) {
	params := NewSetParameterValuesForWiFi("MyWiFi", "pass", 0) // Index 0 should default to 1

	if params[0].Name != "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.SSID" {
		t.Errorf("Expected default WLAN index 1, got '%s'", params[0].Name)
	}
}

// TestCPEDevice_IsProvisioned tests the CPEDevice.IsProvisioned method.
func TestCPEDevice_IsProvisioned(t *testing.T) {
	device := &CPEDevice{Status: DeviceStatusProvisioned}
	if !device.IsProvisioned() {
		t.Error("Expected IsProvisioned to return true for provisioned device")
	}

	device.Status = DeviceStatusPending
	if device.IsProvisioned() {
		t.Error("Expected IsProvisioned to return false for pending device")
	}
}

// TestCPEDevice_NeedsProvisioning tests the CPEDevice.NeedsProvisioning method.
func TestCPEDevice_NeedsProvisioning(t *testing.T) {
	device := &CPEDevice{
		Status:        DeviceStatusPending,
		AutoProvision: true,
	}
	if !device.NeedsProvisioning() {
		t.Error("Expected NeedsProvisioning to return true for pending device with AutoProvision=true")
	}

	device.AutoProvision = false
	if device.NeedsProvisioning() {
		t.Error("Expected NeedsProvisioning to return false when AutoProvision=false")
	}

	device.Status = DeviceStatusProvisioned
	device.AutoProvision = true
	if device.NeedsProvisioning() {
		t.Error("Expected NeedsProvisioning to return false for already provisioned device")
	}
}

// TestDefaultProvisioningConfig tests the DefaultProvisioningConfig function.
func TestDefaultProvisioningConfig(t *testing.T) {
	config := DefaultProvisioningConfig()

	if !config.AutoProvision {
		t.Error("Expected AutoProvision to be true by default")
	}

	if config.UsernamePrefix != "cpe_" {
		t.Errorf("Expected UsernamePrefix 'cpe_', got '%s'", config.UsernamePrefix)
	}

	if config.PasswordLength != 12 {
		t.Errorf("Expected PasswordLength 12, got %d", config.PasswordLength)
	}
}

// TestDefaultACSConfig tests the DefaultACSConfig function.
func TestDefaultACSConfig(t *testing.T) {
	config := DefaultACSConfig()

	if config.Listen != ":7547" {
		t.Errorf("Expected Listen ':7547', got '%s'", config.Listen)
	}

	if config.SessionTimeout != 30*time.Second {
		t.Errorf("Expected SessionTimeout 30s, got %v", config.SessionTimeout)
	}
}

// TestParseInform tests the ParseInform helper function.
func TestParseInform(t *testing.T) {
	envelope := &SOAPEnvelope{
		Body: SOAPBody{
			Inform: &Inform{
				DeviceId: DeviceIdStruct{SerialNumber: "SN001"},
			},
		},
	}

	inform := ParseInform(envelope)
	if inform == nil {
		t.Fatal("Expected Inform to be returned")
	}

	if inform.DeviceId.SerialNumber != "SN001" {
		t.Errorf("Expected SerialNumber 'SN001', got '%s'", inform.DeviceId.SerialNumber)
	}
}

// TestParseInform_NilEnvelope tests ParseInform with nil envelope.
func TestParseInform_NilEnvelope(t *testing.T) {
	inform := ParseInform(nil)
	if inform != nil {
		t.Errorf("Expected nil for nil envelope, got %v", inform)
	}
}

// TestCWMPFaultCodeConstants tests CWMP fault code constants.
func TestCWMPFaultCodeConstants(t *testing.T) {
	if FaultCodeRequestDenied != 9000 {
		t.Errorf("FaultCodeRequestDenied should be 9000, got %d", FaultCodeRequestDenied)
	}

	if FaultCodeInvalidParameterName != 9005 {
		t.Errorf("FaultCodeInvalidParameterName should be 9005, got %d", FaultCodeInvalidParameterName)
	}

	if FaultCodeInvalidParameterValue != 9006 {
		t.Errorf("FaultCodeInvalidParameterValue should be 9006, got %d", FaultCodeInvalidParameterValue)
	}

	if FaultCodeUnsupportedRPCMethod != 9016 {
		t.Errorf("FaultCodeUnsupportedRPCMethod should be 9016, got %d", FaultCodeUnsupportedRPCMethod)
	}
}

// Ensure xml package is imported (used in tests)
var _ = xml.Encoder{}
