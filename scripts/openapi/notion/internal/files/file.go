package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
	"github.com/getkin/kin-openapi/openapi3"
)

var (
	// Static file containing openapi spec.
	// Source: https://developers.notion.com/openapi.json
	//
	//go:embed openapi.json
	apiFile []byte

	OpenApiFile = api.NewFile(apiFile)        // nolint:gochecknoglobals
	OutputDir   = "providers/notion/metadata" // nolint:gochecknoglobals
)

// Document loads the OpenAPI document so resource object schemas can be read from components.
func Document() (*openapi3.T, error) {
	loader := openapi3.NewLoader()

	return loader.LoadFromData(apiFile)
}
