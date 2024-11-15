package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	// Static file containing openapi spec.
	//
	//go:embed rest-v1.json
	apiFileV1 []byte
	//go:embed rest-v2.json
	apiFileV2 []byte

	// Version1FileManager -> https://developer.keap.com/docs/rest/
	Version1FileManager = api3.NewOpenapiFileManager(apiFileV1) // nolint:gochecknoglobals
	// Version2FileManager -> https://developer.keap.com/docs/restv2/
	Version2FileManager = api3.NewOpenapiFileManager(apiFileV2) // nolint:gochecknoglobals
)
