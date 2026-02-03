# Create a personal access token without expiration
resource "lightdash_personal_access_token" "ci_token" {
  description = "CI/CD pipeline token"
}

# Create a personal access token with expiration
resource "lightdash_personal_access_token" "temporary_token" {
  description = "Temporary token for testing"
  expires_at  = "2024-12-31T23:59:59Z"
}

# Output the token value (sensitive)
output "ci_token_value" {
  value     = lightdash_personal_access_token.ci_token.token
  sensitive = true
}
