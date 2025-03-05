package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/github/metadata"
	"github.com/amp-labs/connectors/providers/github/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ //nolint:gochecknoglobals
		"/search/issues",
		"/search/repositories",
		"/search/users",
		"/versions",
		"/user/interaction-limits",
		"/user",
		"/gitignore/templates",
		"/feeds",
		"/meta",
		"/rate_limit",
		"/emojis",
		"/search/commits",
		"/app",
		"/public-key",
		"/app/hook/config",
		"/user/codespaces/secrets/public-key",
		"/search/labels",
		"/search/topics",
		"/search/code",
		"/app/installations",
		"/notifications",
	}

	objectEndpoints = map[string]string{ //nolint:gochecknoglobals
		"/user/issues":                       "user/issues",
		"/search/issues":                     "search/issues",
		"/installation/repositories":         "installation/repositories",
		"/user/memberships/orgs":             "user/memberships/orgs",
		"/marketplace_listing/stubbed/plans": "marketplace_listing/stubbed/plans",
		"/marketplace_listing/plans":         "marketplace_listing/plans",
		"/user/installations":                "user/installations",
		"/gists/starred":                     "gists/starred",
		"/user/starred":                      "user/starred",
		"/gists/public":                      "gists/public",
	}

	overrideDisplayName = map[string]string{ //nolint:gochecknoglobals
		"search/issues":                     "Search Issues",
		"user/issues":                       "User Issues",
		"user/starred":                      "User Starred",
		"gists/starred":                     "Gists Starred",
		"gists/public":                      "Gists Public",
		"user/installations":                "User Installations",
		"installation/repositories":         "Installation Repositories",
		"marketplace_listing/plans":         "Marketplace Listing Plans",
		"marketplace_listing/stubbed/plans": "Marketplace Listing Stubbed Plans",
		"user/memberships/orgs":             "User Memberships Orgs",
	}

	objectNametoResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
		"user/installations":        "installations",
		"installation/repositories": "repositories",
	},
		func(objectName string) (fieldName string) {
			return objectName
		},
	)
)

func main() {
	explorer, err := openapi.ApiManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(api3.CamelCaseToSpaceSeparated, api3.CapitalizeFirstLetterEveryWord),
	)

	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, overrideDisplayName, api3.CustomMappingObjectCheck(objectNametoResponseField),
	)

	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range readObjects { //nolint:gochecknoglobals
		if object.Problem != nil {
			slog.Error("Schema not extracted",
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

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}
