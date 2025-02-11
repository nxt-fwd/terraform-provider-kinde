// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIntegrationBasicWorkflow(t *testing.T) {
	testID := acctest.RandomWithPrefix("tfacc-")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create all resources
			{
				Config: testAccIntegrationBasicConfig(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Organization checks
					resource.TestCheckResourceAttr("kinde_organization.test", "name", testID),
					resource.TestCheckResourceAttrSet("kinde_organization.test", "code"),
					resource.TestCheckResourceAttrSet("kinde_organization.test", "id"),

					// Permission checks
					resource.TestCheckResourceAttr("kinde_permission.test", "name", testID+"-permission"),
					resource.TestCheckResourceAttr("kinde_permission.test", "key", testID+"_permission"),
					resource.TestCheckResourceAttrSet("kinde_permission.test", "id"),

					// Role checks
					resource.TestCheckResourceAttr("kinde_role.test", "name", testID+"-role"),
					resource.TestCheckResourceAttr("kinde_role.test", "key", testID+"_role"),
					resource.TestCheckResourceAttrSet("kinde_role.test", "id"),

					// User checks
					resource.TestCheckResourceAttr("kinde_user.test", "first_name", "Test"),
					resource.TestCheckResourceAttr("kinde_user.test", "last_name", "User"),
					resource.TestCheckResourceAttrSet("kinde_user.test", "id"),

					// Organization user checks
					resource.TestCheckResourceAttrSet("kinde_organization_user.test", "id"),
				),
			},
			// Step 2: Import testing for each resource
			{
				ResourceName:            "kinde_organization.test",
				ImportState:            true,
				ImportStateVerify:      true,
				ImportStateVerifyIgnore: []string{
					"created_on",  // Ignore timestamp fields
					"theme_code",  // Ignore theme code as it's set by the API
				},
			},
			{
				ResourceName:      "kinde_permission.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "kinde_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "kinde_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update testing
			{
				Config: testAccIntegrationBasicConfigUpdate(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Organization checks
					resource.TestCheckResourceAttr("kinde_organization.test", "name", testID+"-updated"),
					
					// Permission checks
					resource.TestCheckResourceAttr("kinde_permission.test", "name", testID+"-permission-updated"),
					
					// Role checks
					resource.TestCheckResourceAttr("kinde_role.test", "name", testID+"-role-updated"),
					
					// User checks
					resource.TestCheckResourceAttr("kinde_user.test", "first_name", "Updated"),
					resource.TestCheckResourceAttr("kinde_user.test", "last_name", "User"),
				),
			},
		},
	})
}

func testAccIntegrationBasicConfig(name string) string {
	return fmt.Sprintf(`
resource "kinde_organization" "test" {
	name = %[1]q
}

resource "kinde_permission" "test" {
	name = "%[1]s-permission"
	key  = "%[1]s_permission"
	description = "Test permission"
}

resource "kinde_role" "test" {
	name = "%[1]s-role"
	key  = "%[1]s_role"
	description = "Test role"
	permissions = [kinde_permission.test.id]
}

resource "kinde_user" "test" {
	first_name = "Test"
	last_name  = "User"
	identities = [
		{
			type  = "email"
			value = "test.user.%[1]s@example.com"
		}
	]
}

resource "kinde_organization_user" "test" {
	organization_code = kinde_organization.test.code
	user_id          = kinde_user.test.id
	roles            = []
}
`, name)
}

func testAccIntegrationBasicConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "kinde_organization" "test" {
	name = "%[1]s-updated"
}

resource "kinde_permission" "test" {
	name = "%[1]s-permission-updated"
	key  = "%[1]s_permission"
	description = "Updated test permission"
}

resource "kinde_role" "test" {
	name = "%[1]s-role-updated"
	key  = "%[1]s_role"
	description = "Updated test role"
	permissions = [kinde_permission.test.id]
}

resource "kinde_user" "test" {
	first_name = "Updated"
	last_name  = "User"
	identities = [
		{
			type  = "email"
			value = "test.user.%[1]s@example.com"
		}
	]
}

resource "kinde_organization_user" "test" {
	organization_code = kinde_organization.test.code
	user_id          = kinde_user.test.id
	roles            = []
}
`, name)
}

func TestAccIntegrationRoleManagement(t *testing.T) {
	testID := acctest.RandomWithPrefix("tfacc-")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create role with multiple permissions
			{
				Config: testAccIntegrationRoleConfig(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.complex", "name", testID+"-role"),
					resource.TestCheckResourceAttr("kinde_role.complex", "key", testID+"_role"),
					resource.TestCheckResourceAttrSet("kinde_role.complex", "id"),
					resource.TestCheckResourceAttr("kinde_permission.first", "name", testID+"-permission1"),
					resource.TestCheckResourceAttr("kinde_permission.second", "name", testID+"-permission2"),
				),
			},
			// Step 2: Update role permissions
			{
				Config: testAccIntegrationRoleConfigUpdate(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.complex", "name", testID+"-role-updated"),
					resource.TestCheckResourceAttr("kinde_permission.third", "name", testID+"-permission3"),
				),
			},
		},
	})
}

func testAccIntegrationRoleConfig(name string) string {
	return fmt.Sprintf(`
resource "kinde_permission" "first" {
	name = "%[1]s-permission1"
	key  = "%[1]s_permission1"
	description = "First test permission"
}

resource "kinde_permission" "second" {
	name = "%[1]s-permission2"
	key  = "%[1]s_permission2"
	description = "Second test permission"
}

resource "kinde_role" "complex" {
	name = "%[1]s-role"
	key  = "%[1]s_role"
	description = "Complex test role"
	permissions = [
		kinde_permission.first.id,
		kinde_permission.second.id
	]
}
`, name)
}

func testAccIntegrationRoleConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "kinde_permission" "first" {
	name = "%[1]s-permission1"
	key  = "%[1]s_permission1"
	description = "First test permission"
}

resource "kinde_permission" "second" {
	name = "%[1]s-permission2"
	key  = "%[1]s_permission2"
	description = "Second test permission"
}

resource "kinde_permission" "third" {
	name = "%[1]s-permission3"
	key  = "%[1]s_permission3"
	description = "Third test permission"
}

resource "kinde_role" "complex" {
	name = "%[1]s-role-updated"
	key  = "%[1]s_role"
	description = "Updated complex test role"
	permissions = [
		kinde_permission.first.id,
		kinde_permission.third.id
	]
}
`, name)
}

func TestAccIntegrationUserOrganizations(t *testing.T) {
	testID := acctest.RandomWithPrefix("tfacc-")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create user with multiple organization memberships
			{
				Config: testAccIntegrationUserOrgsConfig(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_user.multi_org", "first_name", "Multi"),
					resource.TestCheckResourceAttr("kinde_user.multi_org", "last_name", "Org"),
					resource.TestCheckResourceAttr("kinde_organization.first", "name", testID+"-org1"),
					resource.TestCheckResourceAttr("kinde_organization.second", "name", testID+"-org2"),
					resource.TestCheckResourceAttrSet("kinde_organization.first", "code"),
					resource.TestCheckResourceAttrSet("kinde_organization.second", "code"),
				),
			},
			// Step 2: Update organization memberships
			{
				Config: testAccIntegrationUserOrgsConfigUpdate(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_organization.first", "name", testID+"-org1-updated"),
					resource.TestCheckResourceAttr("kinde_organization.second", "name", testID+"-org2-updated"),
				),
			},
		},
	})
}

func testAccIntegrationUserOrgsConfig(name string) string {
	return fmt.Sprintf(`
resource "kinde_organization" "first" {
	name = "%[1]s-org1"
}

resource "kinde_organization" "second" {
	name = "%[1]s-org2"
}

resource "kinde_user" "multi_org" {
	first_name = "Multi"
	last_name  = "Org"
	identities = [
		{
			type  = "email"
			value = "multi.org.%[1]s@example.com"
		}
	]
}

resource "kinde_organization_user" "first" {
	organization_code = kinde_organization.first.code
	user_id          = kinde_user.multi_org.id
	roles            = []
}

resource "kinde_organization_user" "second" {
	organization_code = kinde_organization.second.code
	user_id          = kinde_user.multi_org.id
	roles            = []
}
`, name)
}

func testAccIntegrationUserOrgsConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "kinde_organization" "first" {
	name = "%[1]s-org1-updated"
}

resource "kinde_organization" "second" {
	name = "%[1]s-org2-updated"
}

resource "kinde_user" "multi_org" {
	first_name = "Multi"
	last_name  = "Org"
	identities = [
		{
			type  = "email"
			value = "multi.org.%[1]s@example.com"
		}
	]
}

resource "kinde_organization_user" "first" {
	organization_code = kinde_organization.first.code
	user_id          = kinde_user.multi_org.id
	roles            = []
}

resource "kinde_organization_user" "second" {
	organization_code = kinde_organization.second.code
	user_id          = kinde_user.multi_org.id
	roles            = []
}
`, name)
}