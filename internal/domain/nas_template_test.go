package domain

import "testing"

func TestNASTemplate_ValidTemplate_ShouldPassValidation(t *testing.T) {
	template := &NASTemplate{
		VendorCode: "huawei",
		Name:       "Huawei ME60 Standard",
		Attributes: []TemplateAttribute{
			{AttrName: "Input-Average-Rate", VendorAttr: "Huawei-Input-Average-Rate", ValueType: "integer"},
		},
	}

	err := template.Validate()
	if err != nil {
		t.Fatalf("expected valid template, got error: %v", err)
	}
}

func TestNASTemplate_DuplicateAttributeNames_ShouldFail(t *testing.T) {
	template := &NASTemplate{
		VendorCode: "cisco",
		Name:       "Cisco ASR Duplicate",
		Attributes: []TemplateAttribute{
			{AttrName: "Framed-IP-Address", VendorAttr: "Cisco-AVPair", ValueType: "string"},
			{AttrName: "Framed-IP-Address", VendorAttr: "Cisco-IP", ValueType: "string"},
		},
	}

	err := template.Validate()
	if err == nil {
		t.Fatal("expected validation error for duplicate attributes")
	}
}
