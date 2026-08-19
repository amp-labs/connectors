package main

import (
	"log"
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/pipedrive/metadata"
	"github.com/amp-labs/connectors/providers/pipedrive/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var ignoreEndpoints = []string{ // nolint:gochecknoglobals
	"/projects/boards",
	"/dealFields",
	"/users/*",
	"/itemSearch",
	"/itemSearch/field",
	"/field",
	"/activities/*",
	"/deals/*",
	"/billing/subscriptions/addons",
	// Searching endpoints.
	"*/search",
	"/goals/find",
	// Not an array.
	"/filters/helpers",
	"/userConnections",
	"/userSettings",
	// API - Beta version
	"/organizations/collection",
	"/persons/collection",
}

var displayName = map[string]string{ // nolint:gochecknoglobals
	"activities":                "Activities",
	"activityTypes":             "Activity Types",
	"activityFields":            "Activity Fields",
	"callLogs":                  "Call Logs",
	"currencies":                "Currencies",
	"deals":                     "Deals",
	"dealFields":                "Deal Fields",
	"files":                     "Files",
	"filters":                   "Filters",
	"leadLabels":                "Lead Labels",
	"leadSources":               "Lead Sources",
	"leads":                     "Leads",
	"legacyTeams":               "Legacy Teams",
	"mailboxes/mailThreads":     "Mail Threads",
	"noteFields":                "Note Fields",
	"notes":                     "Notes",
	"organizationFields":        "organization Fields",
	"organizationRelationships": "Organization Relationships",
	"organizations":             "Organizations",
	"permissionSets":            "Permission Sets",
	"personFields":              "Person Fields",
	"persons":                   "Persons",
	"phases":                    "Phases",
	"pipelines":                 "Pipelines",
	"productFields":             "product Fields",
	"products":                  "Products",
	"projectTemplates":          "Project Templates",
	"projects":                  "Projects",
	"recents":                   "Recents",
	"roles":                     "Roles",
	"stages":                    "Stages",
	"tasks":                     "Tasks",
	"users":                     "Users",
	"webhooks":                  "Webhooks",
}

// objectEndpoints maps a URL path to the object name it is exposed as, for paths where the
// last path segment isn't the name we want. Without this the generator would infer
// "mailThreads" from /mailbox/mailThreads.
var objectEndpoints = map[string]string{ // nolint:gochecknoglobals
	"/mailbox/mailThreads": "mailboxes/mailThreads",
}

func main() {
	explorer, err := openapi.FileManager.GetExplorer()
	if err != nil {
		log.Fatalln(err)
	}

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, displayName,
		api3.DataObjectLocator)
	if err != nil {
		log.Fatalln(err)
	}

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	if err := metadata.FileManager.SaveSchemas(schemas); err != nil {
		log.Fatalln(err)
	}

	if err := metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)); err != nil {
		log.Fatalln(err)
	}

	slog.Info("Completed.")
}
