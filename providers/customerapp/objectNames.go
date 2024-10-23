package customerapp

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/providers/customerapp/metadata"
)

const (
	objectNameCollections       = "collections"
	objectNameImports           = "imports"
	objectNameReportingWebhooks = "reporting_webhooks"
	objectNameSegments          = "segments"
	objectNameSnippets          = "snippets"
	objectNameCustomerExports   = "customer_exports"
	objectNameDeliveriesExports = "deliveries_exports"
	objectNameNewsletters       = "newsletters"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = handy.NewSetFromList( //nolint:gochecknoglobals
	metadata.Schemas.GetObjectNames(),
)

var supportedObjectsByCreate = handy.NewSet( //nolint:gochecknoglobals
	objectNameCollections,
	objectNameImports,
	objectNameReportingWebhooks,
	objectNameSegments,
	objectNameSnippets, // create via PUT, which is also an update
	objectNameCustomerExports,
	objectNameDeliveriesExports,
)

var supportedObjectsByUpdate = handy.NewSet( //nolint:gochecknoglobals
	objectNameCollections,
	objectNameReportingWebhooks,
)

var supportedObjectsByDelete = handy.NewSet( //nolint:gochecknoglobals
	objectNameCollections,
	objectNameNewsletters,
	objectNameReportingWebhooks,
	objectNameSegments,
	objectNameSnippets,
)

// ObjectNameToReadResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToReadResponseField = handy.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"object_types":        "types",
	"transactional":       "messages",
	"subscription_topics": "topics",
},
	func(key string) string {
		return key
	},
)

// ObjectNameToWritePath maps ObjectName to URL path used for Write operation.
var ObjectNameToWritePath = handy.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNameCustomerExports:   "exports/customers",
	objectNameDeliveriesExports: "exports/deliveries",
},
	func(key string) string {
		// Non-special object names.
		return key
	},
)

// ObjectNameToWriteResponseField maps ObjectName to the write response field names that hold the object.
var ObjectNameToWriteResponseField = handy.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNameCollections:       "collection",
	objectNameImports:           "import",
	objectNameReportingWebhooks: "", // object is not nested, response body is a webhook object
	objectNameSegments:          "segment",
	objectNameSnippets:          "snippet",
	objectNameCustomerExports:   "export",
	objectNameDeliveriesExports: "export",
},
	func(key string) string {
		// The general pattern is response is stored under object name turned into singular form.
		// This is a fallback, although the list above is exhaustive.
		return naming.NewSingularString(key).String()
	},
)
