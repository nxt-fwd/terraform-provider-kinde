// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/nxt-fwd/kinde-go"
	"github.com/nxt-fwd/kinde-go/api/permissions"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var (
	_ resource.Resource                = &PermissionResource{}
	_ resource.ResourceWithImportState = &PermissionResource{}
)

func NewPermissionResource() resource.Resource {
	return &PermissionResource{}
}

type PermissionResource struct {
	client *permissions.Client
}

func (r *PermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission"
}

func (r *PermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Permissions represent individual access rights that can be assigned to roles. See [documentation](https://docs.kinde.com/kinde-apis/management/#tag/permissions) for more details.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the permission",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the permission",
				Required:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Key identifier of the permission",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the permission",
				Optional:            true,
			},
		},
	}
}

func (r *PermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client.Permissions
}

func (r *PermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PermissionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createParams := expandPermissionCreateParams(plan)
	permission, err := r.client.Create(ctx, createParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Permission",
			fmt.Sprintf("Could not create permission: %s", err),
		)
		return
	}

	// After creation, search for the permission to get its full details
	searchParams := permissions.SearchParams{
		Name: plan.Name.ValueString(),
		Key:  plan.Key.ValueString(),
	}

	permission, err = r.client.Search(ctx, searchParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Permission",
			fmt.Sprintf("Could not read created permission: %s", err),
		)
		return
	}

	state := flattenPermissionResource(permission)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *PermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PermissionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	searchParams := permissions.SearchParams{
		Name: state.Name.ValueString(),
		Key:  state.Key.ValueString(),
	}

	permission, err := r.client.Search(ctx, searchParams)
	if err != nil {
		if err == permissions.ErrPermissionNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Permission",
			fmt.Sprintf("Could not read permission with name %s and key %s: %s", state.Name.ValueString(), state.Key.ValueString(), err),
		)
		return
	}

	state = flattenPermissionResource(permission)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *PermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PermissionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateParams := expandPermissionUpdateParams(plan)
	err := r.client.Update(ctx, plan.ID.ValueString(), updateParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Permission",
			fmt.Sprintf("Could not update permission ID %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	// After update, search for the permission to get its latest state
	searchParams := permissions.SearchParams{
		Name: plan.Name.ValueString(),
		Key:  plan.Key.ValueString(),
	}

	permission, err := r.client.Search(ctx, searchParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Permission",
			fmt.Sprintf("Could not read updated permission: %s", err),
		)
		return
	}

	state := flattenPermissionResource(permission)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *PermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PermissionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Permission",
			fmt.Sprintf("Could not delete permission ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}
}

func (r *PermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Search for the permission by ID
	perms, err := r.client.List(ctx, permissions.ListParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Permission",
			"Could not list permissions: "+err.Error(),
		)
		return
	}

	// Find the permission with the matching ID
	var permission *permissions.Permission
	for _, p := range perms {
		if p.ID == req.ID {
			perm := p // Create a new variable to avoid issues with loop variable capture
			permission = &perm
			break
		}
	}

	if permission == nil {
		resp.Diagnostics.AddError(
			"Error Reading Permission",
			"Could not find permission with ID "+req.ID,
		)
		return
	}

	// Set the state
	state := flattenPermissionResource(permission)
	resp.State.Set(ctx, &state)
}