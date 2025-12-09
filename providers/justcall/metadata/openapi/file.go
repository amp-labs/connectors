package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	// Static file containing openapi spec.
	//
	//go:embed api-v21.json
	apiFile []byte

	FileManager = api3.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals
)

// GetAPIFile returns the raw OpenAPI file bytes for custom parsing.
func GetAPIFile() []byte {
	return apiFile
}
