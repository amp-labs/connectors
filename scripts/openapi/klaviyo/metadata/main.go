package main

import (
	"github.com/amp-labs/connectors/internal/goutils"
	files "github.com/amp-labs/connectors/providers/klaviyo/openapi"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/filters"
	mapping "github.com/amp-labs/connectors/scripts/openapi/internal/api/map"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/output"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/reducers"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		// Create endpoints, not for reading.
		"/client/subscriptions",
		"/client/push-tokens",
		"/client/events",
		"/client/profiles",
		"/client/event-bulk-create",
		"/client/back-in-stock-subscriptions",
	}
)

func main() {
	extractor, err := api.NewFile(files.File).Extractor()
	goutils.MustBeNil(err)

	getSchemas, err := extractor.ExtractListSchemas(api.GET,
		api.List.WithMediaType("application/vnd.api+json"),
		api.List.WithPropertyFlattener(func(objectName, fieldName string) bool {
			// Nested attributes object holds most important fields.
			return fieldName == "attributes"
		}),
		api.List.WithAutoSelectArrayItems(),
		api.List.WithArrayLocator(api.ArrayLocationAtData),
	)
	goutils.MustBeNil(err)

	pipe := pipeline.NewSchemaPipe(getSchemas).
		Filter(filters.KeepWithPath(filters.AndPathMatcher{
			filters.NoIDPath{},
			filters.NewDenyPathStrategy(ignoreEndpoints),
		})).
		Map(mapping.RemoveURLPrefix("/api")).
		Reduce(reducers.ShortestNameFromURL).
		Map(mapping.DisplayNameFromObjectName).
		Map(mapping.DisplayNameFormat(
			mapping.CamelCaseToSpaceSeparated,
			mapping.CapitalizeFirstLetterEveryWord,
		))

	goutils.MustBeNil(output.WriteMetadata(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteQueryParamStats(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteEndpoints(files.OutputDir, &pipe, nil, nil, nil))
}
