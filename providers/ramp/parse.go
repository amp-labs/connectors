package ramp

import (
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// makeNextRecordsURL extracts the next page URL from page.next in the response.
// Ramp returns the full next-page URL or null when there are no more pages.
func makeNextRecordsURL(node *ajson.Node) (string, error) {
	nextURL, err := jsonquery.New(node, "page").StringOptional("next")
	if err != nil || nextURL == nil {
		return "", nil //nolint:nilerr
	}

	return *nextURL, nil
}
