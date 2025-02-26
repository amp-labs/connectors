package blueshift

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// var supportedObjectsByRead = metadata.Schemas.ObjectNames()

var ObjectNametoResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
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
