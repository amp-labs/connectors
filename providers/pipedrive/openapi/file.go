package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	// The file is downloaded from: https://developers.pipedrive.com/docs/api/v1/openapi.yaml
	//
	//go:embed specs.yaml
	apiFile []byte

	FileManager = api3.NewOpenapiFileManager(apiFile) // nolint:gochecknoglobals
)
