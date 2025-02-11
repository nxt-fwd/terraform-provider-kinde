# Basic role with minimal configuration
resource "kinde_role" "basic" {
  name = "basic_role"
  key  = "basic_role"
}

# Role with description
resource "kinde_role" "with_description" {
  name        = "admin_role"
  key         = "admin_role"
  description = "Administrator role with full access"
}

# Role with permissions
resource "kinde_role" "with_permissions" {
  name        = "editor_role"
  key         = "editor_role"
  description = "Editor role with specific permissions"
  permissions = [
    "perm_1234", # Replace with actual permission IDs
    "perm_5678"
  ]
}

# Full example showing all features
resource "kinde_role" "full_example" {
  name        = "full_role"
  key         = "full_role"
  description = "Complete role configuration example"
  permissions = [
    kinde_permission.create_users.id,
    kinde_permission.edit_users.id,
    kinde_permission.delete_users.id
  ]
} 