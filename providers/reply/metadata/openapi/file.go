package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	// Static file containing openapi spec.
	// https://docs.reply.io/api-reference/bundled.yaml
	//
	//go:embed api.reply.yaml
	apiFile []byte

	FileManager = api3.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals
)
