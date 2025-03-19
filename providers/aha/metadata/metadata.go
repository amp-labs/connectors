package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// There is no OpenAPI available. The schemas.json file is created manually.
// nolint:gochecknoglobals
var (
	//go:embed schemas.json
	schemaContent []byte
	FileManager   = scrapper.NewMetadataFileManager[staticschema.FieldMetadataMapV2](
		schemaContent, fileconv.NewSiblingFileLocator())

	Schemas = FileManager.MustLoadSchemas()
)
