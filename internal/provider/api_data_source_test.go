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
				Config: `
				data "kinde_api" "test" {
					id = "3890a3a1b4a145bd87be0b407ea39345"
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.kinde_api.test", "id", "3890a3a1b4a145bd87be0b407ea39345"),
					resource.TestCheckResourceAttr("data.kinde_api.test", "name", "Terraform Acceptance Example API"),
					resource.TestCheckResourceAttr("data.kinde_api.test", "audience", "https://registry.terraform.io/providers/axatol/kinde"),
				),
			},
		},
	})
}
