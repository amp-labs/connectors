package main

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/scripts/openapi/lob/internal/files"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var ignoreEndpoints = []string{ // nolint:gochecknoglobals
	// Accounts is not a plural to signify list of accounts.
	// It is a possessive form.
	// See quote from OpenAPI: "Account's current balance of Lob Credits".
	"/accounts",
	// This is analytics, not the typical resource collection.
	// https://docs.lob.com/#tag/QR-Codes/operation/qr_codes_list
	"/qr_code_analytics",
}

func main() { // nolint:funlen
	explorer, err := files.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjects(http.MethodGet,
		api3.AndPathMatcher{
			api3.IDPathIgnorer{},
			api3.NewDenyPathStrategy(ignoreEndpoints),
		},
		nil, nil,
		arrayLocator,
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		objectName, _ := strings.CutPrefix(object.URLPath, "/")

		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", objectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			fieldMetadata := staticschema.FieldMetadataMapV2{
				field.Name: staticschema.FieldMetadata{
					DisplayName:  formatFieldDisplayName(field.Name),
					ValueType:    utilsopenapi.GetFieldValueType(field),
					ProviderType: field.Type,
					Values:       utilsopenapi.GetFieldValueOptions(field),
				},
			}
			schemas.Add("", objectName, object.DisplayName, object.URLPath, object.ResponseKey,
				fieldMetadata, nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, objectName)
		}
	}

	goutils.MustBeNil(files.OutputLob.FlushSchemas(schemas))
	goutils.MustBeNil(files.OutputLob.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func formatFieldDisplayName(fieldName string) string {
	displayName := fieldName
	displayName = naming.SeparateUnderscoreWords(displayName)
	displayName = naming.CapitalizeFirstLetterEveryWord(displayName)

	return displayName
}

func arrayLocator(objectName, fieldName string) bool {
	slog.Warn("unexpected call to locator, provider API was expected to have no ambiguous array fields",
		"object", objectName)

	return false
}
