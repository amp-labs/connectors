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

	// Schemas is cached data.
	Schemas = scrapper.NewReader[staticschema.FieldMetadataMapV2](schemas).MustLoadSchemas() // nolint:gochecknoglobals
)
