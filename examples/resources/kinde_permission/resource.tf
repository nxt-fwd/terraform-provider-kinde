# Basic permission with minimal configuration
resource "kinde_permission" "basic" {
  name = "view_users"
  key  = "view_users"
}

# Permission with description
resource "kinde_permission" "with_description" {
  name        = "manage_users"
  key         = "manage_users"
  description = "Permission to manage user accounts"
}

# Create multiple related permissions
resource "kinde_permission" "user_permissions" {
  for_each = {
    view   = "View users"
    create = "Create users"
    update = "Update users"
    delete = "Delete users"
  }

  name        = "user_${each.key}"
  key         = "user_${each.key}"
  description = each.value
}

# Full example showing all features
resource "kinde_permission" "full_example" {
  name        = "admin_access"
  key         = "admin_access"
  description = "Full administrative access"
} 