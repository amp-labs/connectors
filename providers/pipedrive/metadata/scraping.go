package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	//go:embed schemasV2.json
	schemasFileV2 []byte

	FileManager = scrapper.NewMetadataFileManager[staticschema.FieldMetadataMapV2]( // nolint:gochecknoglobals
		schemas, fileconv.NewSiblingFileLocator())

	FileManagerV2 = scrapper.NewMetadataFileManager[staticschema.FieldMetadataMapV2]( // nolint:gochecknoglobals
		schemasFileV2, fileconv.NewSiblingFileLocator())

	// Schemas is cached Object schemas.
	Schemas = FileManager.MustLoadSchemas() // nolint:gochecknoglobals

	// Schemas v2 is cached object schemas for pipedrive v2.
	SchemasV2 = FileManagerV2.MustLoadSchemas() // nolint:gochecknoglobals
)
