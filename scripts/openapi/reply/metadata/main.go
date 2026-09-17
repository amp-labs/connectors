// Generates providers/reply/metadata/schemas.json from the Reply V3 OpenAPI spec.
// https://docs.reply.io/api-reference/bundled.yaml
//
// Usage:
//
//	go run ./scripts/openapi/reply/metadata
package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/reply/metadata"
	"github.com/amp-labs/connectors/providers/reply/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	// Entity objects exposed by the connector. Action endpoints (/start, /set-*,
	// bulk operations) and analytics surfaces (ai-sdr, live-data, reporting) are
	// deliberately excluded.
	objectEndpoints = map[string]string{
		"/v3/contacts":               "contacts",
		"/v3/contact-accounts":       "contact-accounts",
		"/v3/contact-lists":          "contact-lists",
		"/v3/contact-account-lists":  "contact-account-lists",
		"/v3/sequences":              "sequences",
		"/v3/sequence-folders":       "sequence-folders",
		"/v3/tasks":                  "tasks",
		"/v3/email-templates":        "email-templates",
		"/v3/email-template-folders": "email-template-folders",
		"/v3/email-accounts":         "email-accounts",
		"/v3/custom-fields":          "custom-fields",
		"/v3/schedules":              "schedules",
		"/v3/linkedin-accounts":      "linkedin-accounts",
		"/v3/inbox/threads":          "inbox/threads",
	}

	displayNameOverride = map[string]string{
		"contacts":               "Contacts",
		"contact-accounts":       "Contact Accounts",
		"contact-lists":          "Contact Lists",
		"contact-account-lists":  "Contact Account Lists",
		"sequences":              "Sequences",
		"sequence-folders":       "Sequence Folders",
		"tasks":                  "Tasks",
		"email-templates":        "Email Templates",
		"email-template-folders": "Email Template Folders",
		"email-accounts":         "Email Accounts",
		"custom-fields":          "Custom Fields",
		"schedules":              "Schedules",
		"linkedin-accounts":      "LinkedIn Accounts",
		"inbox/threads":          "Inbox Threads",
	}

	// Most list endpoints wrap records in {"items": [...], "hasMore": bool}.
	// The exceptions below return a root-level array.
	objectNameToResponseField = datautils.NewDefaultMap(map[string]string{
		"custom-fields":          "",
		"schedules":              "",
		"sequence-folders":       "",
		"email-template-folders": "",
		"linkedin-accounts":      "",
	},
		func(objectName string) string {
			return "items"
		},
	)
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)

	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewAllowPathStrategy(datautils.Map[string, string](objectEndpoints).Keys()),
		objectEndpoints, displayNameOverride,
		api3.CustomMappingObjectCheck(objectNameToResponseField),
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
