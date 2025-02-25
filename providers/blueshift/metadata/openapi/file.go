package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

// Static file containing openapi spec.

var (
	//go:embed api.yaml
	apiFile []byte

	FileManager = api3.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals
)
