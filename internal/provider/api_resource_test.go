// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIResource(t *testing.T) {
	testID := id.PrefixedUniqueId("tfacc-")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "kinde_api" "test" {
					name     = "%[1]s"
					audience = "%[1]s"
				}
				`, testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_api.test", "name", testID),
					resource.TestCheckResourceAttr("kinde_api.test", "audience", testID),
				),
			},
			{
				ResourceName:      "kinde_api.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
