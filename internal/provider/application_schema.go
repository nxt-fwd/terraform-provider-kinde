// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/axatol/kinde-go/api/applications"
	"github.com/axatol/terraform-provider-kinde/internal/serde"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ApplicationResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	LoginURI     types.String `tfsdk:"login_uri"`
	HomepageURI  types.String `tfsdk:"homepage_uri"`
	LogoutURIs   types.List   `tfsdk:"logout_uris"`
	RedirectURIs types.List   `tfsdk:"redirect_uris"`
}

func expandApplicationResourceModel(data ApplicationResourceModel) applications.Application {
	return applications.Application{
		ID:           data.ID.ValueString(),
		Name:         data.Name.ValueString(),
		Type:         applications.Type(data.Type.ValueString()),
		ClientID:     data.ClientID.ValueString(),
		ClientSecret: data.ClientSecret.ValueString(),
	}
}

func expandApplicationUpdateResourceModel(ctx context.Context, data ApplicationResourceModel) (applications.UpdateParams, error) {
	var logoutURIs []string
	if !data.LogoutURIs.IsNull() {
		data.LogoutURIs.ElementsAs(ctx, &logoutURIs, false)
	}

	var redirectURIs []string
	if !data.RedirectURIs.IsNull() {
		data.RedirectURIs.ElementsAs(ctx, &redirectURIs, false)
	}

	return applications.UpdateParams{
		Name:         data.Name.ValueString(),
		LoginURI:     data.LoginURI.ValueString(),
		HomepageURI:  data.HomepageURI.ValueString(),
		LogoutURIs:   logoutURIs,
		RedirectURIs: redirectURIs,
	}, nil
}

func flattenApplicationResource(ctx context.Context, resource *applications.Application, params applications.UpdateParams) (ApplicationResourceModel, diag.Diagnostics) {
	model := ApplicationResourceModel{
		ID:           types.StringValue(resource.ID),
		Name:         types.StringValue(resource.Name),
		Type:         types.StringValue(string(resource.Type)),
		ClientID:     types.StringValue(resource.ClientID),
		ClientSecret: types.StringValue(resource.ClientSecret),
		LoginURI:     types.StringValue(params.LoginURI),
		HomepageURI:  types.StringValue(params.HomepageURI),
	}

	var diags, nestedDiags diag.Diagnostics

	model.LogoutURIs, nestedDiags = serde.FlattenStringList(ctx, params.LogoutURIs)
	diags.Append(nestedDiags...)

	model.RedirectURIs, nestedDiags = serde.FlattenStringList(ctx, params.RedirectURIs)
	diags.Append(nestedDiags...)

	return model, diags
}

type ApplicationDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func expandApplicationDataSourceModel(model ApplicationDataSourceModel) *applications.Application {
	return &applications.Application{
		ID:           model.ID.ValueString(),
		Name:         model.Name.ValueString(),
		Type:         applications.Type(model.Type.ValueString()),
		ClientID:     model.ClientID.ValueString(),
		ClientSecret: model.ClientSecret.ValueString(),
	}
}

func flattenApplicationDataSource(resource *applications.Application) ApplicationDataSourceModel {
	return ApplicationDataSourceModel{
		ID:           types.StringValue(resource.ID),
		Name:         types.StringValue(resource.Name),
		Type:         types.StringValue(string(resource.Type)),
		ClientID:     types.StringValue(resource.ClientID),
		ClientSecret: types.StringValue(resource.ClientSecret),
	}
}
