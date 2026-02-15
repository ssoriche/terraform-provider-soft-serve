package resource

import (
	"context"
	"fmt"

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
	_ resource.Resource                = &RepositoryResource{}
	_ resource.ResourceWithImportState = &RepositoryResource{}
)

type RepositoryResource struct {
	client *ssh.Client
}

type RepositoryResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ProjectName types.String `tfsdk:"project_name"`
	Private     types.Bool   `tfsdk:"private"`
	Hidden      types.Bool   `tfsdk:"hidden"`
}

func NewRepositoryResource() resource.Resource {
	return &RepositoryResource{}
}

func (r *RepositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *RepositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Soft Serve git repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Repository identifier (same as name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Repository name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Repository description.",
				Optional:    true,
				Computed:    true,
			},
			"project_name": schema.StringAttribute{
				Description: "Project name for the repository.",
				Optional:    true,
				Computed:    true,
			},
			"private": schema.BoolAttribute{
				Description: "Whether the repository is private.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"hidden": schema.BoolAttribute{
				Description: "Whether the repository is hidden.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *RepositoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RepositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	opts := ssh.RepoCreateOpts{
		Private: plan.Private.ValueBool(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		opts.Description = plan.Description.ValueString()
	}
	if !plan.ProjectName.IsNull() && !plan.ProjectName.IsUnknown() {
		opts.ProjectName = plan.ProjectName.ValueString()
	}

	if err := r.client.RepoCreate(name, opts); err != nil {
		resp.Diagnostics.AddError("Error creating repository", err.Error())
		return
	}

	// Set hidden after creation if needed
	if plan.Hidden.ValueBool() {
		if err := r.client.RepoSetHidden(name, true); err != nil {
			resp.Diagnostics.AddError("Error setting repository hidden", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(r.readRepoState(name, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readRepoState(state.Name.ValueString(), &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RepositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state RepositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	if !plan.Description.Equal(state.Description) {
		desc := ""
		if !plan.Description.IsNull() {
			desc = plan.Description.ValueString()
		}
		if err := r.client.RepoSetDescription(name, desc); err != nil {
			resp.Diagnostics.AddError("Error updating description", err.Error())
			return
		}
	}

	if !plan.ProjectName.Equal(state.ProjectName) {
		pn := ""
		if !plan.ProjectName.IsNull() {
			pn = plan.ProjectName.ValueString()
		}
		if err := r.client.RepoSetProjectName(name, pn); err != nil {
			resp.Diagnostics.AddError("Error updating project name", err.Error())
			return
		}
	}

	if !plan.Private.Equal(state.Private) {
		if err := r.client.RepoSetPrivate(name, plan.Private.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Error updating private", err.Error())
			return
		}
	}

	if !plan.Hidden.Equal(state.Hidden) {
		if err := r.client.RepoSetHidden(name, plan.Hidden.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Error updating hidden", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(r.readRepoState(name, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RepoDelete(state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting repository", err.Error())
	}
}

func (r *RepositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var model RepositoryResourceModel
	model.Name = types.StringValue(req.ID)

	resp.Diagnostics.Append(r.readRepoState(req.ID, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *RepositoryResource) readRepoState(name string, model *RepositoryResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	info, err := r.client.RepoInfo(name)
	if err != nil {
		diags.AddError("Error reading repository", err.Error())
		return diags
	}

	model.ID = types.StringValue(name)
	model.Name = types.StringValue(info.Repository)
	model.Description = types.StringValue(info.Description)
	model.ProjectName = types.StringValue(info.ProjectName)
	model.Private = types.BoolValue(info.Private)
	model.Hidden = types.BoolValue(info.Hidden)

	return diags
}
