package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewMetadataFileManager[staticschema.FieldMetadataMapV1]( // nolint:gochecknoglobals
		schemas, fileconv.NewSiblingFileLocator())

	// Schemas is cached Object schemas.
	Schemas = FileManager.MustLoadSchemas() // nolint:gochecknoglobals
)
