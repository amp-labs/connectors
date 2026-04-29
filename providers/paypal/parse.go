package paypal

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// nextRecordsURL extracts the next page URL from PayPal's HATEOAS links array.
// PayPal list responses include a root "links" array; the element with rel=="next"
// carries the full absolute URL for the next page.
func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		links, err := jsonquery.New(node).ArrayOptional("links")
		if err != nil || len(links) == 0 {
			return "", nil //nolint:nilerr
		}

		for _, link := range links {
			rel, err := jsonquery.New(link).StringOptional("rel")
			if err != nil || rel == nil || *rel != "next" {
				continue
			}

			href, err := jsonquery.New(link).StringOptional("href")
			if err != nil || href == nil {
				return "", nil //nolint:nilerr
			}

			return *href, nil
		}

		return "", nil
	}
}
