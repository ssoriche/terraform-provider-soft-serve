package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &SoftServeProvider{}

type SoftServeProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SoftServeProvider{
			version: version,
		}
	}
}

func (p *SoftServeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "soft_serve"
	resp.Version = p.version
}

func (p *SoftServeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
}

func (p *SoftServeProvider) Configure(_ context.Context, _ provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *SoftServeProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

func (p *SoftServeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
