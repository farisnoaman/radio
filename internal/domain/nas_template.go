package domain

import (
	"errors"
	"fmt"
	"time"
)

// NASTemplate represents a vendor-specific RADIUS attribute template.
// Templates define how standard RADIUS attributes map to vendor-specific attributes
// for different NAS equipment (Huawei ME60, Cisco ASR, Juniper MX, etc.).
//
// Example:
//	A Huawei ME60 template might map standard "Framed-IP-Address" to
//	vendor-specific "Huawei-Input-Average-Rate" with value type integer.
type NASTemplate struct {
	ID         int64              `json:"id,string" gorm:"primaryKey"`
	TenantID   int64              `json:"tenant_id" gorm:"index"`
	VendorCode string             `json:"vendor_code" gorm:"not null;index"` // huawei, cisco, juniper, ubiquiti
	Name       string             `json:"name" gorm:"not null;size:200"`
	IsDefault  bool               `json:"is_default" gorm:"default:false"`
	Attributes []TemplateAttribute `json:"attributes" gorm:"serializer:json"`
	Remark     string             `json:"remark" gorm:"size:500"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

// TableName specifies the table name for NASTemplate.
func (NASTemplate) TableName() string {
	return "nas_template"
}

// TemplateAttribute defines a single attribute mapping in a template.
type TemplateAttribute struct {
	AttrName     string `json:"attr_name"`              // Standard RADIUS attribute name
	VendorAttr   string `json:"vendor_attr"`            // Vendor-specific attribute identifier
	ValueType    string `json:"value_type"`             // string, integer, ipaddr
	IsRequired   bool   `json:"is_required"`            // Whether this attribute is required
	DefaultValue string `json:"default_value,omitempty"` // Optional default value
}

// Validate checks if the template configuration is valid.
//
// Returns error if:
//   - VendorCode is empty
//   - Name is empty
//   - No attributes defined
//   - Duplicate attribute names
//   - Invalid value types
func (t *NASTemplate) Validate() error {
	if t.VendorCode == "" {
		return errors.New("vendor_code is required")
	}
	if t.Name == "" {
		return errors.New("template name is required")
	}
	if len(t.Attributes) == 0 {
		return errors.New("at least one attribute is required")
	}

	// Check for duplicate attribute names
	attrNames := make(map[string]bool)
	for _, attr := range t.Attributes {
		if attr.AttrName == "" {
			return errors.New("attribute name cannot be empty")
		}
		if attr.VendorAttr == "" {
			return errors.New("vendor attribute cannot be empty")
		}
		if attr.ValueType == "" {
			return errors.New("value type is required")
		}

		// Validate value type
		switch attr.ValueType {
		case "string", "integer", "ipaddr":
			// Valid types
		default:
			return fmt.Errorf("invalid value type: %s (must be string, integer, or ipaddr)", attr.ValueType)
		}

		if attrNames[attr.AttrName] {
			return fmt.Errorf("duplicate attribute name: %s", attr.AttrName)
		}
		attrNames[attr.AttrName] = true
	}

	return nil
}

// GetAttribute returns the template attribute for a given standard attribute name.
func (t *NASTemplate) GetAttribute(attrName string) (*TemplateAttribute, bool) {
	for _, attr := range t.Attributes {
		if attr.AttrName == attrName {
			return &attr, true
		}
	}
	return nil, false
}
