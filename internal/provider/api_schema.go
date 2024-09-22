// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/axatol/kinde-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type APIResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Audience types.String `tfsdk:"audience"`
}

func expandAPIResourceModel(model APIResourceModel) *kinde.API {
	return &kinde.API{
		ID:       model.ID.ValueString(),
		Name:     model.Name.ValueString(),
		Audience: model.Audience.ValueString(),
	}
}

func flattenAPIResource(resource *kinde.API) APIResourceModel {
	return APIResourceModel{
		ID:       types.StringValue(resource.ID),
		Name:     types.StringValue(resource.Name),
		Audience: types.StringValue(resource.Audience),
	}
}

type APIDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Audience types.String `tfsdk:"audience"`
}

func expandAPIDataSourceModel(model APIDataSourceModel) *kinde.API {
	return &kinde.API{
		ID:       model.ID.ValueString(),
		Name:     model.Name.ValueString(),
		Audience: model.Audience.ValueString(),
	}
}

func flattenAPIDataSource(resource *kinde.API) APIDataSourceModel {
	return APIDataSourceModel{
		ID:       types.StringValue(resource.ID),
		Name:     types.StringValue(resource.Name),
		Audience: types.StringValue(resource.Audience),
	}
}
