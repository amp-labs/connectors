package brevo

import "github.com/amp-labs/connectors/internal/datautils"

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{
	"contacts/attributes":                 "attributes",
	"companies/attributes":                "",
	"inbound/events":                      "events",
	"transactionalSMS/statistics/events":  "events",
	"smtp/statistics/events":              "events",
	"smtp/statistics/reports":             "reports",
	"transactionalSMS/statistics/reports": "reports",
},
	func(objectName string) (fieldName string) {
		return objectName
	},
)
