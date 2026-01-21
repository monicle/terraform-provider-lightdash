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
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccProjectResource_create(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	createConfig010, err := ReadAccTestResource([]string{"resources", "lightdash_project", "create", "010_create.tf"})
	if err != nil {
		t.Fatalf("Failed to get createConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + createConfig010,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_project.test_project", "name", "Test Analytics Project"),
					resource.TestCheckResourceAttr("lightdash_project.test_project", "type", "DEFAULT"),
					resource.TestCheckResourceAttr("lightdash_project.test_project", "dbt_version", "v1.8"),
					resource.TestCheckResourceAttrSet("lightdash_project.test_project", "project_uuid"),
					resource.TestCheckResourceAttr("lightdash_project.test_project", "dbt_connection.type", "github"),
				),
			},
		},
	})
}

func TestAccProjectResource_import(t *testing.T) {
	if !isIntegrationTestMode() {
		t.Skip("Skipping acceptance test for resource_lightdash_project")
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		t.Fatalf("Failed to get providerConfig: %v", err)
	}

	importConfig010, err := ReadAccTestResource([]string{"resources", "lightdash_project", "import", "010_import.tf"})
	if err != nil {
		t.Fatalf("Failed to get importConfig: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + importConfig010,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightdash_project.test_project", "name", "Import Test Project"),
					resource.TestCheckResourceAttr("lightdash_project.test_project", "type", "DEFAULT"),
				),
			},
			{
				Config:            providerConfig + importConfig010,
				ResourceName:      "lightdash_project.test_project",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"dbt_connection.personal_access_token", // Sensitive field not returned by API
					"dbt_connection.repository",            // Connection details not returned by API
					"dbt_connection.branch",
					"dbt_connection.project_sub_path",
					"dbt_connection.host_domain",
					"dbt_connection.target",
					"dbt_connection.type",
					"dbt_connection.authorization_method",
				},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					res, ok := state.RootModule().Resources["lightdash_project.test_project"]
					if !ok {
						return "", fmt.Errorf("resource not found in state for import")
					}
					organizationUUID := res.Primary.Attributes["organization_uuid"]
					projectUUID := res.Primary.Attributes["project_uuid"]
					return fmt.Sprintf("organizations/%s/projects/%s", organizationUUID, projectUUID), nil
				},
			},
		},
	})
}
