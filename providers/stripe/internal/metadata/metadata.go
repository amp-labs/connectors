package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/fileregistry"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	fileManager = scrapper.NewReader[staticschema.FieldMetadataMapV2](schemas) // nolint:gochecknoglobals

	// Schemas is cached Object schemas.
	Schemas = fileManager.MustLoadSchemas() // nolint:gochecknoglobals

	//go:embed expandableFields.json
	expandableFieldsFile []byte
	expandableFields     = fileregistry.MustParseJSON[ExpandableFieldsDef](expandableFieldsFile) // nolint:gochecknoglobals
)

// ExpandableFieldsDef stores expandable Stripe fields by resource name.
//
// The map keys represent Stripe resource names, and each value contains the set
// of queryable expand paths for that resource.
type ExpandableFieldsDef map[string]datautils.Set[string]
