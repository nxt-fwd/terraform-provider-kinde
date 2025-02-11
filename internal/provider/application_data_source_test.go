// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "kinde_application" "test" {
					name = "Terraform Acceptance Example Application"
					type = "reg"
				}

				data "kinde_application" "test" {
					id = kinde_application.test.id
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.kinde_application.test", "id"),
					resource.TestCheckResourceAttr("data.kinde_application.test", "name", "Terraform Acceptance Example Application"),
					resource.TestCheckResourceAttr("data.kinde_application.test", "type", "reg"),
				),
			},
		},
	})
}
