package freshchat

import (
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var supportFilteringByTime = datautils.NewStringSet("users") //nolint: gochecknoglobals

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		records, err := jsonquery.New(node).ArrayRequired(objectName)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

func nextRecordsURL(baseURL string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextPage, err := jsonquery.New(node, "links", "next_page").StringOptional("href")
		if err != nil {
			return "", err
		}

		if nextPage == nil {
			return "", nil
		}

		res, err := url.JoinPath(baseURL, *nextPage)
		if err != nil {
			return "", err
		}

		return res, nil
	}
}
