package constantcontact

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/constantcontact/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"campaign_id_xrefs":        "xrefs",
	"contact_id_xrefs":         "xrefs",
	"list_id_xrefs":            "xrefs",
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
