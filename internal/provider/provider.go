// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/axatol/kinde-go"
	"github.com/axatol/kinde-go/api/users"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure KindeProvider satisfies various provider interfaces.
var _ provider.Provider = &KindeProvider{}

// KindeProvider defines the provider implementation.
type KindeProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// KindeProviderModel describes the provider data model.
type KindeProviderModel struct {
	Domain       types.String `tfsdk:"domain"`
	Audience     types.String `tfsdk:"audience"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func (p *KindeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kinde"
	resp.Version = p.version
}

func (p *KindeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "Kinde organisation domain, also set by KINDE_DOMAIN",
				Optional:            true,
			},
			"audience": schema.StringAttribute{
				MarkdownDescription: "Kinde M2M application audience, also set by KINDE_AUDIENCE",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Kinde M2M application client id, also set by KINDE_CLIENT_ID",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Kinde M2M application client secret, also set by KINDE_CLIENT_SECRET",
				Optional:            true,
			},
		},
	}
}

func (p *KindeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data KindeProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := kinde.NewClientOptions()

	if !data.Domain.IsNull() && !data.Domain.IsUnknown() {
		opts.WithDomain(data.Domain.ValueString())
	}

	if !data.Audience.IsNull() && !data.Domain.IsUnknown() {
		opts.WithAudience(data.Audience.ValueString())
	}

	if !data.ClientID.IsNull() && !data.Domain.IsUnknown() {
		opts.WithClientID(data.ClientID.ValueString())
	}

	if !data.ClientSecret.IsNull() && !data.Domain.IsUnknown() {
		opts.WithClientSecret(data.ClientSecret.ValueString())
	}

	client := &kinde.Client{}
	*client = kinde.New(ctx, opts)

	// Validate credentials by making a test API call
	_, err := client.Users.List(ctx, users.ListParams{PageSize: 1})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Kinde Client",
			fmt.Sprintf("Failed to authenticate with Kinde API: %v\n"+
				"Please verify your domain, client_id, client_secret, and audience are correct.", err),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *KindeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAPIResource,
		NewApplicationResource,
		NewOrganizationResource,
		NewUserResource,
	}
}

func (p *KindeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAPIDataSource,
		NewApplicationDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KindeProvider{
			version: version,
		}
	}
}
