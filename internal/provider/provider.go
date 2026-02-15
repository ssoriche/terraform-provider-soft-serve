package provider

import (
	"context"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ssoriche/terraform-provider-soft-serve/internal/ssh"

	softserveresource "github.com/ssoriche/terraform-provider-soft-serve/internal/resource"
)

var _ provider.Provider = &SoftServeProvider{}

type SoftServeProvider struct {
	version string
}

type SoftServeProviderModel struct {
	Host           types.String `tfsdk:"host"`
	Port           types.Int64  `tfsdk:"port"`
	Username       types.String `tfsdk:"username"`
	PrivateKeyPath types.String `tfsdk:"private_key_path"`
	IdentityFile   types.String `tfsdk:"identity_file"`
	UseAgent       types.Bool   `tfsdk:"use_agent"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SoftServeProvider{
			version: version,
		}
	}
}

func (p *SoftServeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "softserve"
	resp.Version = p.version
}

func (p *SoftServeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Charm Soft Serve git server resources via SSH.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Soft Serve SSH host. Can also be set with SOFT_SERVE_HOST. Defaults to localhost.",
				Optional:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Soft Serve SSH port. Can also be set with SOFT_SERVE_PORT. Defaults to 23231.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "SSH username. Can also be set with SOFT_SERVE_USER. Defaults to current OS user.",
				Optional:    true,
			},
			"private_key_path": schema.StringAttribute{
				Description: "Path to SSH private key file. SOFT_SERVE_PRIVATE_KEY env var (key contents) takes precedence.",
				Optional:    true,
			},
			"identity_file": schema.StringAttribute{
				Description: "Path to SSH public key file used to select which agent key to offer (like OpenSSH IdentityFile). Can also be set with SOFT_SERVE_IDENTITY_FILE.",
				Optional:    true,
			},
			"use_agent": schema.BoolAttribute{
				Description: "Whether to use SSH agent for authentication. Can also be set with SOFT_SERVE_USE_AGENT. Defaults to true.",
				Optional:    true,
			},
		},
	}
}

func (p *SoftServeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config SoftServeProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve host
	host := "localhost"
	if envHost := os.Getenv("SOFT_SERVE_HOST"); envHost != "" {
		host = envHost
	}
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	// Resolve port
	port := 23231
	if envPort := os.Getenv("SOFT_SERVE_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}
	if !config.Port.IsNull() {
		port = int(config.Port.ValueInt64())
	}

	// Resolve username
	username := ""
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	if envUser := os.Getenv("SOFT_SERVE_USER"); envUser != "" {
		username = envUser
	}
	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	// Resolve private key
	privateKey := os.Getenv("SOFT_SERVE_PRIVATE_KEY")

	privateKeyPath := ""
	if !config.PrivateKeyPath.IsNull() {
		privateKeyPath = config.PrivateKeyPath.ValueString()
		if strings.HasPrefix(privateKeyPath, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				privateKeyPath = home + privateKeyPath[1:]
			}
		}
	}

	// Resolve identity_file
	identityFile := os.Getenv("SOFT_SERVE_IDENTITY_FILE")
	if !config.IdentityFile.IsNull() {
		identityFile = config.IdentityFile.ValueString()
	}
	if strings.HasPrefix(identityFile, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			identityFile = home + identityFile[1:]
		}
	}

	// Resolve use_agent
	useAgent := true
	if envAgent := os.Getenv("SOFT_SERVE_USE_AGENT"); envAgent != "" {
		useAgent = envAgent == "true" || envAgent == "1"
	}
	if !config.UseAgent.IsNull() {
		useAgent = config.UseAgent.ValueBool()
	}

	// Create SSH client
	client, err := ssh.NewClient(ssh.ClientConfig{
		Host:           host,
		Port:           port,
		Username:       username,
		PrivateKey:     privateKey,
		PrivateKeyPath: privateKeyPath,
		IdentityFile:   identityFile,
		UseAgent:       useAgent,
	})
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unable to create Soft Serve SSH client",
			err.Error(),
		)
		return
	}

	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *SoftServeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		softserveresource.NewRepositoryResource,
		softserveresource.NewUserResource,
		softserveresource.NewRepositoryCollaboratorResource,
		softserveresource.NewServerSettingsResource,
	}
}

func (p *SoftServeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
