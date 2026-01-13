package main

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/scripts/openapi/google/internal/files"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		// Endpoints to ignore because they return single configuration objects,
		// not resource collections. Our schema extraction targets collection-based
		// resources (list/create/update flows), so singleton "settings/profile"
		// endpoints are applicable and therefore excluded.
		"/gmail/v1/users/{userId}/settings/autoForwarding",
		"/gmail/v1/users/{userId}/settings/language",
		"/gmail/v1/users/{userId}/profile",
		"/gmail/v1/users/{userId}/settings/vacation",
		"/gmail/v1/users/{userId}/settings/pop",
		"/gmail/v1/users/{userId}/settings/imap",
	}
)

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range Objects() {
		urlPath, _ := strings.CutPrefix(object.URLPath, "/gmail/v1")
		// All Gmail endpoints require user identifier.
		// Luckily we can reference current user using an alias "me".
		urlPath = strings.ReplaceAll(urlPath, "{userId}", "me")

		objectName := object.ObjectName

		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", objectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			fieldMetadataMap := staticschema.FieldMetadataMapV2{
				field.Name: staticschema.FieldMetadata{
					DisplayName:  fieldNameConvertToDisplayName(field.Name),
					ValueType:    providerTypeConvertToValueType(field.Type),
					ProviderType: field.Type,
					ReadOnly:     goutils.Pointer(false),
					Values:       nil,
				},
			}

			schemas.Add(providers.ModuleGoogleMail, objectName, object.DisplayName, urlPath,
				object.ResponseKey, fieldMetadataMap, nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(files.OutputMail.FlushSchemas(schemas))
	goutils.MustBeNil(files.OutputMail.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func Objects() []metadatadef.Schema {
	explorer, err := files.InputMail.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjects(
		http.MethodGet,
		api3.AndPathMatcher{
			api3.NestedIDPathIgnorer{},
			api3.NewDenyPathStrategy(ignoreEndpoints),
		},
		nil, nil,
		func(objectName, fieldName string) bool {
			return false
		},
	)
	goutils.MustBeNil(err)

	return objects
}

func fieldNameConvertToDisplayName(fieldName string) string {
	return api3.CapitalizeFirstLetterEveryWord(
		api3.CamelCaseToSpaceSeparated(fieldName),
	)
}

func providerTypeConvertToValueType(providerType string) common.ValueType {
	switch providerType {
	case "integer":
		return common.ValueTypeInt
	case "string":
		return common.ValueTypeString
	case "boolean":
		return common.ValueTypeBoolean
	default:
		// Ex: object, array
		return common.ValueTypeOther
	}
}
