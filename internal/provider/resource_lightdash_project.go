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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource              = &projectResource{}
	_ resource.ResourceWithConfigure = &projectResource{}
)

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

// projectResource defines the resource implementation.
type projectResource struct {
	client *api.Client
}

// dbtConnectionModel describes the dbt connection nested object
type dbtConnectionModel struct {
	Type                types.String `tfsdk:"type"`
	AuthorizationMethod types.String `tfsdk:"authorization_method"`
	PersonalAccessToken types.String `tfsdk:"personal_access_token"`
	Repository          types.String `tfsdk:"repository"`
	Branch              types.String `tfsdk:"branch"`
	ProjectSubPath      types.String `tfsdk:"project_sub_path"`
	HostDomain          types.String `tfsdk:"host_domain"`
	Target              types.String `tfsdk:"target"`
}

// warehouseConnectionModel describes the warehouse connection nested object
type warehouseConnectionModel struct {
	Type                 types.String `tfsdk:"type"`
	Project              types.String `tfsdk:"project"`
	Dataset              types.String `tfsdk:"dataset"`
	KeyfileContents      types.String `tfsdk:"keyfile_contents"`
	AuthenticationType   types.String `tfsdk:"authentication_type"`
	Location             types.String `tfsdk:"location"`
	TimeoutSeconds       types.Int64  `tfsdk:"timeout_seconds"`
	MaximumBytesBilled   types.Int64  `tfsdk:"maximum_bytes_billed"`
	Priority             types.String `tfsdk:"priority"`
	Retries              types.Int64  `tfsdk:"retries"`
	StartOfWeek          types.Int64  `tfsdk:"start_of_week"`
}

// projectResourceModel describes the resource data model.
type projectResourceModel struct {
	ID                                   types.String              `tfsdk:"id"`
	OrganizationUUID                     types.String              `tfsdk:"organization_uuid"`
	ProjectUUID                          types.String              `tfsdk:"project_uuid"`
	Name                                 types.String              `tfsdk:"name"`
	Type                                 types.String              `tfsdk:"type"`
	DbtVersion                           types.String              `tfsdk:"dbt_version"`
	DbtConnection                        *dbtConnectionModel       `tfsdk:"dbt_connection"`
	OrganizationWarehouseCredentialsUUID types.String              `tfsdk:"organization_warehouse_credentials_uuid"`
	WarehouseConnection                  *warehouseConnectionModel `tfsdk:"warehouse_connection"`
	UpstreamProjectUUID                  types.String              `tfsdk:"upstream_project_uuid"`
}

func (r *projectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Lightdash project with GitHub dbt connection.",
		Description:         "Manages a Lightdash project",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource identifier. It is computed as `organizations/<organization_uuid>/projects/<project_uuid>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the Lightdash organization.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the project.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the project. Valid values are 'DEFAULT' or 'PREVIEW'.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dbt_version": schema.StringAttribute{
				MarkdownDescription: "The dbt version to use (e.g., 'v1.8', 'v1.9', 'v1.10').",
				Required:            true,
			},
			"dbt_connection": schema.SingleNestedAttribute{
				MarkdownDescription: "The dbt connection configuration for GitHub.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "The type of dbt connection. Currently only 'github' is supported.",
						Required:            true,
					},
					"authorization_method": schema.StringAttribute{
						MarkdownDescription: "The authorization method. Valid values are 'personal_access_token' or 'installation_id'.",
						Required:            true,
					},
					"personal_access_token": schema.StringAttribute{
						MarkdownDescription: "The GitHub personal access token. Required when authorization_method is 'personal_access_token'.",
						Optional:            true,
						Sensitive:           true,
					},
					"repository": schema.StringAttribute{
						MarkdownDescription: "The GitHub repository in the format 'owner/repo'.",
						Required:            true,
					},
					"branch": schema.StringAttribute{
						MarkdownDescription: "The Git branch to use.",
						Required:            true,
					},
					"project_sub_path": schema.StringAttribute{
						MarkdownDescription: "The subdirectory path within the repository where the dbt project is located (e.g., '/' or '/dbt').",
						Required:            true,
					},
					"host_domain": schema.StringAttribute{
						MarkdownDescription: "The GitHub host domain. Optional, for GitHub Enterprise.",
						Optional:            true,
					},
					"target": schema.StringAttribute{
						MarkdownDescription: "The dbt target to use.",
						Optional:            true,
					},
				},
			},
			"organization_warehouse_credentials_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the organization warehouse credentials to use. Mutually exclusive with warehouse_connection.",
				Optional:            true,
			},
			"warehouse_connection": schema.SingleNestedAttribute{
				MarkdownDescription: "The warehouse connection configuration. Mutually exclusive with organization_warehouse_credentials_uuid.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "The type of warehouse. Currently only 'bigquery' is supported.",
						Required:            true,
					},
					"project": schema.StringAttribute{
						MarkdownDescription: "The GCP project ID for BigQuery.",
						Required:            true,
					},
					"dataset": schema.StringAttribute{
						MarkdownDescription: "The BigQuery dataset name.",
						Required:            true,
					},
					"keyfile_contents": schema.StringAttribute{
						MarkdownDescription: "The contents of the service account key file in JSON format.",
						Required:            true,
						Sensitive:           true,
					},
					"authentication_type": schema.StringAttribute{
						MarkdownDescription: "The authentication type for BigQuery. Valid values: 'sso', 'private_key', 'adc'. Optional.",
						Optional:            true,
					},
					"location": schema.StringAttribute{
						MarkdownDescription: "The location of the BigQuery dataset.",
						Optional:            true,
					},
					"timeout_seconds": schema.Int64Attribute{
						MarkdownDescription: "The timeout for BigQuery queries in seconds.",
						Optional:            true,
					},
					"maximum_bytes_billed": schema.Int64Attribute{
						MarkdownDescription: "The maximum bytes that can be billed for a query.",
						Optional:            true,
					},
					"priority": schema.StringAttribute{
						MarkdownDescription: "The priority for BigQuery jobs ('interactive' or 'batch').",
						Optional:            true,
					},
					"retries": schema.Int64Attribute{
						MarkdownDescription: "The number of retries for failed queries.",
						Optional:            true,
					},
					"start_of_week": schema.Int64Attribute{
						MarkdownDescription: "The start of week (0 = Sunday, 1 = Monday, etc.).",
						Optional:            true,
					},
				},
			},
			"upstream_project_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the upstream project for PREVIEW type projects.",
				Optional:            true,
			},
		},
	}
}

func (r *projectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build dbt connection config
	var dbtConnection *models.DbtGithubProjectConfig
	if plan.DbtConnection != nil {
		dbtConnection = &models.DbtGithubProjectConfig{
			Type:                models.DbtProjectTypeGithub,
			AuthorizationMethod: plan.DbtConnection.AuthorizationMethod.ValueString(),
			Repository:          plan.DbtConnection.Repository.ValueString(),
			Branch:              plan.DbtConnection.Branch.ValueString(),
			ProjectSubPath:      plan.DbtConnection.ProjectSubPath.ValueString(),
		}

		if !plan.DbtConnection.PersonalAccessToken.IsNull() {
			token := plan.DbtConnection.PersonalAccessToken.ValueString()
			dbtConnection.PersonalAccessToken = &token
		}

		if !plan.DbtConnection.HostDomain.IsNull() {
			domain := plan.DbtConnection.HostDomain.ValueString()
			dbtConnection.HostDomain = &domain
		}

		if !plan.DbtConnection.Target.IsNull() {
			target := plan.DbtConnection.Target.ValueString()
			dbtConnection.Target = &target
		}
	}

	// Build create project request
	createReq := &models.CreateProject{
		Name:          plan.Name.ValueString(),
		Type:          models.ProjectType(plan.Type.ValueString()),
		DbtVersion:    plan.DbtVersion.ValueString(),
		DbtConnection: dbtConnection,
	}

	if !plan.OrganizationWarehouseCredentialsUUID.IsNull() {
		uuid := plan.OrganizationWarehouseCredentialsUUID.ValueString()
		createReq.OrganizationWarehouseCredentialsUUID = &uuid
	}

	// Build warehouse connection config
	if plan.WarehouseConnection != nil {
		// Parse keyfile contents JSON
		var keyfileMap map[string]interface{}
		if err := json.Unmarshal([]byte(plan.WarehouseConnection.KeyfileContents.ValueString()), &keyfileMap); err != nil {
			resp.Diagnostics.AddError(
				"Error parsing keyfile_contents",
				"Could not parse keyfile_contents as JSON: "+err.Error(),
			)
			return
		}

		warehouseConn := &models.BigQueryCredentials{
			Type:            plan.WarehouseConnection.Type.ValueString(),
			Project:         plan.WarehouseConnection.Project.ValueString(),
			KeyfileContents: keyfileMap,
		}

		if !plan.WarehouseConnection.Dataset.IsNull() {
			dataset := plan.WarehouseConnection.Dataset.ValueString()
			warehouseConn.Dataset = &dataset
		}

		if !plan.WarehouseConnection.AuthenticationType.IsNull() {
			authType := plan.WarehouseConnection.AuthenticationType.ValueString()
			warehouseConn.AuthenticationType = &authType
		}

		if !plan.WarehouseConnection.Location.IsNull() {
			location := plan.WarehouseConnection.Location.ValueString()
			warehouseConn.Location = &location
		}

		if !plan.WarehouseConnection.TimeoutSeconds.IsNull() {
			timeout := int(plan.WarehouseConnection.TimeoutSeconds.ValueInt64())
			warehouseConn.TimeoutSeconds = &timeout
		}

		if !plan.WarehouseConnection.MaximumBytesBilled.IsNull() {
			maxBytes := plan.WarehouseConnection.MaximumBytesBilled.ValueInt64()
			warehouseConn.MaximumBytesBilled = &maxBytes
		}

		if !plan.WarehouseConnection.Priority.IsNull() {
			priority := strings.ToLower(plan.WarehouseConnection.Priority.ValueString())
			warehouseConn.Priority = &priority
		}

		if !plan.WarehouseConnection.Retries.IsNull() {
			retries := int(plan.WarehouseConnection.Retries.ValueInt64())
			warehouseConn.Retries = &retries
		}

		if !plan.WarehouseConnection.StartOfWeek.IsNull() {
			startOfWeek := int(plan.WarehouseConnection.StartOfWeek.ValueInt64())
			warehouseConn.StartOfWeek = &startOfWeek
		}

		createReq.WarehouseConnection = warehouseConn
	}

	if !plan.UpstreamProjectUUID.IsNull() {
		upstreamUUID := plan.UpstreamProjectUUID.ValueString()
		createReq.UpstreamProjectUUID = &upstreamUUID
	}

	// Create project
	createdProject, err := r.client.CreateProjectV1(createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project",
			"Could not create project, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state
	organizationUUID := plan.OrganizationUUID.ValueString()
	stateId := getProjectResourceId(organizationUUID, createdProject.ProjectUUID)
	plan.ID = types.StringValue(stateId)
	plan.ProjectUUID = types.StringValue(createdProject.ProjectUUID)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get project
	project, err := r.client.GetProjectV1(state.ProjectUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading project",
			"Could not read project ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state
	state.Name = types.StringValue(project.ProjectName)
	state.Type = types.StringValue(project.ProjectType)
	state.OrganizationUUID = types.StringValue(project.OrganizationUUID)

	if project.DbtVersion != "" {
		state.DbtVersion = types.StringValue(project.DbtVersion)
	}

	if project.OrganizationWarehouseCredentialsUUID != nil {
		state.OrganizationWarehouseCredentialsUUID = types.StringValue(*project.OrganizationWarehouseCredentialsUUID)
	} else {
		state.OrganizationWarehouseCredentialsUUID = types.StringNull()
	}

	if project.UpstreamProjectUUID != nil {
		state.UpstreamProjectUUID = types.StringValue(*project.UpstreamProjectUUID)
	} else {
		state.UpstreamProjectUUID = types.StringNull()
	}

	// Note: dbt connection credentials are not returned in the API response for security reasons
	// We keep the existing values from the state

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Projects are immutable - any change requires replacement
	// This method exists only to satisfy the resource.Resource interface
	resp.Diagnostics.AddError(
		"Update not supported",
		"Lightdash projects are immutable. Any changes require destroying and recreating the resource.",
	)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Projects are not deleted via Terraform
	// This is a no-op to allow Terraform to remove the resource from state
	// The actual project remains in Lightdash and must be deleted manually via the UI or API
	resp.Diagnostics.AddWarning(
		"Project not deleted",
		"The Lightdash project was removed from Terraform state but still exists in Lightdash. You must manually delete it from the Lightdash UI if desired.",
	)
}

func getProjectResourceId(organizationUUID string, projectUUID string) string {
	return fmt.Sprintf("organizations/%s/projects/%s", organizationUUID, projectUUID)
}
