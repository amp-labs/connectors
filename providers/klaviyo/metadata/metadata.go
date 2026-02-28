package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewReader[staticschema.FieldMetadataMapV2](schemas) // nolint:gochecknoglobals

	// Schemas is cached Object schemas.
	Schemas = FileManager.MustLoadSchemas() // nolint:gochecknoglobals
)
