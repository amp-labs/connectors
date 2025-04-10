package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewExtendedMetadataFileManager[staticschema.FieldMetadataMapV2, CustomProperties]( // nolint:gochecknoglobals,lll
		schemas, fileconv.NewSiblingFileLocator())

	// Schemas is cached Object schemas.
	Schemas = ZendeskSchemas{ // nolint:gochecknoglobals
		Metadata: FileManager.MustLoadSchemas(),
	}
)

type ZendeskSchemas struct {
	*staticschema.Metadata[staticschema.FieldMetadataMapV2, CustomProperties]
}

type CustomProperties struct {
	Pagination  string `json:"pagination,omitempty"`
	Incremental bool   `json:"incremental,omitempty"`
}

func (s ZendeskSchemas) LookupPaginationType(
	moduleID common.ModuleID, objectName string,
) string {
	if len(moduleID) == 0 {
		moduleID = common.ModuleRoot
	}

	ptype := s.Modules[moduleID].Objects[objectName].Custom.Pagination
	if len(ptype) == 0 {
		// If no pagination type is found, the API assumes offset pagination.
		return "offset"
	}

	return ptype
}

func (s ZendeskSchemas) IsIncrementalRead(
	moduleID common.ModuleID, objectName string,
) bool {
	if len(moduleID) == 0 {
		moduleID = common.ModuleRoot
	}

	return s.Modules[moduleID].Objects[objectName].Custom.Incremental
}

// nolint:lll
var pageSizes = map[common.ModuleID]datautils.DefaultMap[string, string]{ //nolint:gochecknoglobals
	"ticketing": datautils.NewDefaultMap(map[string]string{
		// Every object below was verified.
		// Increasing any of them would result in Bad Request error.
		// Note some objects couldn't be tested due to account permissions.
		// These are:
		// * "custom_objects"		=> You don't have custom objects enabled on your account. To enable, go to Admin Center &gt; Objects and rules &gt; Custom objects.
		// * "satisfaction_reasons"	=> [Forbidden]You do not have access to this page.
		"custom_objects":       "100",
		"satisfaction_reasons": "100",
		// Objects with strict page size upper limit:
		"activities":                 "100",
		"audit_logs":                 "100",
		"automations":                "100",
		"brands":                     "100",
		"deleted_tickets":            "100",
		"deleted_users":              "100",
		"email_notifications":        "100",
		"groups":                     "100",
		"group_memberships":          "100",
		"job_statuses":               "100",
		"macros":                     "100",
		"organizations":              "1000",
		"organization_fields":        "100",
		"organization_memberships":   "100",
		"organization_subscriptions": "100",
		"recipient_addresses":        "1000",
		"requests":                   "100",
		"satisfaction_ratings":       "100",
		"suspended_tickets":          "1000",
		"tags":                       "1000",
		"ticket_audits":              "100",
		"ticket_fields":              "100",
		"ticket_metrics":             "100",
		"triggers":                   "100",
		"trigger_categories":         "1000",
		"users":                      "1000",
		"user_fields":                "100",
		"views":                      "100",
	}, func(objectName string) string {
		// Any large page size is accepted by API without an error.
		return "2000"
	}),
	"help-center": datautils.NewDefaultMap(nil,
		func(objectName string) string {
			// Every object uses 100 as a page size.
			return "100"
		},
	),
}

func (s ZendeskSchemas) PageSize(
	moduleID common.ModuleID, objectName string,
) string {
	if len(moduleID) == 0 {
		moduleID = common.ModuleRoot
	}

	// TODO the map should be part of schema.json.
	return pageSizes[moduleID].Get(objectName)
}
