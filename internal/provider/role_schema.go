// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/nxt-fwd/kinde-go/api/roles"
	"github.com/nxt-fwd/terraform-provider-kinde/internal/serde"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Key         types.String `tfsdk:"key"`
	Description types.String `tfsdk:"description"`
	Permissions types.List   `tfsdk:"permissions"`
}

func expandRoleResourceModel(data RoleResourceModel) roles.Role {
	return roles.Role{
		ID:          data.ID.ValueString(),
		Name:        data.Name.ValueString(),
		Key:         data.Key.ValueString(),
		Description: data.Description.ValueString(),
	}
}

func expandRoleCreateParams(data RoleResourceModel) roles.CreateParams {
	return roles.CreateParams{
		Name:        data.Name.ValueString(),
		Key:         data.Key.ValueString(),
		Description: data.Description.ValueString(),
	}
}

func expandRoleUpdateParams(data RoleResourceModel) roles.UpdateParams {
	return roles.UpdateParams{
		Name:        data.Name.ValueString(),
		Key:         data.Key.ValueString(),
		Description: data.Description.ValueString(),
	}
}

func flattenRoleResource(ctx context.Context, resource *roles.Role, permissions []string) (RoleResourceModel, error) {
	permissionsList, diags := serde.FlattenStringList(ctx, permissions)
	if diags.HasError() {
		return RoleResourceModel{}, fmt.Errorf("failed to flatten permissions: %v", diags)
	}

	return RoleResourceModel{
		ID:          types.StringValue(resource.ID),
		Name:        types.StringValue(resource.Name),
		Key:         types.StringValue(resource.Key),
		Description: types.StringValue(resource.Description),
		Permissions: permissionsList,
	}, nil
}

type RoleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Key         types.String `tfsdk:"key"`
	Description types.String `tfsdk:"description"`
	Permissions types.List   `tfsdk:"permissions"`
}

func expandRoleDataSourceModel(model RoleDataSourceModel) *roles.Role {
	return &roles.Role{
		ID:          model.ID.ValueString(),
		Name:        model.Name.ValueString(),
		Key:         model.Key.ValueString(),
		Description: model.Description.ValueString(),
	}
}

func flattenRoleDataSource(ctx context.Context, resource *roles.Role, permissions []string) (RoleDataSourceModel, error) {
	permissionsList, diags := serde.FlattenStringList(ctx, permissions)
	if diags.HasError() {
		return RoleDataSourceModel{}, fmt.Errorf("failed to flatten permissions: %v", diags)
	}

	return RoleDataSourceModel{
		ID:          types.StringValue(resource.ID),
		Name:        types.StringValue(resource.Name),
		Key:         types.StringValue(resource.Key),
		Description: types.StringValue(resource.Description),
		Permissions: permissionsList,
	}, nil
} 
