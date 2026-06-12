package fathom

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// addMeetingsQueryParams adds query parameters to enrich the meetings API response
// with additional data fields. These parameters ensure we retrieve complete meeting
// information including action items, CRM matches.
func addMeetingsQueryParams(url *urlbuilder.URL) {
	url.WithQueryParam("include_action_items", "true")
	url.WithQueryParam("include_crm_matches", "true")
}
