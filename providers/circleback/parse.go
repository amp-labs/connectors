package circleback

import (
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/spyzhov/ajson"
)

// makeNextPageFunc returns the cursor of the next page, or "" on the last page.
// The next page is announced in an RFC 8288 Link header with a path relative to the host:
//
//	Link: </api/meetings?cursor=eyJwYWdlIjoxfQ>; rel="next"
//
// Only the opaque cursor is kept; buildReadRequest sends it with the same filters as the first page.
// https://circleback.ai/docs/api#pagination
func makeNextPageFunc(response *common.JSONHTTPResponse) common.NextPageFunc {
	return func(*ajson.Node) (string, error) {
		link := httpkit.HeaderLink(response, "next")
		if link == "" {
			return "", nil
		}

		nextURL, err := url.Parse(link)
		if err != nil {
			return "", err
		}

		return nextURL.Query().Get(cursorParam), nil
	}
}
