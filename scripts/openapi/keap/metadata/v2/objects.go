package keapv1

import (
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/providers/keap/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		// endpoints for creating fields
		"/v2/notes/model/customFields",
		"/v2/tasks/model/customFields",
		// custom fields and models endpoints to create them are not read candidates
		"/v2/affiliates/model", // array located at "custom_fields"
		"/v2/contacts/model",   // array located at "custom_fields"
		"/v2/notes/model",      // array located at "custom_fields"
		"/v2/tasks/model",      // array located at "custom_fields"
		// singular object
		"/v2/businessProfile",
		// misc
		"/v2/settings/contactOptionTypes", // list of strings not objects
		"/v2/settings/applications:isEnabled",
		"/v2/settings/applications:getConfiguration",
	}
	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		"/v2/tags/categories":      "tags/categories",
		"/v2/contacts/links/types": "contacts/links/types",
		"/v2/automationCategory":   "automationCategory",
	}
	objectNameToResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
		"tags/categories":      "tag_categories",
		"contacts/links/types": "contact_link_types",
		"automationCategory":   "automation_categories",
	},
		func(objectName string) (fieldName string) {
			return objectName
		},
	)
)

func Objects() []metadatadef.Schema {
	explorer, err := openapi.Version2FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			func(displayName string) string {
				displayName, _ = strings.CutPrefix(displayName, "List")
				displayName, _ = strings.CutSuffix(displayName, "Response")

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
		objectEndpoints, nil,
		api3.CustomMappingObjectCheck(objectNameToResponseField),
	)
	goutils.MustBeNil(err)

	return objects
}
