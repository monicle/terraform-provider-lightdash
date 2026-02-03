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
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &personalAccessTokensDataSource{}
	_ datasource.DataSourceWithConfigure = &personalAccessTokensDataSource{}
)

func NewPersonalAccessTokensDataSource() datasource.DataSource {
	return &personalAccessTokensDataSource{}
}

// personalAccessTokensDataSource defines the data source implementation.
type personalAccessTokensDataSource struct {
	client *api.Client
}

// personalAccessTokenModel describes the data source data model for a personal access token.
type personalAccessTokenModel struct {
	TokenUUID   types.String `tfsdk:"token_uuid"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
	RotatedAt   types.String `tfsdk:"rotated_at"`
	LastUsedAt  types.String `tfsdk:"last_used_at"`
}

// personalAccessTokensDataSourceModel describes the data source data model.
type personalAccessTokensDataSourceModel struct {
	ID     types.String               `tfsdk:"id"`
	Tokens []personalAccessTokenModel `tfsdk:"tokens"`
}

func (d *personalAccessTokensDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_personal_access_tokens"
}

func (d *personalAccessTokensDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownDescription, err := readMarkdownDescription(ctx, "internal/provider/docs/data_sources/data_source_lightdash_personal_access_tokens.md")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read markdown description",
			fmt.Sprintf("Unable to read schema markdown description file: %s", err.Error()),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: markdownDescription,
		Description:         "Lightdash personal access tokens data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The data source identifier. It is computed as `personal-access-tokens`.",
				Computed:            true,
			},
			"tokens": schema.ListNestedAttribute{
				MarkdownDescription: "A list of personal access tokens.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token_uuid": schema.StringAttribute{
							MarkdownDescription: "The UUID of the personal access token.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "The description of the personal access token.",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "The timestamp when the personal access token was created.",
							Computed:            true,
						},
						"expires_at": schema.StringAttribute{
							MarkdownDescription: "The expiration date of the personal access token.",
							Computed:            true,
						},
						"rotated_at": schema.StringAttribute{
							MarkdownDescription: "The timestamp when the personal access token was last rotated.",
							Computed:            true,
						},
						"last_used_at": schema.StringAttribute{
							MarkdownDescription: "The timestamp when the personal access token was last used.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *personalAccessTokensDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *personalAccessTokensDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state personalAccessTokensDataSourceModel

	// Retrieve the configuration
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all personal access tokens
	tokens, err := d.client.ListPersonalAccessTokensV1()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get personal access tokens",
			err.Error(),
		)
		return
	}

	// Convert to model
	fetchedTokens := []personalAccessTokenModel{}
	for _, token := range tokens {
		fetchedToken := personalAccessTokenModel{
			TokenUUID:   types.StringValue(token.UUID),
			Description: types.StringValue(token.Description),
			CreatedAt:   types.StringValue(token.CreatedAt),
		}

		// Handle nullable fields
		if token.ExpiresAt != nil {
			fetchedToken.ExpiresAt = types.StringValue(*token.ExpiresAt)
		} else {
			fetchedToken.ExpiresAt = types.StringNull()
		}

		if token.RotatedAt != nil {
			fetchedToken.RotatedAt = types.StringValue(*token.RotatedAt)
		} else {
			fetchedToken.RotatedAt = types.StringNull()
		}

		if token.LastUsedAt != nil {
			fetchedToken.LastUsedAt = types.StringValue(*token.LastUsedAt)
		} else {
			fetchedToken.LastUsedAt = types.StringNull()
		}

		fetchedTokens = append(fetchedTokens, fetchedToken)
	}

	// Sort the tokens by token UUID
	sort.Slice(fetchedTokens, func(i, j int) bool {
		return fetchedTokens[i].TokenUUID.ValueString() < fetchedTokens[j].TokenUUID.ValueString()
	})

	// Set resource ID
	state.ID = types.StringValue("personal-access-tokens")
	state.Tokens = fetchedTokens

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
