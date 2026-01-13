package fathom

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func addMeetingsQueryParams(url *urlbuilder.URL) {
	url.WithQueryParam("include_action_items", "true")
	url.WithQueryParam("include_crm_matches", "true")
	url.WithQueryParam("include_summary", "true")
	url.WithQueryParam("include_transcript", "true")

}
