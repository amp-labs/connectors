package confluence

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	fileManager = scrapper.NewExtendedMetadataFileManager[staticschema.FieldMetadataMapV2, any](schemas, nil)

	// Schemas is cached Object schemas.
	Schemas = fileManager.MustLoadSchemas()
)
