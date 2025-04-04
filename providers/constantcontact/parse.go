package constantcontact

import (
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) makeNextRecordsURL(node *ajson.Node) (string, error) {
	href, err := jsonquery.New(node, "_links", "next").StrWithDefault("href", "")
	if err != nil {
		return "", err
	}

	if len(href) == 0 {
		// Next page doesn't exist
		return "", nil
	}

	url, err := c.RootClient.URL()
	if err != nil {
		return "", err
	}

	fullURL := url.String() + href

	return fullURL, nil
}
