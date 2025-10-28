package crm

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		additional := jsonquery.New(node, "additional_data")

		nextCursor, err := additional.StringOptional("next_cursor")
		if err != nil {
			return "", err
		}

		if nextCursor == nil {
			return "", nil
		}

		url.WithQueryParam("cursor", *nextCursor)

		return url.String(), nil
	}
}
