package main

import (
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/justcall/metadata"
	"github.com/amp-labs/connectors/providers/justcall/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var ignoreEndpoints = []string{ //nolint:gochecknoglobals
	// Endpoints requiring path parameters
	"/v2.1/calls/{id}",
	"/v2.1/calls/{id}/journey",
	"/v2.1/calls/{id}/recording",
	"/v2.1/calls/{id}/voice-agent",
	"/v2.1/users/{id}",
	"/v2.1/phone-numbers/{id}",
	// Utility endpoints
	"/v2.1/phone-numbers/detect-spam",
	"/v2.1/whatsapp/messages/check-reply",
	// Analytics endpoints
	"/v2.1/calls/analytics/account",
	"/v2.1/calls/analytics/agents",
	"/v2.1/calls/analytics/numbers",
	"/v2.1/sales_dialer/analytics",
	// Scheduling endpoints
	"/v2.1/appointments/available-slots",
	// Custom fields endpoint - returns field definitions for sales_dialer/contacts.
	// This is metadata about fields, not actual data records.
	// The connector should fetch custom fields dynamically and add them to contacts metadata.
	"/v2.1/sales_dialer/contacts/custom-fields",
}

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
			replaceSlashWithSpace,
			fixAICapitalization,
		),
		api3.WithArrayItemAutoSelection(),
		api3.WithDuplicatesResolver(api3.SingleItemDuplicatesResolver(objectNameFromPath)),
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

			continue
		}

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot,
				object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}

		slog.Info("extracted object",
			"objectName", object.ObjectName,
			"displayName", object.DisplayName,
			"fields", len(object.Fields),
		)
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

// objectNameFromPath removes version prefix from URL path.
func objectNameFromPath(path string) string {
	name := strings.TrimPrefix(path, "/v2.1/")
	name = strings.TrimPrefix(name, "/")

	return name
}

// replaceSlashWithSpace replaces forward slashes with spaces in display names.
func replaceSlashWithSpace(displayName string) string {
	return strings.ReplaceAll(displayName, "/", " ")
}

// fixAICapitalization replaces "Ai" with "AI" for proper acronym capitalization.
func fixAICapitalization(displayName string) string {
	return strings.ReplaceAll(displayName, "Ai", "AI")
}
