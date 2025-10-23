package loxo

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

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

var objectWithPrefixValue = datautils.NewSet( //nolint:gochecknoglobals
	"scorecard_recommendation_types",
	"scorecard_types",
	"scorecard_templates",
	"scorecard_visibility_types ",
)
