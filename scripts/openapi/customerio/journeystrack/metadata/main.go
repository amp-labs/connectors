package main

import (
	"log"
	"log/slog"

	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/customertrack"
	"github.com/amp-labs/connectors/providers/customertrack/metadata"
	"github.com/amp-labs/connectors/providers/customertrack/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		"/api/v1/accounts/region", // singular object
	}
	objectEndpoints = map[string]string{
		"/api/v1/events":      "events",
		"/api/v1/push/events": "push_events",
	}
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer()
	must(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, nil,
		api3.CustomMappingObjectCheck(customertrack.ObjectNameToResponseField),
	)
	must(err)

	schemas := staticschema.NewMetadata()
	registry := handy.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName, field, object.URLPath, nil)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	must(metadata.FileManager.SaveSchemas(schemas))
	must(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	log.Println("Completed.")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
