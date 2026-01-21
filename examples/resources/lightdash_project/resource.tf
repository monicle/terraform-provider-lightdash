# Create warehouse credentials first
resource "lightdash_warehouse_credentials" "bigquery" {
  organization_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name              = "BigQuery Production"
  warehouse_type    = "bigquery"
  project           = "my-gcp-project-id"
  dataset           = "my_dataset"
  keyfile_contents  = file("${path.module}/service-account-key.json")
}

# Create a project with GitHub dbt connection
resource "lightdash_project" "analytics" {
  organization_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name              = "Analytics Project"
  type              = "DEFAULT"
  dbt_version       = "v1.8"

  # GitHub dbt connection
  dbt_connection = {
    type                  = "github"
    authorization_method  = "personal_access_token"
    personal_access_token = var.github_token
    repository            = "my-org/dbt-project"
    branch                = "main"
    project_sub_path      = "/"
  }

  # Reference warehouse credentials
  organization_warehouse_credentials_uuid = lightdash_warehouse_credentials.bigquery.organization_warehouse_uuid
}

# Create a preview project
resource "lightdash_project" "preview" {
  organization_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name              = "Preview Project"
  type              = "PREVIEW"
  dbt_version       = "v1.8"

  dbt_connection = {
    type                  = "github"
    authorization_method  = "personal_access_token"
    personal_access_token = var.github_token
    repository            = "my-org/dbt-project"
    branch                = "feature/new-metrics"
    project_sub_path      = "/"
    target                = "dev"
  }

  # Preview project references upstream project
  upstream_project_uuid                   = lightdash_project.analytics.project_uuid
  organization_warehouse_credentials_uuid = lightdash_warehouse_credentials.bigquery.organization_warehouse_uuid
}

# Alternative: Create a project with inline warehouse connection
# This is useful when you want to create a project without creating a separate warehouse credentials resource
resource "lightdash_project" "analytics_inline" {
  organization_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name              = "Analytics Project with Inline Credentials"
  type              = "DEFAULT"
  dbt_version       = "v1.10"

  # GitHub dbt connection
  dbt_connection = {
    type                 = "github"
    authorization_method = "installation_id"
    repository           = "my-org/dbt-project"
    branch               = "main"
    project_sub_path     = "/"
  }

  # Inline warehouse connection (instead of organization_warehouse_credentials_uuid)
  warehouse_connection = {
    type             = "bigquery"
    project          = "my-gcp-project-id"
    dataset          = "my_dataset"
    keyfile_contents = file("${path.module}/service-account-key.json")

    # Optional settings
    authentication_type  = "private_key"
    location             = "US"
    timeout_seconds      = 300
    maximum_bytes_billed = 1000000000
    priority             = "interactive"
    retries              = 3
    start_of_week        = 1
  }
}
