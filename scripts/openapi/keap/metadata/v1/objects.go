package keapv1

import (
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers/keap/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

// nolint:lll
var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		// endpoints for creating fields
		"/v1/appointments/model/customFields",
		"/v1/notes/model/customFields",
		"/v1/tasks/model/customFields",
		// custom fields are not read candidates, but its response should be used to enhance metadata
		"/v1/affiliates/model",    // array located at "custom_fields"
		"/v1/appointments/model",  // array located at "custom_fields"
		"/v1/companies/model",     // array located at "custom_fields"
		"/v1/contacts/model",      // array located at "custom_fields"
		"/v1/notes/model",         // array located at "custom_fields"
		"/v1/opportunities/model", // array located at "custom_fields"
		"/v1/orders/model",        // array located at "custom_fields"
		"/v1/subscriptions/model", // array located at "custom_fields"
		"/v1/tasks/model",         // array located at "custom_fields"
		// duplicates
		"/v1/tasks/search", // covered by "/tasks"
		// requires query parameters
		"/v1/affiliates/summaries", // https://developer.infusionsoft.com/docs/rest/#tag/Affiliate/operation/listSummariesUsingGET
		"/v1/products/sync",        // additionally, it is deprecated: https://developer.infusionsoft.com/docs/rest/#tag/Product/operation/listProductsFromSyncTokenUsingGET
		// not applicable
		"/v1/setting/application/enabled",       // retrieves application status
		"/v1/setting/application/configuration", // retrieves application configuration
		"/v1/account/profile",                   // retrieves account profile
		"/v1/oauth/connect/userinfo",            // retrieves User Info
		"/v1/locales/defaultOptions",            // dropdown of default options
		"/v1/setting/contact/optionTypes",       // comma separated "list" of strings
		"/v1/locales/countries",                 // countries is an object not array
	}
	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		// "/v1/products/sync": "synced_products",
	}
	displayNameOverride = map[string]string{ // nolint:gochecknoglobals
		"commissions": "Commissions",
		"emails":      "Emails",
		"merchants":   "Merchants",
		"programs":    "Programs",
		// "summaries":   "Summaries",
	}
	objectNameToResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
		"merchants":     "merchant_accounts",
		"redirectlinks": "redirects",
		// "synced_products": "product_statuses",
	},
		func(objectName string) (fieldName string) {
			return objectName
		},
	)
)

func Objects() []api3.Schema {
	explorer, err := openapi.Version1FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			func(displayName string) string {
				displayName, _ = strings.CutSuffix(displayName, "List")

				return displayName
			},
			api3.Pluralize,
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, displayNameOverride,
		api3.CustomMappingObjectCheck(objectNameToResponseField),
	)
	goutils.MustBeNil(err)

	return objects
}
