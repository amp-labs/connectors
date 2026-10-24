package circleback

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// actionItemStatusParam is the completion filter of GET /action-items, which defaults to
// incomplete items. The docs list no values for it, so action items are read once per status
// an action item can have (PENDING, then DONE), as the action item's own `status` enum names them.
// https://circleback.ai/docs/api/action-items/list-action-items
const actionItemStatusParam = "status"

var actionItemStatuses = []string{"PENDING", "DONE"} // nolint:gochecknoglobals

// actionItemsPage is where an action items read resumes: the status being read and the
// provider's cursor within it. The next page token is "<status>:<cursor>"; an empty cursor
// starts that status from its first page, and an empty token starts the first status.
type actionItemsPage struct {
	status string
	cursor string
}

func (p actionItemsPage) token() string {
	return p.status + ":" + p.cursor
}

func parseActionItemsPage(token common.NextPageToken) (actionItemsPage, error) {
	if token == "" {
		return actionItemsPage{status: actionItemStatuses[0]}, nil
	}

	status, cursor, found := strings.Cut(token.String(), ":")
	if !found || !slices.Contains(actionItemStatuses, status) {
		return actionItemsPage{}, fmt.Errorf("%w: action items page %q", common.ErrNextPageInvalid, token)
	}

	return actionItemsPage{status: status, cursor: cursor}, nil
}

// makeActionItemsNextPageFunc continues within the status the request read while its Link header
// has a next page, then moves on to the next status, and ends after the last one.
func makeActionItemsNextPageFunc(requestURL *url.URL, response *common.JSONHTTPResponse) common.NextPageFunc {
	cursorWithinStatus := makeNextPageFunc(response)

	return func(node *ajson.Node) (string, error) {
		status := requestURL.Query().Get(actionItemStatusParam)

		cursor, err := cursorWithinStatus(node)
		if err != nil {
			return "", err
		}

		if cursor != "" {
			return actionItemsPage{status: status, cursor: cursor}.token(), nil
		}

		index := slices.Index(actionItemStatuses, status)
		if index < 0 || index == len(actionItemStatuses)-1 {
			return "", nil
		}

		return actionItemsPage{status: actionItemStatuses[index+1]}.token(), nil
	}
}
