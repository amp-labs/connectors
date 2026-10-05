package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/scrapper"
)

//nolint:gochecknoglobals
var (
	//go:embed schemas.json
	schemaContent []byte

	Schemas = scrapper.NewReader[staticschema.FieldMetadataMapV2](schemaContent).MustLoadSchemas()
)
