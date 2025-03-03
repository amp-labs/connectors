package blueshift

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/blueshift/metadata"
	"github.com/spyzhov/ajson"
)

const (
	pageSizeKey = "per_page"
	pageSize    = "200"
	pageKey     = "page"
	pageNumber  = "0"
)

func getRecords(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		responseKey := metadata.Schemas.LookupArrayFieldName(staticschema.RootModuleID, objectName)

		rcds, err := jsonquery.New(node).ArrayOptional(responseKey)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(rcds)
	}
}

func makeNextRecordsURL(baseURL string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		totalPages, err := jsonquery.New(node).IntegerOptional("total_pages")
		if err != nil {
			return "", err
		}

		currentPage, err := jsonquery.New(node).IntegerOptional("page")
		if err != nil {
			return "", err
		}

		if totalPages == nil || currentPage == nil || *currentPage >= *totalPages-1 {
			return "", nil
		}

		nextURL := fmt.Sprintf("%s?page=%d&per_page=%s", baseURL, int(*currentPage+1), pageSize)

		return nextURL, nil
	}
}
