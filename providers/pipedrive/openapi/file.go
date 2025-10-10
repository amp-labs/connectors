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

	//go:embed v2.yaml
	apiFilev2 []byte

	FileManager = api3.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals

	FileManagerV2 = api3.NewOpenapiFileManager[any](apiFilev2) // nolint:gochecknoglobals
)
