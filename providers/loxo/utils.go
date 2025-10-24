package loxo

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// The objectNodePath variable maps each object to its corresponding nodePath if present in the response.
// If the response does not contain a nodePath, it returns an empty value, indicating that the object has no nodePath.
// See Example by using the link https://github.com/amp-labs/connectors/pull/2126#discussion_r2454221609
var objectsNodePath = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"companies":      "companies",
	"countries":      "countries",
	"deals":          "deals",
	"email_tracking": "tracking",
	"form_templates": "form_templates",
	"forms":          "forms",
	"jobs":           "results",
	"people":         "people",
	"person_events":  "person_events",
	"placements":     "placements",
	"schedule_items": "schedule_items",
	"scorecards":     "scorecards",
	"sms":            "sms",
	"source_types":   "source_types",
}, func(objectName string) string {
	return ""
},
)
