package httpkit

import (
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/amp-labs/connectors/common"
	lh "github.com/deiu/linkparser"
)

// HeaderLink extracts and parses Link Header of HTTP response.
//
// nolint:lll
// Example: given the header below we want to return one of the URLs based on the relationship name.
// <https://api.capsulecrm.com/api/v2/parties?page=3&perPage=1>; rel="next", <https://api.capsulecrm.com/api/v2/parties?page=1&perPage=1>; rel="prev"
//
// The implementation is delegated to `linkparser` library.
func HeaderLink(resp *common.JSONHTTPResponse, relationshipName string) string {
	link := resp.Headers.Get("Link")

	return lh.ParseHeader(link)[relationshipName]["href"]
}

func Status2xx(code int) bool {
	return 200 <= code && code < 300
}

func Status4xx(code int) bool {
	return 400 <= code && code < 500
}

func Status5xx(code int) bool {
	return 500 <= code && code < 600
}

func ExtractHeader(headers http.Header, name string) (string, error) {
	if headers == nil {
		return "", fmt.Errorf("%w: header '%v'", common.ErrMissingHeader, name)
	}

	if value := headers.Get(name); value != "" {
		return strings.TrimSpace(value), nil
	}

	canonicalName := textproto.CanonicalMIMEHeaderKey(name)
	if values, ok := headers[canonicalName]; ok && len(values) > 0 {
		return strings.TrimSpace(values[0]), nil
	}

	if values, ok := headers[strings.ToLower(name)]; ok && len(values) > 0 {
		return strings.TrimSpace(values[0]), nil
	}

	return "", fmt.Errorf("%w: header '%v'", common.ErrMissingHeader, name)
}
