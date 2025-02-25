package blueshift

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// var supportedObjectsByRead = metadata.Schemas.ObjectNames()

var ObjectNametoResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"email_templates":   "template",
	"sms_templates":     "template",
	"campaigns":         "campaigns",
	"external_fetches":  "template",
	"push_templates":    "template",
	"segments/list":     "segments",
	"tag_contexts/list": "list",
},
	func(objectName string) (fieldName string) {
		return objectName
	},
)
