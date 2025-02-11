// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPermissionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPermissionResourceConfig("test-permission", "test_permission", "Test permission description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_permission.test", "name", "test-permission"),
					resource.TestCheckResourceAttr("kinde_permission.test", "key", "test_permission"),
					resource.TestCheckResourceAttr("kinde_permission.test", "description", "Test permission description"),
					resource.TestCheckResourceAttrSet("kinde_permission.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "kinde_permission.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPermissionResourceConfig("updated-permission", "updated_permission", "Updated test permission description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_permission.test", "name", "updated-permission"),
					resource.TestCheckResourceAttr("kinde_permission.test", "key", "updated_permission"),
					resource.TestCheckResourceAttr("kinde_permission.test", "description", "Updated test permission description"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccPermissionResourceConfig(name string, key string, description string) string {
	return fmt.Sprintf(`
resource "kinde_permission" "test" {
  name        = %[1]q
  key         = %[2]q
  description = %[3]q
}
`, name, key, description)
}
