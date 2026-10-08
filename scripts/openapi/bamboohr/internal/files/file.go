package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	// Static file containing openapi spec.
	//
	//go:embed openapi.yaml
	apiFile []byte

	// ManualObjects holds hand-authored list schemas the OpenAPI extractor cannot produce.
	//
	//go:embed manual-objects.json
	ManualObjects []byte

	FileManager = api3.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals
)
