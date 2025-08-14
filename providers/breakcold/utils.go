package breakcold

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	pageSize                = 100
	objectNameRemindersList = "reminders"
	objectNameLeadsList     = "leads"
)

// The endpoints below use the POST method instead of the GET method.
// https://developer.breakcold.com/v3/api-reference/leads/list-leads-with-pagination-and-filters.
// https://developer.breakcold.com/v3/api-reference/notes/list-notes.
// https://developer.breakcold.com/v3/api-reference/reminders/list-reminders-with-filters-and-pagination.
var getEndpointsPostMethod = datautils.NewSet( //nolint:gochecknoglobals
	"leads",
	"notes",
	"reminders",
)

func makeNextRecordsURL(nodePath string, nextPage int) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		query := jsonquery.New(node)

		record, err := query.ArrayOptional(nodePath)
		if err != nil {
			return "", err
		}

		if len(record) < pageSize {
			return "", nil
		}

		nextPage += 1

		return strconv.Itoa(nextPage), nil
	}
}
