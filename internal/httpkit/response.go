// nolint:lll
package httpkit

import (
	"github.com/amp-labs/connectors/common"
	lh "github.com/deiu/linkparser"
)

// HeaderLink extracts and parses Link Header of HTTP response.
//
// Example: given the header below we want to return one of the URLs based on the relationship name.
// <https://api.capsulecrm.com/api/v2/parties?page=3&perPage=1>; rel="next", <https://api.capsulecrm.com/api/v2/parties?page=1&perPage=1>; rel="prev"
//
// The implementation is delegated to `linkparser` library.
func HeaderLink(resp *common.JSONHTTPResponse, relationshipName string) string {
	link := resp.Headers.Get("Link")

	return lh.ParseHeader(link)[relationshipName]["href"]
}
