package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
)

var (
	// Static file containing openapi spec, from https://api.withorb.com/spec.json.
	//
	//go:embed openapi.json
	apiFile []byte

	OpenApiFile = api.NewFile(apiFile)              // nolint:gochecknoglobals
	OutputDir   = "providers/orb/internal/metadata" // nolint:gochecknoglobals
)
