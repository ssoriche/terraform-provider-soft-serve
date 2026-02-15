package resource

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ssoriche/terraform-provider-soft-serve/internal/ssh"
)

var (
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

type UserResource struct {
	client *ssh.Client
}

type UserResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Username   types.String `tfsdk:"username"`
	Admin      types.Bool   `tfsdk:"admin"`
	PublicKeys types.Set    `tfsdk:"public_keys"`
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Soft Serve user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "User identifier (same as username).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"admin": schema.BoolAttribute{
				Description: "Whether the user is an admin.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"public_keys": schema.SetAttribute{
				Description: "Set of SSH public keys for the user.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := plan.Username.ValueString()

	var keys []string
	if !plan.PublicKeys.IsNull() && !plan.PublicKeys.IsUnknown() {
		resp.Diagnostics.Append(plan.PublicKeys.ElementsAs(ctx, &keys, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	opts := ssh.UserCreateOpts{
		Admin:      plan.Admin.ValueBool(),
		PublicKeys: keys,
	}

	if err := r.client.UserCreate(username, opts); err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readUserState(ctx, username, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readUserState(ctx, state.Username.ValueString(), &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := plan.Username.ValueString()

	// Update admin status
	if !plan.Admin.Equal(state.Admin) {
		if err := r.client.UserSetAdmin(username, plan.Admin.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Error updating admin status", err.Error())
			return
		}
	}

	// Update public keys
	if !plan.PublicKeys.Equal(state.PublicKeys) {
		var planKeys, stateKeys []string
		if !plan.PublicKeys.IsNull() {
			resp.Diagnostics.Append(plan.PublicKeys.ElementsAs(ctx, &planKeys, false)...)
		}
		if !state.PublicKeys.IsNull() {
			resp.Diagnostics.Append(state.PublicKeys.ElementsAs(ctx, &stateKeys, false)...)
		}
		if resp.Diagnostics.HasError() {
			return
		}

		planSet := toStringSet(planKeys)
		stateSet := toStringSet(stateKeys)

		// Remove keys no longer in plan
		for key := range stateSet {
			if _, ok := planSet[key]; !ok {
				if err := r.client.UserRemovePublicKey(username, key); err != nil {
					resp.Diagnostics.AddError("Error removing public key", err.Error())
					return
				}
			}
		}

		// Add new keys
		for key := range planSet {
			if _, ok := stateSet[key]; !ok {
				if err := r.client.UserAddPublicKey(username, key); err != nil {
					resp.Diagnostics.AddError("Error adding public key", err.Error())
					return
				}
			}
		}
	}

	resp.Diagnostics.Append(r.readUserState(ctx, username, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UserDelete(state.Username.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var model UserResourceModel
	model.Username = types.StringValue(req.ID)

	resp.Diagnostics.Append(r.readUserState(ctx, req.ID, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *UserResource) readUserState(ctx context.Context, username string, model *UserResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	info, err := r.client.UserInfo(username)
	if err != nil {
		diags.AddError("Error reading user", err.Error())
		return diags
	}

	model.ID = types.StringValue(username)
	model.Username = types.StringValue(info.Username)
	model.Admin = types.BoolValue(info.Admin)

	if len(info.PublicKeys) > 0 {
		// Sort for deterministic state
		sorted := make([]string, len(info.PublicKeys))
		copy(sorted, info.PublicKeys)
		sort.Strings(sorted)

		keyValues := make([]types.String, len(sorted))
		for i, k := range sorted {
			keyValues[i] = types.StringValue(k)
		}
		keySet, d := types.SetValueFrom(ctx, types.StringType, sorted)
		diags.Append(d...)
		model.PublicKeys = keySet
	} else if !model.PublicKeys.IsNull() {
		// Preserve null vs empty distinction: if plan had keys but server has none,
		// set to empty set; if plan was null, keep null
		keySet, d := types.SetValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		model.PublicKeys = keySet
	}

	return diags
}

func toStringSet(s []string) map[string]struct{} {
	m := make(map[string]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
