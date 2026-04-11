package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// There is no public OpenAPI spec available for Gusto.
	// The schemas.json file is created manually from the API documentation.
	// Reference: https://docs.gusto.com/app-integrations/reference
	//
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewMetadataFileManager[staticschema.FieldMetadataMapV2]( // nolint:gochecknoglobals
		schemas, fileconv.NewSiblingFileLocator())

	Schemas = FileManager.MustLoadSchemas() // nolint:gochecknoglobals
)
