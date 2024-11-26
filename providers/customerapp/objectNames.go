package customerapp

import (
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
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
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByCreate = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCollections,
	objectNameImports,
	objectNameReportingWebhooks,
	objectNameSegments,
	objectNameSnippets, // create via PUT, which is also an update
	objectNameCustomerExports,
	objectNameDeliveriesExports,
)

var supportedObjectsByUpdate = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCollections,
	objectNameReportingWebhooks,
)

var supportedObjectsByDelete = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCollections,
	objectNameNewsletters,
	objectNameReportingWebhooks,
	objectNameSegments,
	objectNameSnippets,
)

// ObjectNameToWritePath maps ObjectName to URL path used for Write operation.
var ObjectNameToWritePath = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNameCustomerExports:   "exports/customers",
	objectNameDeliveriesExports: "exports/deliveries",
},
	func(objectName string) (jsonPath string) {
		return objectName
	},
)

// ObjectNameToWriteResponseField maps ObjectName to the write response field names that hold the object.
var ObjectNameToWriteResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNameCollections:       "collection",
	objectNameImports:           "import",
	objectNameReportingWebhooks: "", // object is not nested, response body is a webhook object
	objectNameSegments:          "segment",
	objectNameSnippets:          "snippet",
	objectNameCustomerExports:   "export",
	objectNameDeliveriesExports: "export",
},
	func(objectName string) string {
		// The general pattern is response is stored under object name turned into singular form.
		// This is a fallback, although the list above is exhaustive.
		return naming.NewSingularString(objectName).String()
	},
)
