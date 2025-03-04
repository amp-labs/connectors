package main

import (
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/blueshift/metadata"
	"github.com/amp-labs/connectors/providers/blueshift/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ //nolint:gochecknoglobals
		"/api/v1/data_connectors",
		"/api/v1/customer_search/show_events",
		"/api/v1/interests/user_subscriptions",
		"/api/v1/custom_user_lists/id/*",
		"/api/v1/customers/*",
		"/api/v1/catalogs/*",
		"/api/v1/email_templates/test_send.json",
		"/api/v1/sms_templates/test_send.json",
		"/api/v1/campaigns.json",
		"/api/v1/event/debug",
		"/api/v1/customers",
		"/api/v1/account_adapters",
		"/api/v1/data_connectors/:data_connector_uuid/debug",
	}

	objectEndpoints = map[string]string{ //nolint:gochecknoglobals
		"/api/v1/tag_contexts/list":     "tag_contexts/list",
		"/api/v1/segments/list":         "segments/list",
		"/api/v1/sms_templates.json":    "sms_templates",
		"/api/v1/email_templates.json":  "email_templates",
		"/api/v2/campaigns.json":        "campaigns",
		"/api/v1/external_fetches.json": "external_fetches",
		"/api/v1/push_templates.json":   "push_templates",
		"/api/v1/onsite_slots.json":     "onsite_slots",
	}

	overrideDisplayName = map[string]string{ //nolint:gochecknoglobals
		"tag_contexts/list": "Tag Contexts",
		"segments/list":     "Segments",
		"onsite_slots":      "Onsite Slots",
	}

	objectNametoResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
		"email_templates":  "results",
		"campaigns":        "results",
		"external_fetches": "results",
		"push_templates":   "results",
		"sms_templates":    "results",
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

		// Remove /api prefix from URLPath
		urlPath := strings.TrimPrefix(object.URLPath, "/api")

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName, urlPath, object.ResponseKey,
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
