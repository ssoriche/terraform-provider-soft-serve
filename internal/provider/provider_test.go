package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestSoftServeProviderMetadata(t *testing.T) {
	p := &SoftServeProvider{version: "1.2.3"}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "softserve" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "softserve")
	}
	if resp.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", resp.Version, "1.2.3")
	}
}

func TestSoftServeProviderSchema(t *testing.T) {
	p := &SoftServeProvider{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %s", resp.Diagnostics)
	}

	expectedAttrs := []string{"host", "port", "username", "private_key_path", "identity_file", "use_agent"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q", attr)
		}
	}

	if len(resp.Schema.Attributes) != len(expectedAttrs) {
		t.Errorf("got %d attributes, want %d", len(resp.Schema.Attributes), len(expectedAttrs))
	}
}

func TestSoftServeProviderSchemaAttributeTypes(t *testing.T) {
	p := &SoftServeProvider{}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	tests := []struct {
		name     string
		attrType string
	}{
		{"host", "StringAttribute"},
		{"port", "Int64Attribute"},
		{"username", "StringAttribute"},
		{"private_key_path", "StringAttribute"},
		{"identity_file", "StringAttribute"},
		{"use_agent", "BoolAttribute"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr, ok := resp.Schema.Attributes[tt.name]
			if !ok {
				t.Fatalf("attribute %q not found", tt.name)
			}
			// All provider attributes should be optional
			if !attr.IsOptional() {
				t.Errorf("attribute %q should be optional", tt.name)
			}
		})
	}
}

func TestSoftServeProviderSchemaAllOptional(t *testing.T) {
	p := &SoftServeProvider{}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	for name, attr := range resp.Schema.Attributes {
		if attr.IsRequired() {
			t.Errorf("attribute %q should be optional, not required", name)
		}
	}
}

func TestSoftServeProviderSchemaDescription(t *testing.T) {
	p := &SoftServeProvider{}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	if resp.Schema.Description == "" {
		t.Error("schema description should not be empty")
	}

	// Verify each attribute has a description
	for name, attr := range resp.Schema.Attributes {
		sa, ok := attr.(schema.StringAttribute)
		if ok && sa.Description == "" {
			t.Errorf("attribute %q missing description", name)
			continue
		}
		ia, ok := attr.(schema.Int64Attribute)
		if ok && ia.Description == "" {
			t.Errorf("attribute %q missing description", name)
			continue
		}
		ba, ok := attr.(schema.BoolAttribute)
		if ok && ba.Description == "" {
			t.Errorf("attribute %q missing description", name)
		}
	}
}

func TestSoftServeProviderResources(t *testing.T) {
	p := &SoftServeProvider{}

	resources := p.Resources(context.Background())

	expectedCount := 4
	if len(resources) != expectedCount {
		t.Fatalf("got %d resources, want %d", len(resources), expectedCount)
	}

	for i, factory := range resources {
		r := factory()
		if r == nil {
			t.Errorf("resource factory [%d] returned nil", i)
		}
	}
}

func TestSoftServeProviderResourceTypes(t *testing.T) {
	p := &SoftServeProvider{}
	resources := p.Resources(context.Background())

	expectedTypes := map[string]bool{
		"softserve_repository":              false,
		"softserve_user":                    false,
		"softserve_repository_collaborator": false,
		"softserve_server_settings":         false,
	}

	for _, factory := range resources {
		r := factory()
		metaResp := &resource.MetadataResponse{}
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "softserve"}, metaResp)

		if _, ok := expectedTypes[metaResp.TypeName]; !ok {
			t.Errorf("unexpected resource type: %q", metaResp.TypeName)
		}
		expectedTypes[metaResp.TypeName] = true
	}

	for typeName, found := range expectedTypes {
		if !found {
			t.Errorf("missing expected resource type: %q", typeName)
		}
	}
}

func TestSoftServeProviderDataSources(t *testing.T) {
	p := &SoftServeProvider{}

	dataSources := p.DataSources(context.Background())

	if dataSources != nil {
		t.Errorf("expected nil data sources, got %d", len(dataSources))
	}
}

func TestNew(t *testing.T) {
	factory := New("test-version")

	p := factory()
	if p == nil {
		t.Fatal("New() factory returned nil provider")
	}

	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.Version != "test-version" {
		t.Errorf("Version = %q, want %q", resp.Version, "test-version")
	}
}

func TestProviderImplementsInterface(t *testing.T) {
	var _ provider.Provider = &SoftServeProvider{}
}
