package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api2"
)

var (
	//go:embed esignature-api.json
	apiFile []byte

	FileManager = api2.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals
)
