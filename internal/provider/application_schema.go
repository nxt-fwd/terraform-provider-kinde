// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/axatol/kinde-go"
	"github.com/axatol/terraform-provider-kinde/internal/serde"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ApplicationResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	LanguageKey  types.String `tfsdk:"language_key"`
	LogoutURIs   types.List   `tfsdk:"logout_uris"`
	RedirectURIs types.List   `tfsdk:"redirect_uris"`
	LoginURI     types.String `tfsdk:"login_uri"`
	HomepageURI  types.String `tfsdk:"homepage_uri"`
}

func expandApplicationResourceModel(model ApplicationResourceModel) *kinde.Application {
	return &kinde.Application{
		ID:           model.ID.ValueString(),
		Name:         model.Name.ValueString(),
		Type:         kinde.ApplicationType(model.Type.ValueString()),
		ClientID:     model.ClientID.ValueString(),
		ClientSecret: model.ClientSecret.ValueString(),
	}
}

func expandApplicationUpdateResourceModel(ctx context.Context, model ApplicationResourceModel) (kinde.UpdateApplicationParams, diag.Diagnostics) {
	params := kinde.UpdateApplicationParams{
		Name:        model.Name.ValueString(),
		LanguageKey: model.LanguageKey.ValueString(),
		LoginURI:    model.LoginURI.ValueString(),
		HomepageURI: model.HomepageURI.ValueString(),
	}

	var diags, nestedDiags diag.Diagnostics

	params.LogoutURIs, nestedDiags = serde.ExpandStringList(ctx, model.LogoutURIs)
	diags.Append(nestedDiags...)

	params.RedirectURIs, nestedDiags = serde.ExpandStringList(ctx, model.RedirectURIs)
	diags.Append(nestedDiags...)

	tflog.Debug(ctx, "Expanded application params", map[string]any{
		"logout_uris":   params.LogoutURIs,
		"redirect_uris": params.RedirectURIs,
	})

	return params, diags
}

func flattenApplicationResource(ctx context.Context, resource *kinde.Application, params kinde.UpdateApplicationParams) (ApplicationResourceModel, diag.Diagnostics) {
	model := ApplicationResourceModel{
		ID:           types.StringValue(resource.ID),
		Name:         types.StringValue(resource.Name),
		Type:         types.StringValue(string(resource.Type)),
		ClientID:     types.StringValue(resource.ClientID),
		ClientSecret: types.StringValue(resource.ClientSecret),
		LanguageKey:  types.StringValue(params.LanguageKey),
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

func expandApplicationDataSourceModel(model ApplicationDataSourceModel) *kinde.Application {
	return &kinde.Application{
		ID:           model.ID.ValueString(),
		Name:         model.Name.ValueString(),
		Type:         kinde.ApplicationType(model.Type.ValueString()),
		ClientID:     model.ClientID.ValueString(),
		ClientSecret: model.ClientSecret.ValueString(),
	}
}

func flattenApplicationDataSource(resource *kinde.Application) ApplicationDataSourceModel {
	return ApplicationDataSourceModel{
		ID:           types.StringValue(resource.ID),
		Name:         types.StringValue(resource.Name),
		Type:         types.StringValue(string(resource.Type)),
		ClientID:     types.StringValue(resource.ClientID),
		ClientSecret: types.StringValue(resource.ClientSecret),
	}
}
