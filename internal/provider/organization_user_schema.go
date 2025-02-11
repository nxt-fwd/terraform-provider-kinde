// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nxt-fwd/kinde-go/api/organizations"
)

type OrganizationUserResourceModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationCode types.String `tfsdk:"organization_code"`
	UserID           types.String `tfsdk:"user_id"`
	Roles            types.List   `tfsdk:"roles"`
	Permissions      types.List   `tfsdk:"permissions"`
}

//nolint:unused
func expandOrganizationUserModel(data OrganizationUserResourceModel) organizations.AddUser {
	var roles []string
	if !data.Roles.IsNull() {
		data.Roles.ElementsAs(context.TODO(), &roles, false)
	}

	var permissions []string
	if !data.Permissions.IsNull() {
		data.Permissions.ElementsAs(context.TODO(), &permissions, false)
	}

	return organizations.AddUser{
		ID:          data.UserID.ValueString(),
		Roles:       roles,
		Permissions: permissions,
	}
}

//nolint:unused
func expandOrganizationUserParams(data OrganizationUserResourceModel) organizations.AddUsersParams {
	return organizations.AddUsersParams{
		Users: []organizations.AddUser{
			expandOrganizationUserModel(data),
		},
	}
}
