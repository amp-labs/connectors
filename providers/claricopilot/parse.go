package claricopilot

import (
	"net/url"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL(previousURL *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pagination, err := jsonquery.New(node).ObjectOptional("pagination")
		if err != nil {
			return "", err
		}

		nextPage, err := jsonquery.New(pagination).IntegerOptional("nextPageSkip")
		if err != nil {
			return "", err
		}

		// Check this to avoid panics
		if nextPage == nil {
			return "", nil
		}

		url, err := urlbuilder.New(previousURL.String())
		if err != nil {
			return "", err
		}

		url.WithQueryParam(skipKey, strconv.FormatInt(*nextPage, 10))

		return url.String(), nil
	}
}
