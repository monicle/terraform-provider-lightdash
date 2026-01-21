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

package models

// DbtProjectType represents the type of dbt project connection
type DbtProjectType string

const (
	DbtProjectTypeGithub DbtProjectType = "github"
	DbtProjectTypeGitlab DbtProjectType = "gitlab"
	DbtProjectTypeDbt    DbtProjectType = "dbt"
)

// DbtGithubProjectConfig represents GitHub dbt project configuration
type DbtGithubProjectConfig struct {
	Type                 DbtProjectType `json:"type"`
	AuthorizationMethod  string         `json:"authorization_method"` // "personal_access_token" or "installation_id"
	PersonalAccessToken  *string        `json:"personal_access_token,omitempty"`
	InstallationID       *string        `json:"installation_id,omitempty"`
	Repository           string         `json:"repository"`
	Branch               string         `json:"branch"`
	ProjectSubPath       string         `json:"project_sub_path"`
	HostDomain           *string        `json:"host_domain,omitempty"`
	Target               *string        `json:"target,omitempty"`
	Environment          []interface{}  `json:"environment,omitempty"`
	Selector             *string        `json:"selector,omitempty"`
}

// Project represents a Lightdash project
type Project struct {
	OrganizationUUID                    string                  `json:"organizationUuid"`
	ProjectUUID                         string                  `json:"projectUuid"`
	Name                                string                  `json:"name"`
	Type                                ProjectType             `json:"type"`
	DbtConnection                       *DbtGithubProjectConfig `json:"dbtConnection,omitempty"`
	DbtVersion                          string                  `json:"dbtVersion"`
	OrganizationWarehouseCredentialsUUID *string                `json:"organizationWarehouseCredentialsUuid,omitempty"`
	WarehouseConnection                 *WarehouseCredentials   `json:"warehouseConnection,omitempty"`
	UpstreamProjectUUID                 *string                 `json:"upstreamProjectUuid,omitempty"`
	PinnedListUUID                      *string                 `json:"pinnedListUuid,omitempty"`
	SchedulerTimezone                   *string                 `json:"schedulerTimezone,omitempty"`
}

// CreateProject represents the request body for creating a project
type CreateProject struct {
	Name                                     string                  `json:"name"`
	Type                                     ProjectType             `json:"type"`
	DbtConnection                            *DbtGithubProjectConfig `json:"dbtConnection"`
	DbtVersion                               string                  `json:"dbtVersion"`
	OrganizationWarehouseCredentialsUUID     *string                 `json:"organizationWarehouseCredentialsUuid,omitempty"`
	WarehouseConnection                      *WarehouseCredentials   `json:"warehouseConnection,omitempty"`
	UpstreamProjectUUID                      *string                 `json:"upstreamProjectUuid,omitempty"`
	CopyWarehouseConnectionFromUpstreamProject *bool                 `json:"copyWarehouseConnectionFromUpstreamProject,omitempty"`
}

// UpdateProject represents the request body for updating a project
type UpdateProject struct {
	Name                                 string                  `json:"name"`
	DbtConnection                        *DbtGithubProjectConfig `json:"dbtConnection"`
	DbtVersion                           string                  `json:"dbtVersion"`
	OrganizationWarehouseCredentialsUUID *string                 `json:"organizationWarehouseCredentialsUuid,omitempty"`
	WarehouseConnection                  *WarehouseCredentials   `json:"warehouseConnection,omitempty"`
}
