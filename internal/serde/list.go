// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package serde

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ExpandStringList(ctx context.Context, input types.List) ([]string, diag.Diagnostics) {
	var output []string
	diags := input.ElementsAs(ctx, &output, true)
	return output, diags
}

func FlattenStringList(ctx context.Context, input []string) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	output, diags := types.ListValueFrom(ctx, types.StringType, input)
	return output, diags
}
