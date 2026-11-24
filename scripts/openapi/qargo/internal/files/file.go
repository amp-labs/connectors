package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
)

var (
	// Static file containing openapi spec.
	// https://api-docs.qargo.com/_bundle/docs.json?download
	//
	//go:embed openapi.json
	apiFile []byte

	// OpenApiData is the raw spec, for component schemas that no list endpoint returns.
	OpenApiData = apiFile                             // nolint:gochecknoglobals
	OpenApiFile = api.NewFile(apiFile)                // nolint:gochecknoglobals
	OutputDir   = "providers/qargo/internal/metadata" // nolint:gochecknoglobals
)
