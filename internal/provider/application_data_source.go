// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/axatol/kinde-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = (*ApplicationDataSource)(nil)

func NewApplicationDataSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

type ApplicationDataSource struct {
	client *kinde.Client
}

func (d *ApplicationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *ApplicationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Applications facilitates the interface for users to authenticate against. See [documentation](https://docs.kinde.com/build/applications/about-applications/) for more details.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Application",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Application.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the application.",
				Computed:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Client id of the application.",
				Computed:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Client secret of the application. Not available for SPA type applications.",
				Computed:            true,
				Sensitive:           true,
			},
			// we can't get these from the api
			// login_uri
			// homepage_uri
			// language_key
			// logout_uris
			// redirect_uris
		},
	}
}

func (d *ApplicationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kinde.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *kinde.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ApplicationDataSourceModel
	if resp.Diagnostics.Append(req.Config.Get(ctx, &config)...); resp.Diagnostics.HasError() {
		return
	}

	resource := expandApplicationDataSourceModel(config)

	tflog.Debug(ctx, "Reading Application", map[string]any{"id": resource.ID})

	resource, err := d.client.GetApplication(ctx, resource.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Application", err.Error())
		return
	}

	tflog.Debug(ctx, "Read Application", map[string]any{"resource": resource})

	state := flattenApplicationDataSource(resource)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
