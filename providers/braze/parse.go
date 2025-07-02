package braze

import "github.com/amp-labs/connectors/internal/datautils"

// These were retrieved these from their API reference documentation,
// specifically from the response samples in the Endpoints section of their respective object APIs.
// https://www.braze.com/docs/api/home
var dataFields = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	"catalogs":                      "catalogs",
	"cdi/integrations":              "results",
	"email/hard_bounces":            "emails",
	"email/unsubscribes":            "emails",
	"campaigns/list":                "campaigns",
	"canvas/list":                   "canvases",
	"events/list":                   "events",
	"events":                        "events",
	"kpi/new_users/data_series":     "data",
	"kpi/dau/data_series":           "data",
	"kpi/mau/data_series":           "data",
	"kpi/uninstalls/data_series":    "data",
	"purchases/product_list":        "products",
	"purchases/revenue_series":      "data",
	"purchases/quantity_series":     "data",
	"segments/list":                 "segments",
	"sessions/data_series":          "data",
	"custom_attributes":             "attributes",
	"sms/invalid_phone_numbers":     "sms",
	"messages/scheduled_broadcasts": "scheduled_broadcasts",
	"preference_center/v1/list":     "preference_centers",
	"content_blocks/list":           "content_blocks",
	"templates/email/list":          "templates",
}, func(key string) string {
	return "data"
})
