// OpenAPI documentation can be found under this official github repository.
// https://github.com/intercom/Intercom-OpenAPI/tree/main
// One of the files is chosen and can be found under intercom repository.
// This script will extract schemas and serve object fields via ListObjectMetadata method.
package main

import (
	"log/slog"

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
	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		// none
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
	objectNameToResponseField = map[string]string{ // nolint:gochecknoglobals
		"admins":        "admins",
		"teams":         "teams",
		"ticket_types":  "ticket_types",
		"events":        "events",
		"segments":      "segments",
		"activity_logs": "activity_logs",
		// the rest uses `data`
	}
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer()
	must(err)

	objects, err := explorer.GetBasicReadObjects(
		ignoreEndpoints, objectEndpoints, displayNameOverride, IsResponseFieldAppropriate,
	)
	must(err)

	schemas := scrapper.NewObjectMetadataResult()

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
	}

	must(metadata.FileManager.SaveSchemas(schemas))

	slog.Info("Completed.")
}

func IsResponseFieldAppropriate(fieldName, objectName string) bool {
	if responseFieldName, ok := objectNameToResponseField[objectName]; ok {
		return fieldName == responseFieldName
	}

	// Other objects have items located under `data` response field.
	return api3.DataObjectCheck(fieldName, objectName)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
