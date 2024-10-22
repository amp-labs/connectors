// OpenAPI documentation can be found under this official github repository.
// https://github.com/intercom/Intercom-OpenAPI/tree/main
// One of the files is chosen and can be found under intercom repository.
// This script will extract schemas and serve object fields via ListObjectMetadata method.
package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/common/handy"
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
	}
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer()
	must(err)

	objects, err := explorer.ReadObjectsGet(
		&api3.DenyPathStrategy{
			Paths: ignoreEndpoints,
		}, nil, displayNameOverride,
		api3.CustomMappingObjectCheck(intercom.ObjectNameToResponseField),
	)
	must(err)

	schemas := scrapper.NewObjectMetadataResult()
	registry := handy.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add(object.ObjectName, object.DisplayName, field, nil)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	must(metadata.FileManager.SaveSchemas(schemas))
	must(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
