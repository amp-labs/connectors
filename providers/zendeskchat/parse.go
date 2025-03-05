package zendeskchat

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL(objectName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		switch objectName {
		case chats:
			nextURL, err := jsonquery.New(node).StringOptional("next_url")
			if err != nil {
				return "", err
			}

			if nextURL != nil {
				return *nextURL, nil
			}

			return "", nil

		default:
			counts, err := jsonquery.New(node).IntegerOptional("count")
			if err != nil {
				return "", err
			}

			if counts == nil || *counts < defaultPageSize {
				return "", nil
			}

			nextURL, err := jsonquery.New(node).StringOptional("next_page")
			if err != nil {
				return "", err
			}

			if nextURL != nil {
				return *nextURL, nil
			}

			return "", nil
		}
	}
}
