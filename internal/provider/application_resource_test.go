// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationResource(t *testing.T) {
	testID := acctest.RandomWithPrefix("tfacc-")
	uri := fmt.Sprintf("http://localhost:%d", acctest.RandIntRange(3000, 4000))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "kinde_application" "test" {
					name          = "%[1]s"
					type          = "reg"
					login_uri     = "%[2]s/oauth/login"
					homepage_uri  = "%[2]s"
					logout_uris   = ["%[2]s/oauth/logout"]
					redirect_uris = ["%[2]s/oauth/redirect"]
				}
				`, testID, uri),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_application.test", "name", testID),
					resource.TestCheckResourceAttr("kinde_application.test", "type", "reg"),
					resource.TestCheckResourceAttr("kinde_application.test", "login_uri", uri+"/oauth/login"),
					resource.TestCheckResourceAttr("kinde_application.test", "homepage_uri", uri),
					resource.TestCheckResourceAttr("kinde_application.test", "logout_uris.0", uri+"/oauth/logout"),
					resource.TestCheckResourceAttr("kinde_application.test", "redirect_uris.0", uri+"/oauth/redirect"),
				),
			},
			{
				ResourceName:      "kinde_application.test",
				ImportState:       true,
				ImportStateVerify: true,
				// the api does not return these values
				ImportStateVerifyIgnore: []string{
					"homepage_uri",
					"login_uri",
					"logout_uris",
					"redirect_uris",
				},
			},
		},
	})
}
