package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	// Static file containing openapi spec.
	//
	//go:embed openapi.json
	apiFile []byte

	FileManager = api3.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals
)

// FileBytes exposes the raw spec for generators that need direct component
// schema access beyond what api3's collection-oriented readers extract.
func FileBytes() []byte {
	return apiFile
}
