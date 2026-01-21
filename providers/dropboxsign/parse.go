package dropboxsign

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		totalPages, err := jsonquery.New(node, "list_info").IntegerRequired("num_pages")
		if err != nil {
			return "", err
		}

		currentPage, err := jsonquery.New(node, "list_info").IntegerRequired("page")
		if err != nil {
			return "", err
		}

		if currentPage >= totalPages {
			return "", nil
		}

		next := strconv.Itoa(int(currentPage + 1))

		return next, nil
	}
}
