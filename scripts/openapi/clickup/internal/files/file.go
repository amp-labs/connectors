package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
)

var (
	// Static files containing openapi spec.
	//go:embed clickup-api-v2.json
	apiFileV2 []byte
	//go:embed clickup-api-v3.json
	apiFileV3 []byte

	// ManualObjects holds hand-authored list schemas the OpenAPI extractor cannot produce.
	//
	//go:embed manual-objects.json
	ManualObjects []byte

	OpenApiFileV2 = api.NewFile(apiFileV2)                // nolint:gochecknoglobals
	OpenApiFileV3 = api.NewFile(apiFileV3)                // nolint:gochecknoglobals
	OutputDir     = "providers/clickup/internal/metadata" // nolint:gochecknoglobals
)
