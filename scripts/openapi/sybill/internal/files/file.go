package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
)

var (
	// Static file containing openapi spec, from https://api.sybill.ai/docs/openapi.yaml.
	//
	//go:embed v1.json
	apiFile []byte

	OpenApiFile = api.NewFile(apiFile)                 // nolint:gochecknoglobals
	OutputDir   = "providers/sybill/internal/metadata" // nolint:gochecknoglobals
)

// Data returns the raw spec, for the conversation detail schema the list extractor does not reach.
func Data() []byte {
	return apiFile
}
