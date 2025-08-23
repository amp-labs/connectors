package breakcold

import "github.com/amp-labs/connectors/internal/datautils"

// The endpoints below use the POST method instead of the GET method.
// https://developer.breakcold.com/v3/api-reference/leads/list-leads-with-pagination-and-filters.
// https://developer.breakcold.com/v3/api-reference/notes/list-notes.
// https://developer.breakcold.com/v3/api-reference/reminders/list-reminders-with-filters-and-pagination.
var getEndpointsPostMethod = datautils.NewSet( //nolint:gochecknoglobals
	"leads",
	"notes",
	"reminders",
)
