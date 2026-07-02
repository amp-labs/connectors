package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/fileregistry"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewMetadataFileManager[staticschema.FieldMetadataMapV2]( // nolint:gochecknoglobals
		schemas, fileconv.NewSiblingFileLocator())

	// Schemas is cached Object schemas.
	Schemas = FileManager.MustLoadSchemas() // nolint:gochecknoglobals

	//go:embed expandableFields.json
	expandableFields []byte

	ExpandableFields = fileregistry.MustParseJSON[ExpandableFieldsDef](expandableFields) // nolint:gochecknoglobals
)

// ExpandableFieldsDef is a registry of object names to expandable query params for a field.
type ExpandableFieldsDef map[string]datautils.Set[string]
