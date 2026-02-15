package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ssoriche/terraform-provider-soft-serve/internal/ssh"
)

var (
	_ resource.Resource                = &RepositoryCollaboratorResource{}
	_ resource.ResourceWithImportState = &RepositoryCollaboratorResource{}
)

type RepositoryCollaboratorResource struct {
	client *ssh.Client
}

type RepositoryCollaboratorResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Repository  types.String `tfsdk:"repository"`
	Username    types.String `tfsdk:"username"`
	AccessLevel types.String `tfsdk:"access_level"`
}

func NewRepositoryCollaboratorResource() resource.Resource {
	return &RepositoryCollaboratorResource{}
}

func (r *RepositoryCollaboratorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_collaborator"
}

func (r *RepositoryCollaboratorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a collaborator on a Soft Serve repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Collaborator identifier (repository/username).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"repository": schema.StringAttribute{
				Description: "Repository name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username of the collaborator.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"access_level": schema.StringAttribute{
				Description: "Access level: no-access, read-only, read-write, or admin-access.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("read-write"),
				Validators: []validator.String{
					stringvalidator.OneOf("no-access", "read-only", "read-write", "admin-access"),
				},
			},
		},
	}
}

func (r *RepositoryCollaboratorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RepositoryCollaboratorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepositoryCollaboratorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo := plan.Repository.ValueString()
	username := plan.Username.ValueString()
	accessLevel := plan.AccessLevel.ValueString()

	if err := r.client.CollabAdd(repo, username, accessLevel); err != nil {
		resp.Diagnostics.AddError("Error adding collaborator", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readCollabState(repo, username, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryCollaboratorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepositoryCollaboratorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readCollabState(state.Repository.ValueString(), state.Username.ValueString(), &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RepositoryCollaboratorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RepositoryCollaboratorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo := plan.Repository.ValueString()
	username := plan.Username.ValueString()
	accessLevel := plan.AccessLevel.ValueString()

	// collab add with a different access level updates the existing entry
	if err := r.client.CollabAdd(repo, username, accessLevel); err != nil {
		resp.Diagnostics.AddError("Error updating collaborator", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readCollabState(repo, username, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryCollaboratorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepositoryCollaboratorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.CollabRemove(state.Repository.ValueString(), state.Username.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error removing collaborator", err.Error())
	}
}

func (r *RepositoryCollaboratorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected format: repository/username, got: %s", req.ID))
		return
	}

	var model RepositoryCollaboratorResourceModel
	model.Repository = types.StringValue(parts[0])
	model.Username = types.StringValue(parts[1])

	resp.Diagnostics.Append(r.readCollabState(parts[0], parts[1], &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *RepositoryCollaboratorResource) readCollabState(repo, username string, model *RepositoryCollaboratorResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	collabs, err := r.client.CollabList(repo)
	if err != nil {
		diags.AddError("Error listing collaborators", err.Error())
		return diags
	}

	for _, c := range collabs {
		if c.Username == username {
			model.ID = types.StringValue(repo + "/" + username)
			model.Repository = types.StringValue(repo)
			model.Username = types.StringValue(username)
			accessLevel := c.AccessLevel
			if accessLevel == "" {
				accessLevel = "read-write"
			}
			model.AccessLevel = types.StringValue(accessLevel)
			return diags
		}
	}

	diags.AddError("Collaborator not found",
		fmt.Sprintf("User %q is not a collaborator on repository %q", username, repo))
	return diags
}
