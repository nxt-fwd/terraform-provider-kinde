package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleResource(t *testing.T) {
	testID := acctest.RandomWithPrefix("tfacc-")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRoleResourceConfig(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.test", "name", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "key", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "description", "Test role"),
					resource.TestCheckResourceAttrSet("kinde_role.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "kinde_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccRoleResourceConfigUpdate(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.test", "name", testID+"-updated"),
					resource.TestCheckResourceAttr("kinde_role.test", "key", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "description", "Updated test role"),
				),
			},
		},
	})
}

func TestAccRoleResource_PermissionOrdering(t *testing.T) {
	testID := acctest.RandomWithPrefix("tfacc")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create role with permissions in one order
			{
				Config: testAccRoleResourceConfig_WithPermissions(testID, testID, "Test role with permissions", []string{
					"0194f574-a027-2209-b0d5-0ffc38f47407",
					"0194f574-9fe4-98e1-bba4-e54d0b7989f8",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.test", "name", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "key", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "description", "Test role with permissions"),
					resource.TestCheckResourceAttr("kinde_role.test", "permissions.#", "2"),
					// Check that both permissions exist in the set
					resource.TestCheckTypeSetElemAttr("kinde_role.test", "permissions.*", "0194f574-a027-2209-b0d5-0ffc38f47407"),
					resource.TestCheckTypeSetElemAttr("kinde_role.test", "permissions.*", "0194f574-9fe4-98e1-bba4-e54d0b7989f8"),
				),
			},
			// Update with same permissions in different order - should not trigger a change
			{
				Config: testAccRoleResourceConfig_WithPermissions(testID, testID, "Test role with permissions", []string{
					"0194f574-9fe4-98e1-bba4-e54d0b7989f8",
					"0194f574-a027-2209-b0d5-0ffc38f47407",
				}),
				PlanOnly: true,
			},
		},
	})
}

func TestAccRoleResource_RemovePermissions(t *testing.T) {
	testID := acctest.RandomWithPrefix("tfacc")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create role with two permissions
			{
				Config: testAccRoleResourceConfig_WithPermissions(testID, testID, "Test role with permissions", []string{
					"0194f574-a027-2209-b0d5-0ffc38f47407",
					"0194f574-9fe4-98e1-bba4-e54d0b7989f8",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.test", "name", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "key", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "description", "Test role with permissions"),
					resource.TestCheckResourceAttr("kinde_role.test", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("kinde_role.test", "permissions.*", "0194f574-a027-2209-b0d5-0ffc38f47407"),
					resource.TestCheckTypeSetElemAttr("kinde_role.test", "permissions.*", "0194f574-9fe4-98e1-bba4-e54d0b7989f8"),
				),
			},
			// Remove one permission
			{
				Config: testAccRoleResourceConfig_WithPermissions(testID, testID, "Test role with permissions", []string{
					"0194f574-a027-2209-b0d5-0ffc38f47407",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.test", "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr("kinde_role.test", "permissions.*", "0194f574-a027-2209-b0d5-0ffc38f47407"),
				),
			},
			// Remove all permissions
			{
				Config: testAccRoleResourceConfig_WithPermissions(testID, testID, "Test role with permissions", []string{}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.test", "permissions.#", "0"),
				),
			},
		},
	})
}

func testAccRoleResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "kinde_role" "test" {
	name        = %[1]q
	key         = %[1]q
	description = "Test role"
}
`, name)
}

func testAccRoleResourceConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "kinde_role" "test" {
	name        = "%[1]s-updated"
	key         = %[1]q
	description = "Updated test role"
}
`, name)
}

func testAccRoleResourceConfig_WithPermissions(name, key, description string, permissions []string) string {
	if len(permissions) == 0 {
		return fmt.Sprintf(`
resource "kinde_role" "test" {
	name        = %q
	key         = %q
	description = %q
}
`, name, key, description)
	}

	permissionsStr := "["
	for i, p := range permissions {
		if i > 0 {
			permissionsStr += ", "
		}
		permissionsStr += fmt.Sprintf(`"%s"`, p)
	}
	permissionsStr += "]"

	return fmt.Sprintf(`
resource "kinde_role" "test" {
	name        = %q
	key         = %q
	description = %q
	permissions = %s
}
`, name, key, description, permissionsStr)
}
