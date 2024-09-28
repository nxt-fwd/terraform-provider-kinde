// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/axatol/kinde-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = (*ApplicationResource)(nil)
var _ resource.ResourceWithImportState = (*ApplicationResource)(nil)

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

type ApplicationResource struct {
	client *kinde.Client
}

func (r *ApplicationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ApplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Applications facilitates the interface for users to authenticate against. See [documentation](https://docs.kinde.com/build/applications/about-applications/) for more details.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the application",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the application. Currently, there is no way to change this via the management application.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the application",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Client id of the application",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Client secret of the application",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"login_uri": schema.StringAttribute{
				MarkdownDescription: "Login uri of the application",
				Optional:            true,
				Computed:            true,
			},
			"homepage_uri": schema.StringAttribute{
				MarkdownDescription: "Homepage uri of the application",
				Optional:            true,
				Computed:            true,
			},
			"language_key": schema.StringAttribute{
				MarkdownDescription: "Language key of the application",
				Optional:            true,
				Computed:            true,
			},
			"logout_uris": schema.ListAttribute{
				MarkdownDescription: "Logout uris of the application",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"redirect_uris": schema.ListAttribute{
				MarkdownDescription: "Redirect uris of the application",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *ApplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationResourceModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...); resp.Diagnostics.HasError() {
		return
	}

	resource := expandApplicationResourceModel(plan)

	createParams := kinde.CreateApplicationParams{
		Name: resource.Name,
		Type: resource.Type,
	}

	tflog.Debug(ctx, "Creating application", map[string]any{"params": createParams})

	createdResource, err := r.client.CreateApplication(ctx, createParams)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create application", err.Error())
		return
	}

	tflog.Debug(ctx, "Created application", map[string]any{"resource": resource})

	// refetching because create only returns the id
	newResource, err := r.client.GetApplication(ctx, createdResource.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get application", err.Error())
		return
	}

	updateParams, diags := expandApplicationUpdateResourceModel(ctx, plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating application", map[string]any{"params": updateParams})

	// updating because the api does not return these fields
	if err := r.client.UpdateApplication(ctx, newResource.ID, updateParams); err != nil {
		resp.Diagnostics.AddError("Failed to update application", err.Error())
		return
	}

	tflog.Debug(ctx, "Updated application", map[string]any{"id": newResource.ID})

	data, diags := flattenApplicationResource(ctx, newResource, updateParams)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationResourceModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &state)...); resp.Diagnostics.HasError() {
		return
	}

	resource := expandApplicationResourceModel(state)

	tflog.Debug(ctx, "Reading application", map[string]any{"id": resource.ID})

	resource, err := r.client.GetApplication(ctx, resource.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get application", err.Error())
		return
	}

	tflog.Debug(ctx, "Read application", map[string]any{"resource": resource})

	updateParams, diags := expandApplicationUpdateResourceModel(ctx, state)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	data, diags := flattenApplicationResource(ctx, resource, updateParams)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationResourceModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...); resp.Diagnostics.HasError() {
		return
	}

	resource := expandApplicationResourceModel(plan)

	tflog.Debug(ctx, "Updating application", map[string]any{"id": resource.ID})

	params, diags := expandApplicationUpdateResourceModel(ctx, plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateApplication(ctx, resource.ID, params); err != nil {
		resp.Diagnostics.AddError("Failed to update application", err.Error())
		return
	}

	tflog.Debug(ctx, "Updated application", map[string]any{"id": resource.ID})

	updateParams, diags := expandApplicationUpdateResourceModel(ctx, plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	data, diags := flattenApplicationResource(ctx, resource, updateParams)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, data)
}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationResourceModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &state)...); resp.Diagnostics.HasError() {
		return
	}

	resource := expandApplicationResourceModel(state)

	tflog.Debug(ctx, "Deleting application", map[string]any{"id": resource.ID})

	if err := r.client.DeleteApplication(ctx, resource.ID); err != nil {
		resp.Diagnostics.AddError("Failed to delete application", err.Error())
		return
	}

	tflog.Debug(ctx, "Deleted application", map[string]any{"id": resource.ID})
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing application", map[string]any{"id": req.ID})

	resource, err := r.client.GetApplication(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import application", err.Error())
		return
	}

	tflog.Debug(ctx, "Imported application", map[string]any{"resource": resource})

	data, diags := flattenApplicationResource(ctx, resource, kinde.UpdateApplicationParams{})
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
