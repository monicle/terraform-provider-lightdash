// Copyright 2023 Ubie, inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource              = &personalAccessTokenResource{}
	_ resource.ResourceWithConfigure = &personalAccessTokenResource{}
)

func NewPersonalAccessTokenResource() resource.Resource {
	return &personalAccessTokenResource{}
}

// personalAccessTokenResource defines the resource implementation.
type personalAccessTokenResource struct {
	client *api.Client
}

// personalAccessTokenResourceModel describes the resource data model.
type personalAccessTokenResourceModel struct {
	ID          types.String `tfsdk:"id"`
	TokenUUID   types.String `tfsdk:"token_uuid"`
	Description types.String `tfsdk:"description"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
	Token       types.String `tfsdk:"token"`
}

func (r *personalAccessTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_personal_access_token"
}

func (r *personalAccessTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/resources/resource_lightdash_personal_access_token.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Manages a Lightdash personal access token",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource identifier. It is computed as `personal-access-tokens/<token_uuid>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the personal access token.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the personal access token.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The expiration date of the personal access token in ISO 8601 format (e.g., '2024-12-31T23:59:59Z'). If not set, the token will not expire.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the personal access token was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The personal access token value. This is only available after creation and cannot be retrieved later.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *personalAccessTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *personalAccessTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan personalAccessTokenResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare the request
	createRequest := &models.CreatePersonalAccessToken{
		Description:   plan.Description.ValueString(),
		AutoGenerated: false,
	}

	// Set expires_at if provided
	if !plan.ExpiresAt.IsNull() && !plan.ExpiresAt.IsUnknown() {
		expiresAt := plan.ExpiresAt.ValueString()
		createRequest.ExpiresAt = &expiresAt
	}

	// Create the personal access token
	tflog.Info(ctx, fmt.Sprintf("Creating personal access token with description: %s", plan.Description.ValueString()))
	createdToken, err := r.client.CreatePersonalAccessTokenV1(createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating personal access token",
			"Could not create personal access token, unexpected error: "+err.Error(),
		)
		return
	}

	// Assign the plan values to the state
	stateId := getPersonalAccessTokenResourceId(createdToken.UUID)
	plan.ID = types.StringValue(stateId)
	plan.TokenUUID = types.StringValue(createdToken.UUID)
	plan.Description = types.StringValue(createdToken.Description)
	plan.CreatedAt = types.StringValue(createdToken.CreatedAt)
	plan.Token = types.StringValue(createdToken.Token)

	// Set expires_at from response
	if createdToken.ExpiresAt != nil {
		plan.ExpiresAt = types.StringValue(*createdToken.ExpiresAt)
	} else {
		plan.ExpiresAt = types.StringNull()
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *personalAccessTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state personalAccessTokenResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the token UUID from state
	tokenUuid := state.TokenUUID.ValueString()

	// List all personal access tokens to find the current one
	tokens, err := r.client.ListPersonalAccessTokensV1()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading personal access token",
			"Could not read personal access token ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Find the token with matching UUID
	var foundToken *models.PersonalAccessToken
	for _, token := range tokens {
		if token.UUID == tokenUuid {
			foundToken = &token
			break
		}
	}

	// If token not found, remove from state
	if foundToken == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with fetched values (keep token as-is since it's not returned by list)
	state.Description = types.StringValue(foundToken.Description)
	state.CreatedAt = types.StringValue(foundToken.CreatedAt)

	if foundToken.ExpiresAt != nil {
		state.ExpiresAt = types.StringValue(*foundToken.ExpiresAt)
	} else {
		state.ExpiresAt = types.StringNull()
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *personalAccessTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Personal access tokens cannot be updated, they must be recreated
	// This is handled by the RequiresReplace plan modifier on the description and expires_at attributes
	resp.Diagnostics.AddError(
		"Update not supported",
		"Personal access tokens cannot be updated. Changes require recreation of the resource.",
	)
}

func (r *personalAccessTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state personalAccessTokenResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the personal access token
	tokenUuid := state.TokenUUID.ValueString()
	tflog.Info(ctx, fmt.Sprintf("Deleting personal access token %s", tokenUuid))
	err := r.client.DeletePersonalAccessTokenV1(tokenUuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting personal access token",
			"Could not delete personal access token, unexpected error: "+err.Error(),
		)
		return
	}
}

func getPersonalAccessTokenResourceId(tokenUuid string) string {
	return fmt.Sprintf("personal-access-tokens/%s", tokenUuid)
}
