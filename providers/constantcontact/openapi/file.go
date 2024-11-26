package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api2"
)

var (
	// Static file containing openapi spec.
	//
	//go:embed swagger.yaml
	apiFile []byte

	FileManager = api2.NewOpenapiFileManager(apiFile) // nolint:gochecknoglobals
)
