package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/getkin/kin-openapi/openapi3"
)

var (
	//go:embed openapi.json
	apiFile []byte

	// FileManager reads the Notion OpenAPI spec.
	// Source: https://developers.notion.com/openapi.json
	FileManager = api3.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals
)

// Document loads the OpenAPI document so resource object schemas can be read directly.
func Document() (*openapi3.T, error) {
	loader := openapi3.NewLoader()

	return loader.LoadFromData(apiFile)
}
