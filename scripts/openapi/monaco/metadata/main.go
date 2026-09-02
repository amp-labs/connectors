// Generates providers/monaco/metadata/schemas.json from the Monaco Public API OpenAPI spec.
//
// Refresh the spec first (https://api.monaco.com/openapi.json is public, no auth needed):
//
//	curl -s -o providers/monaco/metadata/openapi/openapi.json https://api.monaco.com/openapi.json
//
// Then run from repo root:
//
//	go run ./scripts/openapi/monaco/metadata
package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/monaco/metadata"
	"github.com/amp-labs/connectors/providers/monaco/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// postListEndpoints maps URL path → ObjectName for the paginated collections.
// Monaco lists these via POST with a JSON body carrying page/page_size/filters/sort,
// not via GET with query parameters.
//
//nolint:gochecknoglobals
var postListEndpoints = map[string]string{
	"/v1/accounts/list":      "accounts",
	"/v1/audiences/list":     "audiences",
	"/v1/campaigns/list":     "campaigns",
	"/v1/contacts/list":      "contacts",
	"/v1/meetings/list":      "meetings",
	"/v1/opportunities/list": "opportunities",
	"/v1/sequences/list":     "sequences",
	"/v1/tasks/list":         "tasks",
}

// getListEndpoints maps URL path → ObjectName for the collections Monaco serves
// over GET. These return the full set in one shot: the response carries `data`
// but no `pagination` object, so the read layer must not expect a next page.
//
//nolint:gochecknoglobals
var getListEndpoints = map[string]string{
	"/v1/sequence-templates": "sequenceTemplates",
	"/v1/tags/":              "tags",
	"/v1/users/":             "users",
}

// displayNameOverride is spelled out for every object. Without it the explorer
// derives the name from the response envelope's schema title, which yields
// useless values like "Publiclistresponse[Accountresponse]".
//
//nolint:gochecknoglobals
var displayNameOverride = map[string]string{
	"accounts":          "Accounts",
	"audiences":         "Audiences",
	"campaigns":         "Campaigns",
	"contacts":          "Contacts",
	"meetings":          "Meetings",
	"opportunities":     "Opportunities",
	"sequenceTemplates": "Sequence Templates",
	"sequences":         "Sequences",
	"tags":              "Tags",
	"tasks":             "Tasks",
	"users":             "Users",
}

// Note on paths: staticschema.Add runs refactorLongestCommonPath, which hoists
// the longest shared prefix ("/v1") into the module's path and leaves each
// object holding only the remainder ("/contacts/list"). The read layer must
// join module path + object path to rebuild the request URL. Do not re-prefix
// "/v1" here -- it would be hoisted a second time, yielding a module path of
// "/v1/v1".

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range Objects() {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"urlPath", object.URLPath,
				"error", object.Problem,
			)

			continue
		}

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot, object.ObjectName, object.DisplayName,
				object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.", "objects", len(postListEndpoints)+len(getListEndpoints))
}

// Objects reads schemas from both operations and combines them. Every Monaco
// response nests the records under `data`, whether or not it is paginated,
// hence DataObjectLocator for both passes.
func Objects() []metadatadef.Schema {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithArrayItemAutoSelection(),
		api3.WithDisplayNamePostProcessors(
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	postObjects, err := explorer.ReadObjectsPost(
		api3.NewAllowPathStrategy(datautils.Map[string, string](postListEndpoints).Keys()),
		postListEndpoints,
		displayNameOverride,
		api3.DataObjectLocator,
	)
	goutils.MustBeNil(err)

	getObjects, err := explorer.ReadObjectsGet(
		api3.NewAllowPathStrategy(datautils.Map[string, string](getListEndpoints).Keys()),
		getListEndpoints,
		displayNameOverride,
		api3.DataObjectLocator,
	)
	goutils.MustBeNil(err)

	return postObjects.Combine(getObjects)
}
