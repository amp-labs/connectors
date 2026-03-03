package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewReader[staticschema.FieldMetadataMapV2](schemas)

	Schemas = FileManager.MustLoadSchemas()
)
