terraform {
  required_providers {
    kinde = {
      source = "nxt-fwd/kinde"
    }
  }
}

provider "kinde" {
  # Configure using environment variables:
  # KINDE_DOMAIN
  # KINDE_AUDIENCE
  # KINDE_CLIENT_ID
  # KINDE_CLIENT_SECRET
}

# Create the read-only role
resource "kinde_role" "readonly" {
  name        = "Read Only"
  key         = "readonly"
  description = "Role with read-only access to resources"
  permissions = [
    "view_users",
    "view_organizations",
    "view_applications"
  ]
}

# Create the admin role
resource "kinde_role" "admin" {
  name        = "Administrator"
  key         = "admin"
  description = "Role with full administrative access"
  permissions = [
    "view_users",
    "create_users",
    "update_users",
    "delete_users",
    "view_organizations",
    "create_organizations",
    "update_organizations",
    "delete_organizations",
    "view_applications",
    "create_applications",
    "update_applications",
    "delete_applications"
  ]
}

# Create a read-only user
resource "kinde_user" "readonly_user" {
  first_name = "John"
  last_name  = "Reader"
  identities = [
    {
      type  = "email"
      value = "john.reader@example.com"
    }
  ]
}

# Create an admin user
resource "kinde_user" "admin_user" {
  first_name = "Jane"
  last_name  = "Admin"
  identities = [
    {
      type  = "email"
      value = "jane.admin@example.com"
    }
  ]
}

# Assign the read-only role to the read-only user
resource "kinde_user_role" "readonly_assignment" {
  user_id = kinde_user.readonly_user.id
  role_id = kinde_role.readonly.id
}

# Assign the admin role to the admin user
resource "kinde_user_role" "admin_assignment" {
  user_id = kinde_user.admin_user.id
  role_id = kinde_role.admin.id
}

# Output the role IDs for reference
output "readonly_role_id" {
  value = kinde_role.readonly.id
}

output "admin_role_id" {
  value = kinde_role.admin.id
}

# Output the user IDs for reference
output "readonly_user_id" {
  value = kinde_user.readonly_user.id
}

output "admin_user_id" {
  value = kinde_user.admin_user.id
} 