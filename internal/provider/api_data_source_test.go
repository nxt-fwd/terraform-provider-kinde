// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.kinde_api.test", "name", "Terraform Acceptance Test API"),
					resource.TestCheckResourceAttr("data.kinde_api.test", "audience", "https://registry.terraform.io/providers/nxt-fwd/kinde"),
				),
			},
		},
	})
}

func testAccAPIDataSourceConfig() string {
	return `
resource "kinde_api" "test" {
	name     = "Terraform Acceptance Test API"
	audience = "https://registry.terraform.io/providers/nxt-fwd/kinde"
}

data "kinde_api" "test" {
	id = kinde_api.test.id
}
`
}
