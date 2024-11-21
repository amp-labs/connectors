package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/getresponse/metadata"
	"github.com/amp-labs/connectors/providers/getresponse/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		// Endpoints returning single item.
		"/accounts",
		"/accounts/badge",
		"/accounts/billing",
		"/accounts/blocklists",
		"/accounts/callbacks",
		"/file-library/quota",
		"/statistics/ecommerce/performance",
		"/statistics/ecommerce/revenue",
		"/sms", // Sadly, this is a single object.
		"/sms-automation",
		// Non-GET operation endpoints.
		"/search-contacts/contacts",
	}
	objectEndpoints = map[string]string{
		"/autoresponders/statistics":       "autoresponders-statistics",
		"/newsletters/statistics":          "newsletters-statistics",
		"/rss-newsletters/statistics":      "rss-newsletters-statistics",
		"/transactional-emails/statistics": "transactional-emails-statistics",
	}
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, nil,
		api3.DataObjectLocator,
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName,
				field, object.URLPath, object.ResponseKey, nil)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}
