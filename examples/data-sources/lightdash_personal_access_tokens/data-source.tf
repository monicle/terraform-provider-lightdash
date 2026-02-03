# List all personal access tokens for the authenticated user
data "lightdash_personal_access_tokens" "all" {}

# Output the list of tokens
output "tokens" {
  value = data.lightdash_personal_access_tokens.all.tokens
}
