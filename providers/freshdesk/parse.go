package freshdesk

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL(link string) common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {
		// if link is available, this indicates there's a next page url else there isn't.
		// link format `<https://ampersand.freshdesk.com/api/v2/contacts?per_page=1&page=2>; rel="next"`
		return common.ParseNexPageLinkHeader(link), nil
	}
}
