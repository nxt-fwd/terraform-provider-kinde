// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/nxt-fwd/kinde-go"
	"github.com/nxt-fwd/kinde-go/api/applications"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ApplicationResource{}
	_ resource.ResourceWithImportState = &ApplicationResource{}
)

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

type ApplicationResource struct {
	client *applications.Client
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
				Description: "The login URI of the application.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"homepage_uri": schema.StringAttribute{
				Description: "The homepage URI of the application.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"logout_uris": schema.ListAttribute{
				Description: "The logout URIs of the application.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"redirect_uris": schema.ListAttribute{
				Description: "The redirect URIs of the application.",
				Optional:    true,
				ElementType: types.StringType,
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

	if client.Applications == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Applications Client",
			"Expected configured applications client. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = client.Applications
	tflog.Debug(ctx, "Application resource configured")
}

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Starting application creation")

	var plan ApplicationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Not Configured",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	tflog.Debug(ctx, "Creating application with params", map[string]interface{}{
		"name": plan.Name.ValueString(),
		"type": plan.Type.ValueString(),
	})

	createParams := applications.CreateParams{
		Name: plan.Name.ValueString(),
		Type: applications.Type(plan.Type.ValueString()),
	}

	app, err := r.client.Create(ctx, createParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Application",
			fmt.Sprintf("Could not create application: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Application created successfully", map[string]interface{}{
		"id": app.ID,
	})

	plan.ID = types.StringValue(app.ID)
	plan.ClientID = types.StringValue(app.ClientID)
	plan.ClientSecret = types.StringValue(app.ClientSecret)

	// Update the application with additional settings if provided
	var logoutURIs []string
	if !plan.LogoutURIs.IsNull() {
		diags = plan.LogoutURIs.ElementsAs(ctx, &logoutURIs, false)
		resp.Diagnostics.Append(diags...)
	}

	var redirectURIs []string
	if !plan.RedirectURIs.IsNull() {
		diags = plan.RedirectURIs.ElementsAs(ctx, &redirectURIs, false)
		resp.Diagnostics.Append(diags...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Only update if any of the optional fields are set
	if !plan.LoginURI.IsNull() || !plan.HomepageURI.IsNull() || len(logoutURIs) > 0 || len(redirectURIs) > 0 {
		tflog.Debug(ctx, "Updating application with additional settings", map[string]interface{}{
			"id":            app.ID,
			"has_login":     !plan.LoginURI.IsNull(),
			"has_homepage":  !plan.HomepageURI.IsNull(),
			"logout_count":  len(logoutURIs),
			"redirect_count": len(redirectURIs),
		})

		updateParams := applications.UpdateParams{
			LoginURI:     plan.LoginURI.ValueString(),
			HomepageURI:  plan.HomepageURI.ValueString(),
			LogoutURIs:   logoutURIs,
			RedirectURIs: redirectURIs,
		}

		err = r.client.Update(ctx, app.ID, updateParams)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Application",
				fmt.Sprintf("Could not update application ID %s: %s", app.ID, err),
			)
			return
		}

		tflog.Debug(ctx, "Application updated successfully")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Application creation completed")
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Application",
			fmt.Sprintf("Could not read application ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	// Set all values from application data
	state.ID = types.StringValue(app.ID)
	state.Name = types.StringValue(app.Name)
	state.Type = types.StringValue(string(app.Type))
	state.ClientID = types.StringValue(app.ClientID)
	state.ClientSecret = types.StringValue(app.ClientSecret)

	// Set optional values
	if app.LoginURI != "" {
		state.LoginURI = types.StringValue(app.LoginURI)
	} else {
		state.LoginURI = types.StringNull()
	}

	if app.HomepageURI != "" {
		state.HomepageURI = types.StringValue(app.HomepageURI)
	} else {
		state.HomepageURI = types.StringNull()
	}

	// Handle URI lists
	// Keep logout and redirect URIs from state as they are not returned by the API
	if state.LogoutURIs.IsNull() {
		state.LogoutURIs = types.ListNull(types.StringType)
	}
	if state.RedirectURIs.IsNull() {
		state.RedirectURIs = types.ListNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Helper function to convert string slice to []attr.Value
func listValuesFromStrings(strings []string) []attr.Value {
	values := make([]attr.Value, len(strings))
	for i, s := range strings {
		values[i] = types.StringValue(s)
	}
	return values
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var logoutURIs []string
	diags = plan.LogoutURIs.ElementsAs(ctx, &logoutURIs, false)
	resp.Diagnostics.Append(diags...)

	var redirectURIs []string
	diags = plan.RedirectURIs.ElementsAs(ctx, &redirectURIs, false)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateParams := applications.UpdateParams{
		Name:         plan.Name.ValueString(),
		LoginURI:     plan.LoginURI.ValueString(),
		HomepageURI:  plan.HomepageURI.ValueString(),
		LogoutURIs:   logoutURIs,
		RedirectURIs: redirectURIs,
	}

	err := r.client.Update(ctx, plan.ID.ValueString(), updateParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Application",
			fmt.Sprintf("Could not update application ID %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Application",
			fmt.Sprintf("Could not delete application ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
