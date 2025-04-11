package main

import (
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/insightly/metadata"
	"github.com/amp-labs/connectors/providers/insightly/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		// Base endpoint replaced with equivalent which supports incremental reading.
		"/CommunityComments",
		"/CommunityForums",
		"/CommunityPosts",
		"/Contacts",
		"/DocumentTemplates",
		"/Emails",
		"/Events",
		"/ForumCategories",
		"/KnowledgeArticle",
		"/KnowledgeArticleCategory",
		"/KnowledgeArticleFolder",
		"/Leads",
		"/Milestones",
		"/Notes",
		"/Opportunities",
		"/OpportunityLineItem",
		"/Organisations",
		"/Pricebook",
		"/PricebookEntry",
		"/Product",
		"/Projects",
		"/Prospect",
		"/Quotation",
		"/QuotationLineItem",
		"/Tasks",
		"/Ticket",
		"/Users",
		// Singular object.
		"/Instance",
		"/Users/Me",
		// Requires Query params
		"*/SearchByTag",
	}
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
		api3.WithDuplicatesResolver(api3.SingleItemDuplicatesResolver(func(endpoint string) string {
			// Objects that support incremental read end with "Search" uri part.
			// Remove the verb to get objectName as a noun.
			objectName, _ := strings.CutPrefix(endpoint, "/")
			objectName, _ = strings.CutSuffix(objectName, "/Search")

			return objectName
		})),
		api3.WithParameterFilterGetMethod(api3.OnlyOptionalQueryParameters),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, nil,
		api3.DataObjectLocator,
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot,
				object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
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
