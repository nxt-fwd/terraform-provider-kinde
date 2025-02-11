package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationResource(t *testing.T) {
	testID := acctest.RandomWithPrefix("tfacc-")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOrganizationResourceConfig(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_organization.test", "name", testID),
					resource.TestCheckResourceAttr("kinde_organization.test", "code", testID),
					resource.TestCheckResourceAttrSet("kinde_organization.test", "id"),
					resource.TestCheckResourceAttrSet("kinde_organization.test", "created_on"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "kinde_organization.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccOrganizationResourceConfigUpdate(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_organization.test", "name", testID+"-updated"),
					resource.TestCheckResourceAttr("kinde_organization.test", "code", testID),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccOrganizationResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "kinde_organization" "test" {
	name = %[1]q
	code = %[1]q
}
`, name)
}

func testAccOrganizationResourceConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "kinde_organization" "test" {
	name = "%[1]s-updated"
	code = %[1]q
}
`, name)
} 