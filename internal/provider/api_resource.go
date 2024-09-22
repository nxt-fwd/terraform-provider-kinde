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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = (*APIResource)(nil)
var _ resource.ResourceWithImportState = (*APIResource)(nil)

func NewAPIResource() resource.Resource {
	return &APIResource{}
}

type APIResource struct {
	client *kinde.Client
}

func (r *APIResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api"
}

func (r *APIResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "APIs represent the resource server to authorise against. See [documentation](https://docs.kinde.com/developer-tools/your-apis/register-manage-apis/) for more details.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the API",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the API. Currently, there is no way to change this via the management API.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"audience": schema.StringAttribute{
				MarkdownDescription: "Audience of the API",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *APIResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *APIResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan APIResourceModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...); resp.Diagnostics.HasError() {
		return
	}

	resource := expandAPIResourceModel(plan)
	params := kinde.CreateAPIParams{
		Name:     resource.Name,
		Audience: resource.Audience,
	}

	tflog.Debug(ctx, "Creating API", map[string]any{"params": params})

	resource, err := r.client.CreateAPI(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API", err.Error())
		return
	}

	// filling in other fields because the api does not return them
	resource.Name = params.Name
	resource.Audience = params.Audience

	tflog.Debug(ctx, "Created API", map[string]any{"resource": resource})

	data := flattenAPIResource(resource)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state APIResourceModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &state)...); resp.Diagnostics.HasError() {
		return
	}

	resource := expandAPIResourceModel(state)
	params := kinde.GetAPIParams{
		ID: resource.ID,
	}

	tflog.Debug(ctx, "Reading API", map[string]any{"params": params})

	resource, err := r.client.GetAPI(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get API", err.Error())
		return
	}

	tflog.Debug(ctx, "Read API", map[string]any{"resource": resource})

	data := flattenAPIResource(resource)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan APIResourceModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...); resp.Diagnostics.HasError() {
		return
	}

	// there is nothing to do here

	resp.State.Set(ctx, plan)
}

func (r *APIResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state APIResourceModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &state)...); resp.Diagnostics.HasError() {
		return
	}

	resource := expandAPIResourceModel(state)
	params := kinde.DeleteAPIParams{
		ID: resource.ID,
	}

	tflog.Debug(ctx, "Deleting API", map[string]any{"params": params})

	if err := r.client.DeleteAPI(ctx, params); err != nil {
		resp.Diagnostics.AddError("Failed to delete API", err.Error())
		return
	}

	tflog.Debug(ctx, "Deleted API", map[string]any{"id": resource.ID})
}

func (r *APIResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	params := kinde.GetAPIParams{
		ID: req.ID,
	}

	tflog.Debug(ctx, "Importing API", map[string]any{"params": params})

	resource, err := r.client.GetAPI(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import API", err.Error())
		return
	}

	tflog.Debug(ctx, "Imported API", map[string]any{"resource": resource})

	state := flattenAPIResource(resource)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
