package resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ssoriche/terraform-provider-soft-serve/internal/ssh"
)

var (
	_ resource.Resource                = &ServerSettingsResource{}
	_ resource.ResourceWithImportState = &ServerSettingsResource{}
)

type ServerSettingsResource struct {
	client *ssh.Client
}

type ServerSettingsResourceModel struct {
	ID           types.String `tfsdk:"id"`
	AllowKeyless types.Bool   `tfsdk:"allow_keyless"`
	AnonAccess   types.String `tfsdk:"anon_access"`
}

func NewServerSettingsResource() resource.Resource {
	return &ServerSettingsResource{}
}

func (r *ServerSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_settings"
}

func (r *ServerSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Soft Serve server settings. This is a singleton resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Always \"settings\".",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allow_keyless": schema.BoolAttribute{
				Description: "Whether to allow keyless access to repositories.",
				Optional:    true,
				Computed:    true,
			},
			"anon_access": schema.StringAttribute{
				Description: "Default access level for anonymous users: no-access, read-only, read-write, or admin-access.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("no-access", "read-only", "read-write", "admin-access"),
				},
			},
		},
	}
}

func (r *ServerSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*ssh.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ssh.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *ServerSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServerSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.applySettings(&plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readSettingsState(&plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServerSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServerSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readSettingsState(&state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ServerSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServerSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.applySettings(&plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readSettingsState(&plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServerSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton resource: delete just removes from state, no server-side action
}

func (r *ServerSettingsResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var model ServerSettingsResourceModel

	resp.Diagnostics.Append(r.readSettingsState(&model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServerSettingsResource) applySettings(model *ServerSettingsResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if !model.AllowKeyless.IsNull() && !model.AllowKeyless.IsUnknown() {
		if err := r.client.SettingsSetAllowKeyless(model.AllowKeyless.ValueBool()); err != nil {
			diags.AddError("Error setting allow-keyless", err.Error())
			return diags
		}
	}

	if !model.AnonAccess.IsNull() && !model.AnonAccess.IsUnknown() {
		if err := r.client.SettingsSetAnonAccess(model.AnonAccess.ValueString()); err != nil {
			diags.AddError("Error setting anon-access", err.Error())
			return diags
		}
	}

	return diags
}

func (r *ServerSettingsResource) readSettingsState(model *ServerSettingsResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue("settings")

	allowKeyless, err := r.client.SettingsGetAllowKeyless()
	if err != nil {
		diags.AddError("Error reading allow-keyless", err.Error())
		return diags
	}
	model.AllowKeyless = types.BoolValue(allowKeyless)

	anonAccess, err := r.client.SettingsGetAnonAccess()
	if err != nil {
		diags.AddError("Error reading anon-access", err.Error())
		return diags
	}
	model.AnonAccess = types.StringValue(anonAccess)

	return diags
}
