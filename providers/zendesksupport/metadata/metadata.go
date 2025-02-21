package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewExtendedMetadataFileManager[staticschema.FieldMetadataMapV1, CustomProperties]( // nolint:gochecknoglobals,lll
		schemas, fileconv.NewSiblingFileLocator())

	// Schemas is cached Object schemas.
	Schemas = ZendeskSchemas{ // nolint:gochecknoglobals
		Metadata: FileManager.MustLoadSchemas(),
	}
)

type ZendeskSchemas struct {
	*staticschema.Metadata[staticschema.FieldMetadataMapV1, CustomProperties]
}

type CustomProperties struct {
	Pagination string `json:"pagination,omitempty"`
}

func (s ZendeskSchemas) LookupPaginationType(
	moduleID common.ModuleID, objectName string,
) (string, bool) {
	if len(moduleID) == 0 {
		moduleID = staticschema.RootModuleID
	}

	ptype := s.Modules[moduleID].Objects[objectName].Custom.Pagination
	if len(ptype) == 0 {
		return "", false
	}

	return ptype, true
}
