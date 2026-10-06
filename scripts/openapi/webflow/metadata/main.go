// Generates providers/webflow/metadata/schemas.json from the Webflow Data API v2
// OpenAPI spec (scripts/openapi/webflow/internal/files/v2.yml).
//
// Run from the repo root:
//
//	go run ./scripts/openapi/webflow/metadata
//
// Objects are curated explicitly (allow-list). Almost every Webflow list
// endpoint is scoped to a site (/sites/{site_id}/...); the {site_id}
// placeholder is kept verbatim in schemas.json and the read connector
// substitutes the site at request time (same approach as Mailgun's
// {domain_name}). ReadObjects is used instead of ReadObjectsGet because the
// latter drops every path containing a path parameter.
//
// Excluded on purpose:
//   - CMS items (/collections/{collection_id}/items): fields are defined per
//     collection, so they need dynamic metadata rather than a static schema.
//   - Two-level nested lists (comment replies, per-form submissions).
//   - Analyze reports, DOM/content endpoints, singletons and token endpoints.
package main

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/webflow/metadata"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/scripts/openapi/webflow/internal/files"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// The spec's server URL is https://api.webflow.com/v2 while the catalog base
// URL is https://api.webflow.com, so the version is prepended to every path.
const apiVersionPrefix = "/v2"

type liveFields struct {
	ObjectName string                          `json:"objectName"`
	Fields     staticschema.FieldMetadataMapV2 `json:"fields"`
}

// supportedObjects maps spec URL paths to object names.
// Object names default to the last URL segment (what Webflow's docs call the
// resource); the two exceptions are segments that are meaningless on their own.
//
//nolint:gochecknoglobals
var supportedObjects = map[string]string{
	// Account-scoped.
	"/sites": "sites",

	// Site-scoped.
	"/sites/{site_id}/activity_logs":            "activity_logs",
	"/sites/{site_id}/asset_folders":            "asset_folders",
	"/sites/{site_id}/assets":                   "assets",
	"/sites/{site_id}/collections":              "collections",
	"/sites/{site_id}/comments":                 "comments",
	"/sites/{site_id}/components":               "components",
	"/sites/{site_id}/custom_code/blocks":       "custom_code_blocks",
	"/sites/{site_id}/custom_domains":           "custom_domains",
	"/sites/{site_id}/custom_fonts":             "custom_fonts",
	"/sites/{site_id}/form_submissions":         "form_submissions",
	"/sites/{site_id}/forms":                    "forms",
	"/sites/{site_id}/integrations/google_tags": "google_tags",
	"/sites/{site_id}/orders":                   "orders",
	"/sites/{site_id}/pages":                    "pages",
	"/sites/{site_id}/products":                 "products",
	"/sites/{site_id}/redirects":                "redirects",
	"/sites/{site_id}/registered_scripts":       "registered_scripts",
	"/sites/{site_id}/webhooks":                 "webhooks",
}

//nolint:gochecknoglobals
var displayNameOverrides = map[string]string{
	"activity_logs": "Site Activity Logs",
	"comments":      "Comment Threads",
}

func pathsOf(m map[string]string) []string {
	paths := make([]string, 0, len(m))
	for path := range m {
		paths = append(paths, path)
	}

	return paths
}

func main() {
	explorer, err := files.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			func(displayName string) string {
				return strings.ReplaceAll(displayName, "_", " ")
			},
			api3.SlashesToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		// Every list response has exactly one array next to an optional
		// pagination object, so the records array is picked automatically.
		api3.WithArrayItemAutoSelection(),
		// GET /sites/{site_id}/products returns items of {product, skus}.
		// The record is kept as-is (no flattening) so metadata matches the
		// raw records that Read returns.
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjects("GET",
		api3.AndPathMatcher{
			api3.NewAllowPathStrategy(pathsOf(supportedObjects)),
			// Allows the single {site_id} scope, rejects deeper nesting.
			api3.NestedIDPathIgnorer{},
		},
		supportedObjects,
		displayNameOverrides,
		nil,
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"urlPath", object.URLPath,
				"error", object.Problem,
			)

			continue
		}

		urlPath := apiVersionPrefix + object.URLPath

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot, object.ObjectName, object.DisplayName,
				urlPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	addLiveFields(schemas)

	// FlushSchemas keeps full paths; SaveSchemas would factor the longest
	// common prefix ("/v2/sites") out into the module path, which is both
	// deprecated and misleading here.
	goutils.MustBeNil(metadata.FileManager.FlushSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.", "objects", len(objects), "expected", len(supportedObjects))
}

// addLiveFields merges undocumented-but-observed fields (files.LiveFields) into existing objects.
func addLiveFields(schemas *staticschema.Metadata[staticschema.FieldMetadataMapV2, any]) {
	var overlays []liveFields

	goutils.MustBeNil(json.Unmarshal(files.LiveFields, &overlays))

	objects := schemas.Modules[common.ModuleRoot].Objects

	for _, overlay := range overlays {
		object, ok := objects[overlay.ObjectName]
		if !ok {
			slog.Error("live field overlay targets unknown object", "objectName", overlay.ObjectName)

			continue
		}

		schemas.Add(common.ModuleRoot, overlay.ObjectName, object.DisplayName,
			object.URLPath, object.ResponseKey, overlay.Fields, nil, nil)
	}
}
