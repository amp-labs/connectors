package confluence

import (
	_ "embed"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	fileManager = scrapper.NewExtendedMetadataFileManager[staticschema.FieldMetadataMapV2, any](schemas, nil)

	// Schemas is cached Object schemas.
	Schemas = confluenceSchemas{ // nolint:gochecknoglobals
		Metadata: fileManager.MustLoadSchemas(),
	}
)

type confluenceSchemas struct {
	*staticschema.Metadata[staticschema.FieldMetadataMapV2, any]
}

func (s *confluenceSchemas) Select(objectNames []string) (*common.ListObjectMetadataResult, error) {
	// Case-insensitive object names.
	objects := make([]string, len(objectNames))
	for index, name := range objectNames {
		objects[index] = strings.ToLower(name)
	}

	return s.Metadata.Select(providers.ModuleAtlassianConfluence, objects)
}
