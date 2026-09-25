package main

import (
	"strings"

	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers/intercom"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/filters"
	mapping "github.com/amp-labs/connectors/scripts/openapi/internal/api/map"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/merging"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/output"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/reducers"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/amp-labs/connectors/scripts/openapi/sellsy/internal/files"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		// Singular object.
		"*/metas",
		"/accounts/conformities",
		"/email/authenticate",
		"/quotas",
		"/scopes",
		"/scopes/tree",
		"/search",
		"/settings/accounting-charts",
		"/settings/email",
	}
	searchEndpoints = []string{
		"*/search",
	}
	ignoreSearchEndpoints = []string{
		// Singular object.
		"/estimates/search",
	}
)

func main() {
	params := []api.ListOption{
		api.List.WithAutoSelectArrayItems(),
		api.List.WithArrayLocator(
			api.ArrayLocationFromMap(intercom.ObjectNameToResponseField),
		),
	}

	explorer, err := files.OpenAPIFile.Extractor()
	goutils.MustBeNil(err)

	getSchemas, err := explorer.ExtractListSchemas(api.GET, params...)
	goutils.MustBeNil(err)
	postSchemas, err := explorer.ExtractListSchemas(api.POST, params...)
	goutils.MustBeNil(err)

	// Select objects with GET operation.
	listPipeline := pipeline.NewSchemaPipe(getSchemas).
		Filter(filters.KeepWithPath(filters.AndPathMatcher{
			filters.NoIDPath{},
			filters.NewDenyPathStrategy(ignoreEndpoints),
		})).
		Reduce(reducers.ShortestNameFromURL)

	// Select objects with POST operation that ends with "/search"
	searchPipeline := pipeline.NewSchemaPipe(postSchemas).
		Filter(filters.KeepWithPath(filters.AndPathMatcher{
			filters.NoIDPath{},
			filters.NewAllowPathStrategy(searchEndpoints),
			filters.NewDenyPathStrategy(ignoreSearchEndpoints),
		})).
		Reduce(reducers.ShortestNameFromURL).
		Map(normalizeSearchObjectName)

	// When object has both search (POST) and normal (GET) read endpoints,
	// we choose search which allows incremental read.
	allObjectsPipe := pipeline.Combine(
		listPipeline, searchPipeline,
		merging.CombineByObjectName,
		merging.ChooseRight,
	)

	// Format display name.
	// (Despite the source GET/POST endpoint, the formating is the same).
	allObjectsPipe = allObjectsPipe.
		Map(mapping.DisplayNameFromObjectName).
		Map(mapping.DisplayNameFormat(
			mapping.SlashesToSpaceSeparated,
			mapping.CamelCaseToSpaceSeparated,
			mapping.CapitalizeFirstLetterEveryWord,
		))

	goutils.MustBeNil(output.WriteMetadata(files.OutputSellsyDir, allObjectsPipe))
	goutils.MustBeNil(output.WriteQueryParamStats(files.OutputSellsyDir, allObjectsPipe))
	goutils.MustBeNil(output.WriteEndpoints(files.OutputSellsyDir, &allObjectsPipe, nil, nil, nil))
}

func normalizeSearchObjectName(schema spec.Schema) spec.Schema {
	// Remove search suffix if any.
	schema.ObjectName, _ = strings.CutSuffix(schema.ObjectName, "/search")

	return schema
}
