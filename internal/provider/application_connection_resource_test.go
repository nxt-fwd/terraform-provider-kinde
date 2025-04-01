// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationConnectionResource(t *testing.T) {
	testID := acctest.RandomWithPrefix("tfacc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApplicationConnectionResourceConfig(testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("kinde_application_connection.test", "application_id", "kinde_application.test", "id"),
					resource.TestCheckResourceAttrPair("kinde_application_connection.test", "connection_id", "kinde_connection.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "kinde_application_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccApplicationConnectionResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "kinde_application" "test" {
	name = %[1]q
	type = "reg"
}

resource "kinde_connection" "test" {
	name         = %[1]q
	display_name = "Test OAuth2 Connection"
	strategy     = "oauth2:google"
	options = {
		client_id     = "test-client-id"
		client_secret = "test-client-secret"
	}
}

resource "kinde_application_connection" "test" {
	application_id = kinde_application.test.id
	connection_id  = kinde_connection.test.id
}
`, name)
}
