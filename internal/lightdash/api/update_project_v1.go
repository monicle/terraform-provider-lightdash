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

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

type UpdateProjectV1Response struct {
	Results models.Project `json:"results,omitempty"`
	Status  string         `json:"status"`
}

func (c *Client) UpdateProjectV1(projectUuid string, project *models.UpdateProject) (*models.Project, error) {
	// Marshal the request body
	marshalled, err := json.Marshal(project)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the request
	path := fmt.Sprintf("%s/api/v1/projects/%s", c.HostUrl, projectUuid)
	req, err := http.NewRequest("PATCH", path, bytes.NewReader(marshalled))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %v, body: %s", err, string(marshalled))
	}

	// Do request
	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v, body: %s", err, string(marshalled))
	}

	// Unmarshal the response
	response := UpdateProjectV1Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v, body: %s", err, string(body))
	}

	// Validate that the project UUID is present in the response
	if response.Results.ProjectUUID == "" {
		return nil, fmt.Errorf("project UUID is missing in the response")
	}

	return &response.Results, nil
}
