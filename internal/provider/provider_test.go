// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"kinde": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	requiredEnvVars := []string{
		"KINDE_DOMAIN",
		"KINDE_AUDIENCE",
		"KINDE_CLIENT_ID",
		"KINDE_CLIENT_SECRET",
	}

	for _, envVar := range requiredEnvVars {
		if val := os.Getenv(envVar); val == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}
}
