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
	expandableFieldsFile []byte
	expandableFields     = fileregistry.MustParseJSON[ExpandableFieldsDef](expandableFieldsFile) // nolint:gochecknoglobals
)

// ExpandableFieldsDef stores expandable Stripe fields by resource name.
//
// The map keys represent Stripe resource names, and each value contains the set
// of queryable expand paths for that resource.
type ExpandableFieldsDef map[string]datautils.Set[string]
