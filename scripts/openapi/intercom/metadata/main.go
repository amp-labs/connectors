// OpenAPI documentation can be found under this official github repository.
// https://github.com/intercom/Intercom-OpenAPI/tree/main
// One of the files is chosen and can be found under intercom repository.
// This script will extract schemas and serve object fields via ListObjectMetadata method.
package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/intercom"
	"github.com/amp-labs/connectors/providers/intercom/metadata"
	"github.com/amp-labs/connectors/providers/intercom/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		"/visitors", // doesn't hold a list
		"/me",
		"/tickets/search",
		"/contacts/search",
		"/companies/scroll", // covered by companies
		"/conversations/search",
		"/articles/search", // this one is similar to /articles
	}
	displayNameOverride = map[string]string{ // nolint:gochecknoglobals
		"activity_logs":   "Activity Logs",
		"data_attributes": "Data Attributes",
		"contacts":        "Contacts",
		"teams":           "Teams",
		"conversations":   "Conversations",
		"segments":        "Segments",
		"news_items":      "News Items",
		"newsfeeds":       "Newsfeeds",
		"tickets":         "Tickets",
	}
	searchEndpoints = []string{ // nolint:gochecknoglobals
		"*/search",
	}
	searchObjectEndpoints = map[string]string{ // nolint:gochecknoglobals
		"/contacts/search":      "contacts",
		"/conversations/search": "conversations",
		"/articles/search":      "articles",
		"/tickets/search":       "tickets",
	}
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer()
	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, displayNameOverride,
		api3.CustomMappingObjectCheck(intercom.ObjectNameToResponseField),
	)
	goutils.MustBeNil(err)

	searchObjects, err := explorer.ReadObjectsPost(
		api3.NewAllowPathStrategy(searchEndpoints),
		searchObjectEndpoints, displayNameOverride,
		api3.CustomMappingObjectCheck(intercom.ObjectNameToResponseField),
	)
	goutils.MustBeNil(err)

	objects := searchObjects.Combine(readObjects)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV1]()
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
				staticschema.FieldMetadataMapV1{
					field.Name: field.Name,
				}, nil)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}
