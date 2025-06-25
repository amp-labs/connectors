package pardot

import (
	_ "embed"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewExtendedMetadataFileManager[staticschema.FieldMetadataMapV2, any]( // nolint:gochecknoglobals,lll
		schemas, fileconv.NewSiblingFileLocator())

	// Schemas is cached Object schemas.
	Schemas = pardotSchemas{ // nolint:gochecknoglobals
		Metadata: FileManager.MustLoadSchemas(),
	}
)

type pardotSchemas struct {
	*staticschema.Metadata[staticschema.FieldMetadataMapV2, any]
}

func (s *pardotSchemas) Select(objectNames []string) (*common.ListObjectMetadataResult, error) {
	// Case-insensitive object names.
	objects := make([]string, len(objectNames))
	for index, name := range objectNames {
		objects[index] = strings.ToLower(name)
	}

	return s.Metadata.Select(providers.ModuleSalesforceAccountEngagement, objects)
}
