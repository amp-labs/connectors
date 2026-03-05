package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata. It is embedded.
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewReader[staticschema.FieldMetadataMapV2](schemas) // nolint:gochecknoglobals

	Schemas = FileManager.MustLoadSchemas() // nolint:gochecknoglobals
)
