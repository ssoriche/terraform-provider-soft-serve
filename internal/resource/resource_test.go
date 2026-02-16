package resource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// --- Repository Resource Tests ---

func TestRepositoryResourceMetadata(t *testing.T) {
	r := NewRepositoryResource()
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "softserve"}, resp)

	if resp.TypeName != "softserve_repository" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "softserve_repository")
	}
}

func TestRepositoryResourceSchema(t *testing.T) {
	r := NewRepositoryResource()
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %s", resp.Diagnostics)
	}

	expectedAttrs := []string{"id", "name", "description", "project_name", "private", "hidden"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q", attr)
		}
	}

	if len(resp.Schema.Attributes) != len(expectedAttrs) {
		t.Errorf("got %d attributes, want %d", len(resp.Schema.Attributes), len(expectedAttrs))
	}
}

func TestRepositoryResourceSchemaRequired(t *testing.T) {
	r := NewRepositoryResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	nameAttr := resp.Schema.Attributes["name"]
	if !nameAttr.IsRequired() {
		t.Error("name attribute should be required")
	}

	idAttr := resp.Schema.Attributes["id"]
	if !idAttr.IsComputed() {
		t.Error("id attribute should be computed")
	}
}

func TestRepositoryResourceSchemaOptionalComputed(t *testing.T) {
	r := NewRepositoryResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	optionalComputed := []string{"description", "project_name", "private", "hidden"}
	for _, name := range optionalComputed {
		attr := resp.Schema.Attributes[name]
		if !attr.IsOptional() {
			t.Errorf("%q should be optional", name)
		}
		if !attr.IsComputed() {
			t.Errorf("%q should be computed", name)
		}
	}
}

func TestRepositoryResourceSchemaDefaults(t *testing.T) {
	r := NewRepositoryResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	// private and hidden should have defaults
	privateAttr, ok := resp.Schema.Attributes["private"].(schema.BoolAttribute)
	if !ok {
		t.Fatal("private attribute should be BoolAttribute")
	}
	if privateAttr.Default == nil {
		t.Error("private attribute should have a default value")
	}

	hiddenAttr, ok := resp.Schema.Attributes["hidden"].(schema.BoolAttribute)
	if !ok {
		t.Fatal("hidden attribute should be BoolAttribute")
	}
	if hiddenAttr.Default == nil {
		t.Error("hidden attribute should have a default value")
	}
}

func TestRepositoryResourceSchemaNameRequiresReplace(t *testing.T) {
	r := NewRepositoryResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	nameAttr, ok := resp.Schema.Attributes["name"].(schema.StringAttribute)
	if !ok {
		t.Fatal("name attribute should be StringAttribute")
	}
	if len(nameAttr.PlanModifiers) == 0 {
		t.Error("name attribute should have plan modifiers (RequiresReplace)")
	}
}

func TestRepositoryResourceDescription(t *testing.T) {
	r := NewRepositoryResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	if resp.Schema.Description == "" {
		t.Error("schema description should not be empty")
	}
}

func TestRepositoryResourceImplementsInterfaces(t *testing.T) {
	r := NewRepositoryResource()
	if _, ok := r.(resource.ResourceWithImportState); !ok {
		t.Error("RepositoryResource should implement ResourceWithImportState")
	}
}

func TestRepositoryResourceConfigure_NilProviderData(t *testing.T) {
	r := &RepositoryResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: nil,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error with nil provider data: %s", resp.Diagnostics)
	}
}

func TestRepositoryResourceConfigure_WrongType(t *testing.T) {
	r := &RepositoryResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "wrong-type",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error with wrong provider data type")
	}
}

// --- User Resource Tests ---

func TestUserResourceMetadata(t *testing.T) {
	r := NewUserResource()
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "softserve"}, resp)

	if resp.TypeName != "softserve_user" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "softserve_user")
	}
}

func TestUserResourceSchema(t *testing.T) {
	r := NewUserResource()
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %s", resp.Diagnostics)
	}

	expectedAttrs := []string{"id", "username", "admin", "public_keys"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q", attr)
		}
	}

	if len(resp.Schema.Attributes) != len(expectedAttrs) {
		t.Errorf("got %d attributes, want %d", len(resp.Schema.Attributes), len(expectedAttrs))
	}
}

func TestUserResourceSchemaRequired(t *testing.T) {
	r := NewUserResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	usernameAttr := resp.Schema.Attributes["username"]
	if !usernameAttr.IsRequired() {
		t.Error("username attribute should be required")
	}

	idAttr := resp.Schema.Attributes["id"]
	if !idAttr.IsComputed() {
		t.Error("id attribute should be computed")
	}
}

func TestUserResourceSchemaUsernameRequiresReplace(t *testing.T) {
	r := NewUserResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	usernameAttr, ok := resp.Schema.Attributes["username"].(schema.StringAttribute)
	if !ok {
		t.Fatal("username attribute should be StringAttribute")
	}
	if len(usernameAttr.PlanModifiers) == 0 {
		t.Error("username attribute should have plan modifiers (RequiresReplace)")
	}
}

func TestUserResourceSchemaAdminDefault(t *testing.T) {
	r := NewUserResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	adminAttr, ok := resp.Schema.Attributes["admin"].(schema.BoolAttribute)
	if !ok {
		t.Fatal("admin attribute should be BoolAttribute")
	}
	if adminAttr.Default == nil {
		t.Error("admin attribute should have a default value")
	}
	if !adminAttr.Optional {
		t.Error("admin attribute should be optional")
	}
	if !adminAttr.Computed {
		t.Error("admin attribute should be computed")
	}
}

func TestUserResourceSchemaPublicKeysIsSet(t *testing.T) {
	r := NewUserResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	_, ok := resp.Schema.Attributes["public_keys"].(schema.SetAttribute)
	if !ok {
		t.Error("public_keys attribute should be SetAttribute")
	}
}

func TestUserResourceImplementsInterfaces(t *testing.T) {
	r := NewUserResource()
	if _, ok := r.(resource.ResourceWithImportState); !ok {
		t.Error("UserResource should implement ResourceWithImportState")
	}
}

func TestUserResourceConfigure_NilProviderData(t *testing.T) {
	r := &UserResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: nil,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error with nil provider data: %s", resp.Diagnostics)
	}
}

func TestUserResourceConfigure_WrongType(t *testing.T) {
	r := &UserResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: 42,
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error with wrong provider data type")
	}
}

// --- Repository Collaborator Resource Tests ---

func TestRepositoryCollaboratorResourceMetadata(t *testing.T) {
	r := NewRepositoryCollaboratorResource()
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "softserve"}, resp)

	if resp.TypeName != "softserve_repository_collaborator" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "softserve_repository_collaborator")
	}
}

func TestRepositoryCollaboratorResourceSchema(t *testing.T) {
	r := NewRepositoryCollaboratorResource()
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %s", resp.Diagnostics)
	}

	expectedAttrs := []string{"id", "repository", "username", "access_level"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q", attr)
		}
	}

	if len(resp.Schema.Attributes) != len(expectedAttrs) {
		t.Errorf("got %d attributes, want %d", len(resp.Schema.Attributes), len(expectedAttrs))
	}
}

func TestRepositoryCollaboratorResourceSchemaRequired(t *testing.T) {
	r := NewRepositoryCollaboratorResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	repoAttr := resp.Schema.Attributes["repository"]
	if !repoAttr.IsRequired() {
		t.Error("repository attribute should be required")
	}

	usernameAttr := resp.Schema.Attributes["username"]
	if !usernameAttr.IsRequired() {
		t.Error("username attribute should be required")
	}
}

func TestRepositoryCollaboratorResourceSchemaRequiresReplace(t *testing.T) {
	r := NewRepositoryCollaboratorResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	repoAttr, ok := resp.Schema.Attributes["repository"].(schema.StringAttribute)
	if !ok {
		t.Fatal("repository attribute should be StringAttribute")
	}
	if len(repoAttr.PlanModifiers) == 0 {
		t.Error("repository attribute should have plan modifiers (RequiresReplace)")
	}

	usernameAttr, ok := resp.Schema.Attributes["username"].(schema.StringAttribute)
	if !ok {
		t.Fatal("username attribute should be StringAttribute")
	}
	if len(usernameAttr.PlanModifiers) == 0 {
		t.Error("username attribute should have plan modifiers (RequiresReplace)")
	}
}

func TestRepositoryCollaboratorResourceSchemaAccessLevelDefault(t *testing.T) {
	r := NewRepositoryCollaboratorResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	accessAttr, ok := resp.Schema.Attributes["access_level"].(schema.StringAttribute)
	if !ok {
		t.Fatal("access_level attribute should be StringAttribute")
	}
	if accessAttr.Default == nil {
		t.Error("access_level attribute should have a default value")
	}
	if !accessAttr.Optional {
		t.Error("access_level attribute should be optional")
	}
	if !accessAttr.Computed {
		t.Error("access_level attribute should be computed")
	}
}

func TestRepositoryCollaboratorResourceSchemaAccessLevelValidators(t *testing.T) {
	r := NewRepositoryCollaboratorResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	accessAttr, ok := resp.Schema.Attributes["access_level"].(schema.StringAttribute)
	if !ok {
		t.Fatal("access_level attribute should be StringAttribute")
	}
	if len(accessAttr.Validators) == 0 {
		t.Error("access_level attribute should have validators")
	}
}

func TestRepositoryCollaboratorResourceImplementsInterfaces(t *testing.T) {
	r := NewRepositoryCollaboratorResource()
	if _, ok := r.(resource.ResourceWithImportState); !ok {
		t.Error("RepositoryCollaboratorResource should implement ResourceWithImportState")
	}
}

func TestRepositoryCollaboratorResourceConfigure_NilProviderData(t *testing.T) {
	r := &RepositoryCollaboratorResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: nil,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error with nil provider data: %s", resp.Diagnostics)
	}
}

func TestRepositoryCollaboratorResourceConfigure_WrongType(t *testing.T) {
	r := &RepositoryCollaboratorResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: struct{}{},
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error with wrong provider data type")
	}
}

// --- Server Settings Resource Tests ---

func TestServerSettingsResourceMetadata(t *testing.T) {
	r := NewServerSettingsResource()
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "softserve"}, resp)

	if resp.TypeName != "softserve_server_settings" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "softserve_server_settings")
	}
}

func TestServerSettingsResourceSchema(t *testing.T) {
	r := NewServerSettingsResource()
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %s", resp.Diagnostics)
	}

	expectedAttrs := []string{"id", "allow_keyless", "anon_access"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q", attr)
		}
	}

	if len(resp.Schema.Attributes) != len(expectedAttrs) {
		t.Errorf("got %d attributes, want %d", len(resp.Schema.Attributes), len(expectedAttrs))
	}
}

func TestServerSettingsResourceSchemaNoRequired(t *testing.T) {
	r := NewServerSettingsResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	// All settings attributes should be optional or computed, none required
	for name, attr := range resp.Schema.Attributes {
		if attr.IsRequired() {
			t.Errorf("attribute %q should not be required", name)
		}
	}
}

func TestServerSettingsResourceSchemaAnonAccessValidator(t *testing.T) {
	r := NewServerSettingsResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	anonAttr, ok := resp.Schema.Attributes["anon_access"].(schema.StringAttribute)
	if !ok {
		t.Fatal("anon_access attribute should be StringAttribute")
	}
	if len(anonAttr.Validators) == 0 {
		t.Error("anon_access attribute should have validators")
	}
}

func TestServerSettingsResourceSchemaIDComputed(t *testing.T) {
	r := NewServerSettingsResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	idAttr := resp.Schema.Attributes["id"]
	if !idAttr.IsComputed() {
		t.Error("id attribute should be computed")
	}
}

func TestServerSettingsResourceIsSingleton(t *testing.T) {
	r := NewServerSettingsResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	if resp.Schema.Description == "" {
		t.Error("schema description should not be empty")
	}
}

func TestServerSettingsResourceDeleteIsNoop(t *testing.T) {
	r := &ServerSettingsResource{}
	resp := &resource.DeleteResponse{}

	// Delete on a singleton should not fail (it's a no-op)
	r.Delete(context.Background(), resource.DeleteRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Delete() should be a no-op for singleton resource, got errors: %s", resp.Diagnostics)
	}
}

func TestServerSettingsResourceImplementsInterfaces(t *testing.T) {
	r := NewServerSettingsResource()
	if _, ok := r.(resource.ResourceWithImportState); !ok {
		t.Error("ServerSettingsResource should implement ResourceWithImportState")
	}
}

func TestServerSettingsResourceConfigure_NilProviderData(t *testing.T) {
	r := &ServerSettingsResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: nil,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error with nil provider data: %s", resp.Diagnostics)
	}
}

func TestServerSettingsResourceConfigure_WrongType(t *testing.T) {
	r := &ServerSettingsResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: false,
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error with wrong provider data type")
	}
}

// --- Helper Function Tests ---

func TestToStringSet(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  map[string]struct{}
	}{
		{
			name:  "empty slice",
			input: []string{},
			want:  map[string]struct{}{},
		},
		{
			name:  "nil slice",
			input: nil,
			want:  map[string]struct{}{},
		},
		{
			name:  "single element",
			input: []string{"a"},
			want:  map[string]struct{}{"a": {}},
		},
		{
			name:  "multiple elements",
			input: []string{"a", "b", "c"},
			want: map[string]struct{}{
				"a": {},
				"b": {},
				"c": {},
			},
		},
		{
			name:  "duplicates are deduplicated",
			input: []string{"a", "b", "a", "c", "b"},
			want: map[string]struct{}{
				"a": {},
				"b": {},
				"c": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toStringSet(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("length = %d, want %d", len(got), len(tt.want))
			}
			for k := range tt.want {
				if _, ok := got[k]; !ok {
					t.Errorf("missing key %q", k)
				}
			}
		})
	}
}
