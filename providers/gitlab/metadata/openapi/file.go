package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api2"
)

var (
	//go:embed openapi_v2.yaml
	apiFile    []byte
	ApiManager = api2.NewOpenapiFileManager[any](apiFile) //nolint:gochecknoglobals
)
