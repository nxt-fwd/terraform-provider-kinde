// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nxt-fwd/kinde-go"
	"github.com/nxt-fwd/kinde-go/api/roles"
)

var (
	_ resource.Resource                = &RoleResource{}
	_ resource.ResourceWithImportState = &RoleResource{}
)

func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

type RoleResource struct {
	client *roles.Client
}

func (r *RoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *RoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Roles represent collections of permissions that can be assigned to users. See [documentation](https://docs.kinde.com/kinde-apis/management/#tag/roles) for more details.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the role",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the role",
				Required:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Key identifier of the role",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the role. This field is required because the Kinde API does not properly handle unsetting or empty descriptions once they are set. To maintain consistent behavior and prevent state drift, we require a description for all roles.",
				Required:            true,
			},
			"permissions": schema.SetAttribute{
				MarkdownDescription: "List of permission IDs associated with this role",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *RoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kinde.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *kinde.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client.Roles
}

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RoleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create role with basic details first
	createParams := expandRoleCreateParams(plan)
	role, err := r.client.Create(ctx, createParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Role",
			fmt.Sprintf("Could not create role: %s", err),
		)
		return
	}

	// Get the complete role data
	role, err = r.client.Get(ctx, role.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Role",
			fmt.Sprintf("Could not read created role: %s", err),
		)
		return
	}

	// Update permissions if specified
	if !plan.Permissions.IsNull() {
		var permissions []string
		diags = plan.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Sort permissions for consistent ordering
		sorted := make([]string, len(permissions))
		copy(sorted, permissions)
		sort.Strings(sorted)

		permissionItems := make([]roles.UpdatePermissionItem, len(sorted))
		for i, p := range sorted {
			permissionItems[i] = roles.UpdatePermissionItem{
				ID: p,
			}
		}

		updatePermParams := roles.UpdatePermissionsParams{
			Permissions: permissionItems,
		}

		_, err = r.client.UpdatePermissions(ctx, role.ID, updatePermParams)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Role Permissions",
				fmt.Sprintf("Could not set permissions for role: %s", err),
			)
			return
		}

		// Get the updated role to ensure we have all fields and permissions
		role, err = r.client.Get(ctx, role.ID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Updated Role",
				fmt.Sprintf("Could not read updated role: %s", err),
			)
			return
		}
	}

	state, err := flattenRoleResource(ctx, role, role.Permissions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting Role State",
			fmt.Sprintf("Could not set role state: %s", err),
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Helper function to sort permissions without modifying original
func sortPermissions(permissions []string) []string {
	sorted := make([]string, len(permissions))
	copy(sorted, permissions)
	sort.Strings(sorted)
	return sorted
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.client.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Role",
			fmt.Sprintf("Could not read role ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	// Sort permissions without modifying original
	sortedPerms := sortPermissions(role.Permissions)
	
	// If no permissions are returned, explicitly set to null
	var permissionsSet types.Set
	if len(sortedPerms) > 0 {
		permissionsSet, diags = types.SetValueFrom(ctx, types.StringType, sortedPerms)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		permissionsSet = types.SetNull(types.StringType)
	}

	state = RoleResourceModel{
		ID:          types.StringValue(role.ID),
		Name:        types.StringValue(role.Name),
		Key:         types.StringValue(role.Key),
		Description: types.StringValue(role.Description),
		Permissions: permissionsSet,
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state RoleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for comparison
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// First update role details
	updateParams := expandRoleUpdateParams(plan)
	_, err := r.client.Update(ctx, plan.ID.ValueString(), updateParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Role",
			fmt.Sprintf("Could not update role ID %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	// Handle permissions update if the field is set in the plan
	var planPerms []string
	if !plan.Permissions.IsNull() {
		diags = plan.Permissions.ElementsAs(ctx, &planPerms, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Get current permissions from state
	var statePerms []string
	if !state.Permissions.IsNull() {
		diags = state.Permissions.ElementsAs(ctx, &statePerms, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// First remove permissions that are in state but not in plan
	for _, statePerm := range statePerms {
		found := false
		for _, planPerm := range planPerms {
			if statePerm == planPerm {
				found = true
				break
			}
		}
		if !found {
			err = r.client.RemovePermission(ctx, plan.ID.ValueString(), statePerm)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Removing Permission",
					fmt.Sprintf("Could not remove permission %s from role %s: %s", statePerm, plan.ID.ValueString(), err),
				)
				return
			}
		}
	}

	// Then add any new permissions that are in plan but not in state
	var permsToAdd []string
	for _, planPerm := range planPerms {
		found := false
		for _, statePerm := range statePerms {
			if planPerm == statePerm {
				found = true
				break
			}
		}
		if !found {
			permsToAdd = append(permsToAdd, planPerm)
		}
	}

	if len(permsToAdd) > 0 {
		// Sort permissions for consistent ordering
		sortedPerms := sortPermissions(permsToAdd)
		permissionItems := make([]roles.UpdatePermissionItem, len(sortedPerms))
		for i, p := range sortedPerms {
			permissionItems[i] = roles.UpdatePermissionItem{
				ID: p,
			}
		}

		updatePermParams := roles.UpdatePermissionsParams{
			Permissions: permissionItems,
		}

		_, err = r.client.UpdatePermissions(ctx, plan.ID.ValueString(), updatePermParams)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Adding Permissions",
				fmt.Sprintf("Could not add permissions to role: %s", err),
			)
			return
		}
	}

	// Get the updated role to ensure we have all fields and permissions
	role, err := r.client.Get(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Role",
			fmt.Sprintf("Could not read updated role: %s", err),
		)
		return
	}

	// Sort permissions without modifying original before setting state
	sortedPerms := sortPermissions(role.Permissions)
	
	// If no permissions are returned, explicitly set to null
	var permissionsSet types.Set
	if len(sortedPerms) > 0 {
		permissionsSet, diags = types.SetValueFrom(ctx, types.StringType, sortedPerms)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		permissionsSet = types.SetNull(types.StringType)
	}

	state = RoleResourceModel{
		ID:          types.StringValue(role.ID),
		Name:        types.StringValue(role.Name),
		Key:         types.StringValue(role.Key),
		Description: types.StringValue(role.Description),
		Permissions: permissionsSet,
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Role",
			fmt.Sprintf("Could not delete role ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	role, err := r.client.Get(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Kinde Role",
			"Could not read Kinde role ID "+req.ID+": "+err.Error(),
		)
		return
	}

	// Sort the role's permissions for consistent ordering
	sortedPermissions := sortStringSlice(role.Permissions)

	state, err := flattenRoleResource(ctx, role, sortedPermissions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting Role State",
			"Could not set role state: "+err.Error(),
		)
		return
	}

	resp.State.Set(ctx, &state)
}
