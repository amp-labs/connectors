package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
)

var (
	// Static file containing openapi spec.
	//
	//go:embed stable.json
	apiFile []byte

	OpenApiFile = api.NewFile(apiFile)                  // nolint:gochecknoglobals
	OutputDir   = "providers/klaviyo/internal/metadata" // nolint:gochecknoglobals
)
