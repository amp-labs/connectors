package openapi

import (
	_ "embed"
)

var (
	// Static file containing openapi spec.
	//
	//go:embed stable.json
	apiFile []byte

	File      = apiFile                      // nolint:gochecknoglobals
	OutputDir = "providers/klaviyo/metadata" // nolint:gochecknoglobals
)
