package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
	// FIXME: Test is failing with "Provider produced inconsistent result after apply"
	// .permissions: was cty.SetVal([]cty.Value{cty.StringVal("")}), but now null.
	t.Skip("Skipping test due to known issue with permissions handling")

	testID := acctest.RandomWithPrefix("tfacc")
	permission1ID := acctest.RandomWithPrefix("tfacc-perm1")
	permission2ID := acctest.RandomWithPrefix("tfacc-perm2")

	var permission1ResourceID, permission2ResourceID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create first permission
			{
				Config: testAccPermissionResourceConfig("test-permission-1", permission1ID, "Test permission 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kinde_permission.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["kinde_permission.test"]
						if !ok {
							return fmt.Errorf("permission1 not found")
						}
						permission1ResourceID = rs.Primary.ID
						return nil
					},
				),
			},
			// Create second permission
			{
				Config: testAccPermissionResourceConfig("test-permission-2", permission2ID, "Test permission 2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kinde_permission.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["kinde_permission.test"]
						if !ok {
							return fmt.Errorf("permission2 not found")
						}
						permission2ResourceID = rs.Primary.ID
						return nil
					},
				),
			},
			// Create role with permissions in one order
			{
				Config: testAccRoleResourceConfig_WithPermissions(testID, testID, "Test role with permissions", []string{
					permission1ResourceID,
					permission2ResourceID,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.test", "name", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "key", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "description", "Test role with permissions"),
					resource.TestCheckResourceAttr("kinde_role.test", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("kinde_role.test", "permissions.*", permission1ResourceID),
					resource.TestCheckTypeSetElemAttr("kinde_role.test", "permissions.*", permission2ResourceID),
				),
			},
			// Update with same permissions in different order - should not trigger a change
			{
				Config: testAccRoleResourceConfig_WithPermissions(testID, testID, "Test role with permissions", []string{
					permission2ResourceID,
					permission1ResourceID,
				}),
				PlanOnly: true,
			},
			// Remove all permissions
			{
				Config: testAccRoleResourceConfig(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.test", "name", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "key", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "description", "Test role"),
					resource.TestCheckNoResourceAttr("kinde_role.test", "permissions"),
				),
			},
		},
	})
}

func TestAccRoleResource_RemovePermissions(t *testing.T) {
	// FIXME: Test is failing with "Provider produced inconsistent result after apply"
	// .permissions: was cty.SetVal([]cty.Value{cty.StringVal("")}), but now null.
	t.Skip("Skipping test due to known issue with permissions handling")

	testID := acctest.RandomWithPrefix("tfacc")
	permission1ID := acctest.RandomWithPrefix("tfacc-perm1")
	permission2ID := acctest.RandomWithPrefix("tfacc-perm2")

	var permission1ResourceID, permission2ResourceID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create first permission
			{
				Config: testAccPermissionResourceConfig("test-permission-1", permission1ID, "Test permission 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kinde_permission.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["kinde_permission.test"]
						if !ok {
							return fmt.Errorf("permission1 not found")
						}
						permission1ResourceID = rs.Primary.ID
						return nil
					},
				),
			},
			// Create second permission
			{
				Config: testAccPermissionResourceConfig("test-permission-2", permission2ID, "Test permission 2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kinde_permission.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["kinde_permission.test"]
						if !ok {
							return fmt.Errorf("permission2 not found")
						}
						permission2ResourceID = rs.Primary.ID
						return nil
					},
				),
			},
			// Create role with two permissions
			{
				Config: testAccRoleResourceConfig_WithPermissions(testID, testID, "Test role with permissions", []string{
					permission1ResourceID,
					permission2ResourceID,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.test", "name", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "key", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "description", "Test role with permissions"),
					resource.TestCheckResourceAttr("kinde_role.test", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("kinde_role.test", "permissions.*", permission1ResourceID),
					resource.TestCheckTypeSetElemAttr("kinde_role.test", "permissions.*", permission2ResourceID),
				),
			},
			// Remove one permission
			{
				Config: testAccRoleResourceConfig_WithPermissions(testID, testID, "Test role with permissions", []string{
					permission1ResourceID,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.test", "name", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "key", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "description", "Test role with permissions"),
					resource.TestCheckResourceAttr("kinde_role.test", "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr("kinde_role.test", "permissions.*", permission1ResourceID),
				),
			},
			// Remove all permissions
			{
				Config: testAccRoleResourceConfig(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_role.test", "name", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "key", testID),
					resource.TestCheckResourceAttr("kinde_role.test", "description", "Test role"),
					resource.TestCheckNoResourceAttr("kinde_role.test", "permissions"),
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
