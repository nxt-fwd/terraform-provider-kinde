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
				data "kinde_application" "test" {
					id = "f61f05b791e142dcb44f113b54b2eee6"
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.kinde_application.test", "id", "f61f05b791e142dcb44f113b54b2eee6"),
					resource.TestCheckResourceAttr("data.kinde_application.test", "name", "Terraform Acceptance Example Application"),
					resource.TestCheckResourceAttr("data.kinde_application.test", "type", "reg"),
				),
			},
		},
	})
}
