package main

import (
	"log"
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/constantcontact/metadata"
	"github.com/amp-labs/connectors/providers/constantcontact/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		"/account/summary",                  // returns single customer
		"/account/summary/physical_address", // returns single address
		"/contacts/counts",                  // object describing statistics on contacts
		// Requires query parameters
		"/emails/campaign_id_xrefs",
		"/contacts/contact_id_xrefs",
		"/contact_lists/list_id_xrefs",
		// Partner accounts need auth token to be in 2 headers, usual auth and custom x-api-key.
		"/partner/*",
	}
	objectEndpoints = map[string]string{
		"/account/emails": "account_emails",
		"/emails":         "email_campaigns",
	}
	displayNameOverride = map[string]string{
		// "campaign_id_xrefs": "Email Campaign Identifiers",
		// "contact_id_xrefs":  "Contact Identifiers",
		// "list_id_xrefs":     "List Identifiers",
	}
	objectNameToResponseField = datautils.NewDefaultMap(map[string]string{
		// "campaign_id_xrefs":        "xrefs",
		// "contact_id_xrefs":         "xrefs",
		// "list_id_xrefs":            "xrefs",
		"accounts":                 "site_owner_list",
		"email_campaign_summaries": "bulk_email_campaign_summaries",
		"contact_tags":             "tags",
		"contact_lists":            "lists",
		"contact_custom_fields":    "custom_fields",
		"email_campaigns":          "campaigns",
		"account_emails":           "", // response is already an array, empty refers to current
		"privileges":               "",
		"subscriptions":            "",
	},
		func(objectName string) (fieldName string) {
			return objectName
		},
	)
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		))
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, displayNameOverride,
		api3.CustomMappingObjectCheck(objectNameToResponseField),
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV1]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				staticschema.FieldMetadataMapV1{
					field: field,
				}, nil)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	log.Println("Completed.")
}
