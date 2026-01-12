package shared //nolint:revive

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func GetNextPageURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// The response is a JSON object with a "links" property.
		// The "links" property is an array of objects with a "rel" property and a "href" property.
		// We need to find the "next" link and return the "href" property.
		links, err := jsonquery.New(node).ArrayRequired("links")
		if err != nil {
			return "", err
		}

		for _, link := range links {
			rel, err := jsonquery.New(link).StringOptional("rel")
			if err != nil {
				return "", err
			}

			if rel != nil && *rel == "next" {
				return jsonquery.New(link).StringRequired("href")
			}
		}

		return "", nil
	}
}
