package provider

import (
	"context"
	"fmt"

	"github.com/axatol/kinde-go"
	"github.com/axatol/kinde-go/api/users"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

var (
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client *users.Client
}

type UserResourceModel struct {
	ID              types.String `tfsdk:"id"`
	FirstName       types.String `tfsdk:"first_name"`
	LastName        types.String `tfsdk:"last_name"`
	IsSuspended     types.Bool   `tfsdk:"is_suspended"`
	OrganizationCode types.String `tfsdk:"organization_code"`
	CreatedOn       types.String `tfsdk:"created_on"`
	UpdatedOn       types.String `tfsdk:"updated_on"`
	Identities      types.List   `tfsdk:"identities"`
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Kinde user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"first_name": schema.StringAttribute{
				Description: "The first name of the user.",
				Optional:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "The last name of the user.",
				Optional:    true,
			},
			"is_suspended": schema.BoolAttribute{
				Description: "Whether the user is suspended.",
				Optional:    true,
				Computed:    true,
			},
			"organization_code": schema.StringAttribute{
				Description: "The code of the organization the user belongs to.",
				Optional:    true,
			},
			"created_on": schema.StringAttribute{
				Description: "The timestamp when the user was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_on": schema.StringAttribute{
				Description: "The timestamp when the user was last updated.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identities": schema.ListNestedAttribute{
				Description: "Identities for the user (email, username, phone, etc.).",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The type of identity (email, username, phone, enterprise, social).",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "The value of the identity.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client.Users
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Starting user creation")

	var plan UserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one email identity is provided
	var identities []struct {
		Type  string `tfsdk:"type"`
		Value string `tfsdk:"value"`
	}
	diags = plan.Identities.ElementsAs(ctx, &identities, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasEmail := false
	var createIdentities []users.Identity
	for _, identity := range identities {
		if identity.Type == string(users.IdentityTypeEmail) {
			hasEmail = true
		}
		details := make(map[string]string)
		switch identity.Type {
		case string(users.IdentityTypeEmail):
			details["email"] = identity.Value
		case string(users.IdentityTypeUsername):
			details["username"] = identity.Value
		case string(users.IdentityTypePhone):
			details["phone"] = identity.Value
		}
		createIdentities = append(createIdentities, users.Identity{
			Type:    identity.Type,
			Details: details,
		})
	}

	if !hasEmail {
		resp.Diagnostics.AddError(
			"Missing Email Identity",
			"At least one email identity must be provided for the user.",
		)
		return
	}

	createParams := users.CreateParams{
		Profile: users.Profile{
			GivenName:  plan.FirstName.ValueString(),
			FamilyName: plan.LastName.ValueString(),
		},
		OrgCode:    plan.OrganizationCode.ValueString(),
		Identities: createIdentities,
	}

	user, err := r.client.Create(ctx, createParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating User",
			fmt.Sprintf("Could not create user: %s", err),
		)
		return
	}

	// Set computed fields
	plan.ID = types.StringValue(user.ID)
	plan.CreatedOn = types.StringValue(user.CreatedOn.String())
	plan.UpdatedOn = types.StringValue(user.UpdatedOn.String())

	// Handle is_suspended
	isSuspended := false
	if !plan.IsSuspended.IsNull() {
		isSuspended = plan.IsSuspended.ValueBool()
	}

	// If is_suspended was set or is true, update the user
	if isSuspended {
		updateParams := users.UpdateParams{
			IsSuspended: &isSuspended,
		}
		
		user, err = r.client.Update(ctx, user.ID, updateParams)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating User",
				fmt.Sprintf("Could not update user suspension status: %s", err),
			)
			return
		}
	}

	// Get the final state of the user
	user, err = r.client.Get(ctx, user.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created User",
			fmt.Sprintf("Could not read created user ID %s: %s", user.ID, err),
		)
		return
	}

	plan.IsSuspended = types.BoolValue(user.IsSuspended)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading User",
			fmt.Sprintf("Could not read user ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	// Get user identities
	identities, err := r.client.GetIdentities(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading User Identities",
			fmt.Sprintf("Could not read identities for user ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	// Convert identities to Terraform state
	var tfIdentities []struct {
		Type  string `tfsdk:"type"`
		Value string `tfsdk:"value"`
	}

	for _, identity := range identities {
		tfIdentities = append(tfIdentities, struct {
			Type  string `tfsdk:"type"`
			Value string `tfsdk:"value"`
		}{
			Type:  identity.Type,
			Value: identity.Name,
		})
	}

	identitiesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type":  types.StringType,
			"value": types.StringType,
		},
	}, tfIdentities)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set all fields from API response
	state.ID = types.StringValue(user.ID)
	state.FirstName = types.StringValue(user.FirstName)
	state.LastName = types.StringValue(user.LastName)
	state.IsSuspended = types.BoolValue(user.IsSuspended)
	state.CreatedOn = types.StringValue(user.CreatedOn.String())
	state.UpdatedOn = types.StringValue(user.UpdatedOn.String())
	state.Identities = identitiesList

	// Preserve organization_code from state as it's not returned by the API
	if !state.OrganizationCode.IsNull() {
		state.OrganizationCode = state.OrganizationCode
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state UserResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one email identity is provided
	var identities []struct {
		Type  string `tfsdk:"type"`
		Value string `tfsdk:"value"`
	}
	diags = plan.Identities.ElementsAs(ctx, &identities, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasEmail := false
	for _, identity := range identities {
		if identity.Type == string(users.IdentityTypeEmail) {
			hasEmail = true
			break
		}
	}

	if !hasEmail {
		resp.Diagnostics.AddError(
			"Missing Email Identity",
			"At least one email identity must be provided for the user.",
		)
		return
	}

	// Note: Identities can only be added, not updated or removed
	var stateIdentities []struct {
		Type  string `tfsdk:"type"`
		Value string `tfsdk:"value"`
	}
	diags = state.Identities.ElementsAs(ctx, &stateIdentities, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add any new identities
	existingIdentities := make(map[string]bool)
	for _, identity := range stateIdentities {
		key := identity.Type + ":" + identity.Value
		existingIdentities[key] = true
	}

	for _, identity := range identities {
		key := identity.Type + ":" + identity.Value
		if !existingIdentities[key] {
			addIdentityParams := users.AddIdentityParams{
				Type:  users.IdentityType(identity.Type),
				Value: identity.Value,
			}

			_, err := r.client.AddIdentity(ctx, plan.ID.ValueString(), addIdentityParams)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Adding User Identity",
					fmt.Sprintf("Could not add identity to user %s: %s", plan.ID.ValueString(), err),
				)
				return
			}
		}
	}

	isSuspended := plan.IsSuspended.ValueBool()
	updateParams := users.UpdateParams{
		GivenName:   plan.FirstName.ValueString(),
		FamilyName:  plan.LastName.ValueString(),
		IsSuspended: &isSuspended,
	}

	user, err := r.client.Update(ctx, plan.ID.ValueString(), updateParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating User",
			fmt.Sprintf("Could not update user ID %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	// After update, get the full user details to ensure we have the latest state
	user, err = r.client.Get(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated User",
			fmt.Sprintf("Could not read updated user ID %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	plan.FirstName = types.StringValue(user.FirstName)
	plan.LastName = types.StringValue(user.LastName)
	plan.IsSuspended = types.BoolValue(user.IsSuspended)
	plan.UpdatedOn = types.StringValue(user.UpdatedOn.String())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting User",
			fmt.Sprintf("Could not delete user ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
} 