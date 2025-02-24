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
	Pagination  string `json:"pagination,omitempty"`
	PageSize    string `json:"pageSize,omitempty"`
	Incremental bool   `json:"incremental,omitempty"`
}

func (s ZendeskSchemas) LookupPaginationType(
	moduleID common.ModuleID, objectName string,
) string {
	if len(moduleID) == 0 {
		moduleID = staticschema.RootModuleID
	}

	ptype := s.Modules[moduleID].Objects[objectName].Custom.Pagination
	if len(ptype) == 0 {
		// If no pagination type is found, the API assumes offset pagination.
		return "offset"
	}

	return ptype
}

func (s ZendeskSchemas) LookupPageSizeQP(
	moduleID common.ModuleID, objectName string,
) string {
	if len(moduleID) == 0 {
		moduleID = staticschema.RootModuleID
	}

	// https://developer.zendesk.com/api-reference/ticketing/ticket-management/incremental_exports/#per_page
	pageSizeQueryParam := s.Modules[moduleID].Objects[objectName].Custom.PageSize
	if len(pageSizeQueryParam) == 0 {
		return "count"
	}

	return pageSizeQueryParam
}


func (s ZendeskSchemas) IsIncrementalRead(
	moduleID common.ModuleID, objectName string,
) bool {
	if len(moduleID) == 0 {
		moduleID = staticschema.RootModuleID
	}

	return s.Modules[moduleID].Objects[objectName].Custom.Incremental
}
