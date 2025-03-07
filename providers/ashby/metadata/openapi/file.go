package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	//go:embed api.json
	apiFile    []byte
	ApiManager = api3.NewOpenapiFileManager[any](apiFile) //nolint:gochecknoglobals
)
