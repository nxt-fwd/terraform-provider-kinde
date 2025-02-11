package provider

import (
	"fmt"
	"strings"
)

// Package provider contains the provider implementation.

// splitID splits a colon-separated ID into its parts and validates the number of parts.
func splitID(id string, expectedParts int, format string) ([]string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != expectedParts {
		return nil, fmt.Errorf("invalid ID format. Expected format: %s", format)
	}
	return parts, nil
}
