package freshdesk

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL(link string) common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {
		// if link is available, this indicates there's a next page url else there isn't.
		// link format `<https://ampersand.freshdesk.com/api/v2/contacts?per_page=1&page=2>; rel="next"`
		return ParseNexPageLinkHeader(link), nil
	}
}

// ParseNexPageLinkHeader extracts the next page URL from the Link Header response.
func ParseNexPageLinkHeader(linkHeader string) string {
	var url string

	if linkHeader == "" {
		return "" // this indicates we're done.
	}

	links := strings.Split(linkHeader, ",")
	// `<https://ampersand.freshdesk.com/api/v2/contacts?per_page=1&page=2>; rel="next"
	for _, link := range links {
		if strings.Contains(link, `rel="next"`) {
			urls := strings.Split(link, ";")
			url = strings.TrimPrefix(urls[0], "<")
			url = strings.TrimRight(url, ">")
		}
	}

	return url
}
