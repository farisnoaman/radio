package templates

import (
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

// SeedDefaultTemplates creates default vendor templates for common NAS equipment.
// These templates provide out-of-the-box support for major vendors.
func SeedDefaultTemplates(db *gorm.DB) error {
	templates := []domain.NASTemplate{
		// Huawei ME60/MediaAccess template
		{
			VendorCode: "huawei",
			Name:       "Huawei ME60 Standard",
			IsDefault:  true,
			Attributes: []domain.TemplateAttribute{
				{AttrName: "Framed-IP-Address", VendorAttr: "Huawei-IP-Address", ValueType: "ipaddr", IsRequired: true},
				{AttrName: "Input-Average-Rate", VendorAttr: "Huawei-Input-Average-Rate", ValueType: "integer"},
				{AttrName: "Output-Average-Rate", VendorAttr: "Huawei-Output-Average-Rate", ValueType: "integer"},
				{AttrName: "Acct-Interim-Interval", VendorAttr: "Huawei-Acct-Interim-Interval", ValueType: "integer"},
			},
			Remark: "Standard template for Huawei ME60 series BRAS",
		},
		// Cisco ASR/ISR template
		{
			VendorCode: "cisco",
			Name:       "Cisco ASR Standard",
			IsDefault:  true,
			Attributes: []domain.TemplateAttribute{
				{AttrName: "Framed-IP-Address", VendorAttr: "Cisco-AVPair = \"ip:addr\"", ValueType: "string", IsRequired: true},
				{AttrName: "Cisco-AVPair", VendorAttr: "Cisco-AVPair", ValueType: "string"},
				{AttrName: "Session-Timeout", VendorAttr: "Session-Timeout", ValueType: "integer"},
			},
			Remark: "Standard template for Cisco ASR/ISR series routers",
		},
		// Juniper MX template
		{
			VendorCode: "juniper",
			Name:       "Juniper MX Standard",
			IsDefault:  true,
			Attributes: []domain.TemplateAttribute{
				{AttrName: "Framed-IP-Address", VendorAttr: "Juniper-Local-Frame-IP-Address", ValueType: "ipaddr", IsRequired: true},
				{AttrName: "Juniper-Local-Loopback-IP", VendorAttr: "Juniper-Local-Loopback-IP", ValueType: "ipaddr"},
				{AttrName: "Input-Filter", VendorAttr: "Juniper-Input-Filter", ValueType: "string"},
			},
			Remark: "Standard template for Juniper MX series routers",
		},
		// Ubiquiti UniFi template
		{
			VendorCode: "ubiquiti",
			Name:       "Ubiquiti UniFi Standard",
			IsDefault:  true,
			Attributes: []domain.TemplateAttribute{
				{AttrName: "Framed-IP-Address", VendorAttr: "Framed-IP-Address", ValueType: "ipaddr", IsRequired: true},
				{AttrName: "Ubiquiti-Policy-Name", VendorAttr: "Ubiquiti-Policy-Name", ValueType: "string"},
				{AttrName: "Tunnel-Type", VendorAttr: "Tunnel-Type", ValueType: "integer"},
			},
			Remark: "Standard template for Ubiquiti UniFi access points",
		},
	}

	for _, template := range templates {
		// Check if template already exists
		var count int64
		db.Model(&domain.NASTemplate{}).
			Where("vendor_code = ? AND name = ?", template.VendorCode, template.Name).
			Count(&count)

		if count == 0 {
			if err := db.Create(&template).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
