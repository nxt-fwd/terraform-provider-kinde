// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/nxt-fwd/kinde-go/api/permissions"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PermissionResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Key         types.String `tfsdk:"key"`
	Description types.String `tfsdk:"description"`
}

func expandPermissionResourceModel(d PermissionResourceModel) *permissions.Permission {
	return &permissions.Permission{
		ID:          d.ID.ValueString(),
		Name:        d.Name.ValueString(),
		Key:         d.Key.ValueString(),
		Description: d.Description.ValueString(),
	}
}

func expandPermissionCreateParams(d PermissionResourceModel) permissions.CreateParams {
	return permissions.CreateParams{
		Name:        d.Name.ValueString(),
		Key:         d.Key.ValueString(),
		Description: d.Description.ValueString(),
	}
}

func expandPermissionUpdateParams(d PermissionResourceModel) permissions.UpdateParams {
	return permissions.UpdateParams{
		Name:        d.Name.ValueString(),
		Key:         d.Key.ValueString(),
		Description: d.Description.ValueString(),
	}
}

func flattenPermissionResource(permission *permissions.Permission) PermissionResourceModel {
	return PermissionResourceModel{
		ID:          types.StringValue(permission.ID),
		Name:        types.StringValue(permission.Name),
		Key:         types.StringValue(permission.Key),
		Description: types.StringValue(permission.Description),
	}
}

type PermissionDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Key         types.String `tfsdk:"key"`
	Description types.String `tfsdk:"description"`
}

func expandPermissionDataSourceModel(model PermissionDataSourceModel) *permissions.Permission {
	return &permissions.Permission{
		ID:          model.ID.ValueString(),
		Name:        model.Name.ValueString(),
		Key:         model.Key.ValueString(),
		Description: model.Description.ValueString(),
	}
}

func flattenPermissionDataSource(permission *permissions.Permission) PermissionDataSourceModel {
	return PermissionDataSourceModel{
		ID:          types.StringValue(permission.ID),
		Name:        types.StringValue(permission.Name),
		Key:         types.StringValue(permission.Key),
		Description: types.StringValue(permission.Description),
	}
} 
