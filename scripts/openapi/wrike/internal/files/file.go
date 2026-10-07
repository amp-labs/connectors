package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	// Static file containing the Wrike API v4 OpenAPI spec (OpenAPI 3.0.1, JSON).
	// Source: https://developers.wrike.com/openapi/wrike_api_v4_ver154.yaml
	// (served with a .yaml extension but JSON encoded).
	//
	//go:embed wrike_api_v4_ver154.json
	apiFile []byte

	// LiveFields lists fields observed in live API responses (2026-10-07) that
	// the spec does not declare. The generator merges them on top of the
	// spec-derived fields for objects that already exist; the file never
	// introduces new objects.
	//
	//go:embed live-fields.json
	LiveFields []byte

	FileManager = api3.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals
)
