package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nxt-fwd/kinde-go"
	"github.com/nxt-fwd/kinde-go/api/users"
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
	ID               types.String `tfsdk:"id"`
	FirstName        types.String `tfsdk:"first_name"`
	LastName         types.String `tfsdk:"last_name"`
	IsSuspended      types.Bool   `tfsdk:"is_suspended"`
	OrganizationCode types.String `tfsdk:"organization_code"`
	CreatedOn        types.String `tfsdk:"created_on"`
	UpdatedOn        types.String `tfsdk:"updated_on"`
	Identities       types.List   `tfsdk:"identities"`
}

// Add custom plan modifier
type identitiesModifier struct{}

func (m identitiesModifier) Description(ctx context.Context) string {
	return "Preserves OAuth2 identities during plan phase."
}

func (m identitiesModifier) MarkdownDescription(ctx context.Context) string {
	return "Preserves OAuth2 identities during plan phase."
}

func (m identitiesModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// If there's no state, we have nothing to preserve
	if req.StateValue.IsNull() {
		return
	}

	// If the plan is unknown, we can't modify it
	if req.PlanValue.IsUnknown() {
		return
	}

	// Get current state identities
	var stateIdentities []struct {
		Type  string `tfsdk:"type"`
		Value string `tfsdk:"value"`
	}
	resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &stateIdentities, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get planned identities
	var plannedIdentities []struct {
		Type  string `tfsdk:"type"`
		Value string `tfsdk:"value"`
	}
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &plannedIdentities, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract OAuth2 identities from state
	var oauth2Identities []struct {
		Type  string `tfsdk:"type"`
		Value string `tfsdk:"value"`
	}
	for _, identity := range stateIdentities {
		if strings.HasPrefix(identity.Type, "oauth2:") {
			oauth2Identities = append(oauth2Identities, identity)
		}
	}

	// If there are no OAuth2 identities, no modification needed
	if len(oauth2Identities) == 0 {
		return
	}

	// Merge OAuth2 identities with planned identities
	mergedIdentities := append(plannedIdentities, oauth2Identities...)

	// Create new list value with merged identities
	newList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type":  types.StringType,
			"value": types.StringType,
		},
	}, mergedIdentities)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = newList
}

// Add custom plan modifier for is_suspended
type isSuspendedModifier struct{}

func (m isSuspendedModifier) Description(ctx context.Context) string {
	return "Ensures is_suspended remains null when not configured."
}

func (m isSuspendedModifier) MarkdownDescription(ctx context.Context) string {
	return "Ensures is_suspended remains null when not configured."
}

func (m isSuspendedModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// If the value is not in config, keep it null
	if req.ConfigValue.IsNull() {
		resp.PlanValue = types.BoolNull()
		return
	}
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
				Description: "The first name of the user. Empty strings are not supported and will be ignored by the API, preserving the existing value.",
				Optional:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "The last name of the user. Empty strings are not supported and will be ignored by the API, preserving the existing value.",
				Optional:    true,
			},
			"is_suspended": schema.BoolAttribute{
				Description: "Whether the user is suspended. When not configured, the value will remain null and won't be managed by Terraform.",
				Optional:    true,
				Computed:    false,
				PlanModifiers: []planmodifier.Bool{
					isSuspendedModifier{},
				},
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
				PlanModifiers: []planmodifier.List{
					identitiesModifier{},
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

	// Initialize empty profile
	profile := users.Profile{}

	// Only set first_name if it's not null
	if !plan.FirstName.IsNull() {
		profile.GivenName = plan.FirstName.ValueString()
	}

	// Only set last_name if it's not null
	if !plan.LastName.IsNull() {
		profile.FamilyName = plan.LastName.ValueString()
	}

	// Create user with profile and identities
	createParams := users.CreateParams{
		Profile:    profile,
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

	// Only update is_suspended if it's explicitly set in the configuration
	if !plan.IsSuspended.IsNull() {
		isSuspended := plan.IsSuspended.ValueBool()
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

	// Set all fields from API response
	plan.ID = types.StringValue(user.ID)

	// Handle first_name: if not in plan, keep it that way
	if !plan.FirstName.IsNull() {
		if user.FirstName == "" {
			plan.FirstName = types.StringValue("")
		} else {
			plan.FirstName = types.StringValue(user.FirstName)
		}
	}

	// Handle last_name: if not in plan, keep it that way
	if !plan.LastName.IsNull() {
		if user.LastName == "" {
			plan.LastName = types.StringValue("")
		} else {
			plan.LastName = types.StringValue(user.LastName)
		}
	}

	plan.CreatedOn = types.StringValue(user.CreatedOn.String())
	plan.UpdatedOn = types.StringValue(user.UpdatedOn.String())

	// Only set is_suspended in state if it was configured
	if !plan.IsSuspended.IsNull() {
		plan.IsSuspended = types.BoolValue(user.IsSuspended)
	}

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
		// Preserve the original identity type from state if it exists
		var originalType string
		if !state.Identities.IsNull() {
			var stateIdentities []struct {
				Type  string `tfsdk:"type"`
				Value string `tfsdk:"value"`
			}
			diags = state.Identities.ElementsAs(ctx, &stateIdentities, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Find matching identity in state
			for _, stateIdentity := range stateIdentities {
				if stateIdentity.Value == identity.Name {
					originalType = stateIdentity.Type
					break
				}
			}
		}

		// Use original type if found, otherwise use API type
		identityType := identity.Type
		if originalType != "" {
			identityType = originalType
		}

		tfIdentities = append(tfIdentities, struct {
			Type  string `tfsdk:"type"`
			Value string `tfsdk:"value"`
		}{
			Type:  identityType,
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

	// During import or if the fields were already in state, set them from the API response
	if !state.FirstName.IsNull() || req.State.Raw.IsNull() {
		state.FirstName = types.StringValue(user.FirstName)
	}
	if !state.LastName.IsNull() || req.State.Raw.IsNull() {
		state.LastName = types.StringValue(user.LastName)
	}

	// For Read operations, preserve the existing is_suspended state
	// If it was null before, keep it null
	if !state.IsSuspended.IsNull() {
		state.IsSuspended = types.BoolValue(user.IsSuspended)
	}

	state.CreatedOn = types.StringValue(user.CreatedOn.String())
	state.UpdatedOn = types.StringValue(user.UpdatedOn.String())
	state.Identities = identitiesList

	// Organization code is preserved from state as it's not returned by the API
	// No action needed as the value is already in state.OrganizationCode

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Always initialize empty update params
	updateParams := users.UpdateParams{}

	// Only set FirstName if it's not null and not empty
	if !plan.FirstName.IsNull() && plan.FirstName.ValueString() != "" {
		firstName := plan.FirstName.ValueString()
		updateParams.GivenName = firstName
	}

	// Only set LastName if it's not null and not empty
	if !plan.LastName.IsNull() && plan.LastName.ValueString() != "" {
		lastName := plan.LastName.ValueString()
		updateParams.FamilyName = lastName
	}

	// Only include is_suspended in update if it's configured
	if !plan.IsSuspended.IsNull() {
		isSuspended := plan.IsSuspended.ValueBool()
		updateParams.IsSuspended = &isSuspended
	}

	user, err := r.client.Update(ctx, plan.ID.ValueString(), updateParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating User",
			fmt.Sprintf("Could not update user ID %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	// Get current identities from the API to preserve OAuth2 identities
	currentIdentities, err := r.client.GetIdentities(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading User Identities",
			fmt.Sprintf("Could not read identities for user %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	// Extract OAuth2 identities to preserve
	var oauth2Identities []struct {
		Type  string `tfsdk:"type"`
		Value string `tfsdk:"value"`
	}
	for _, identity := range currentIdentities {
		if strings.HasPrefix(identity.Type, "oauth2:") {
			oauth2Identities = append(oauth2Identities, struct {
				Type  string `tfsdk:"type"`
				Value string `tfsdk:"value"`
			}{
				Type:  identity.Type,
				Value: identity.Name,
			})
		}
	}

	// Get planned identities
	var plannedIdentities []struct {
		Type  string `tfsdk:"type"`
		Value string `tfsdk:"value"`
	}
	diags := plan.Identities.ElementsAs(ctx, &plannedIdentities, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Merge OAuth2 identities with planned identities
	mergedIdentities := append(plannedIdentities, oauth2Identities...)

	// Validate that at least one email identity is provided
	hasEmail := false
	for _, identity := range mergedIdentities {
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

	// Get current state identities for comparison
	existingIdentities := make(map[string]bool)
	var stateIdentities []struct {
		Type  string `tfsdk:"type"`
		Value string `tfsdk:"value"`
	}
	diags = state.Identities.ElementsAs(ctx, &stateIdentities, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, identity := range stateIdentities {
		key := identity.Type + ":" + identity.Value
		existingIdentities[key] = true
	}

	for _, identity := range mergedIdentities {
		// Skip OAuth2 identities as they are managed externally
		if strings.HasPrefix(identity.Type, "oauth2:") {
			continue
		}

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

	// Get final state of the user
	user, err = r.client.Get(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated User",
			fmt.Sprintf("Could not read updated user %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	// Get final identities
	finalIdentities, err := r.client.GetIdentities(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading User Identities",
			fmt.Sprintf("Could not read identities for user %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	// Convert final identities to Terraform state format
	var tfIdentities []struct {
		Type  string `tfsdk:"type"`
		Value string `tfsdk:"value"`
	}
	for _, identity := range finalIdentities {
		tfIdentities = append(tfIdentities, struct {
			Type  string `tfsdk:"type"`
			Value string `tfsdk:"value"`
		}{
			Type:  identity.Type,
			Value: identity.Name,
		})
	}

	// Update plan with final state
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
	plan.Identities = identitiesList

	// Only set name fields in state if they were in the plan
	// This ensures that omitted fields stay omitted
	if !plan.FirstName.IsNull() {
		plan.FirstName = types.StringValue(user.FirstName)
	}
	if !plan.LastName.IsNull() {
		plan.LastName = types.StringValue(user.LastName)
	}

	// Only set is_suspended in state if it was configured
	if !plan.IsSuspended.IsNull() {
		plan.IsSuspended = types.BoolValue(user.IsSuspended)
	}

	plan.CreatedOn = types.StringValue(user.CreatedOn.String())
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

	if err := r.client.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting User",
			fmt.Sprintf("Could not delete user ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get the user by ID
	user, err := r.client.Get(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading User",
			fmt.Sprintf("Could not read user ID %s: %s", req.ID, err),
		)
		return
	}

	// Get user identities
	identities, err := r.client.GetIdentities(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading User Identities",
			fmt.Sprintf("Could not read identities for user %s: %s", req.ID, err),
		)
		return
	}

	// Convert identities to Terraform state format
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

	// Create identities list
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

	// Set all fields in state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), user.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("first_name"), user.FirstName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("last_name"), user.LastName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_on"), user.CreatedOn.String())...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("updated_on"), user.UpdatedOn.String())...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("identities"), identitiesList)...)
}
