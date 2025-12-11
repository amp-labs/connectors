package outplay

import (
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func extractMetadataRecords(res map[string]any, objectName string) ([]any, error) {
	if objectName == ObjectNameCallAnalysis {
		data, ok := res["data"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("couldn't convert the response field 'data' to a map: %w", common.ErrMissingExpectedValues)
		}

		records, ok := data["data"].([]any)
		if !ok {
			return nil, fmt.Errorf("couldn't convert the nested response field 'data' to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
		}

		return records, nil
	}

	records, ok := res["data"].([]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the response field 'data' to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	return records, nil
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		hasMore, err := jsonquery.New(node, "pagination").BoolRequired("hasmorerecords")
		if err != nil {
			return "", err
		}

		if !hasMore {
			return "", nil
		}

		currentPage, err := jsonquery.New(node, "pagination").IntegerWithDefault("page", 1)
		if err != nil {
			return "", err
		}

		return strconv.Itoa(int(currentPage) + 1), nil
	}
}

func nextRecordsURLForCallAnalysis() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		paginationNode, err := jsonquery.New(node, "data").ObjectRequired("pagination")
		if err != nil {
			return "", err
		}

		hasMore, err := jsonquery.New(paginationNode).BoolRequired("hasmorerecords")
		if err != nil {
			return "", err
		}

		if !hasMore {
			return "", nil
		}

		currentPage, err := jsonquery.New(paginationNode).IntegerWithDefault("page", 1)
		if err != nil {
			return "", err
		}

		return strconv.Itoa(int(currentPage) + 1), nil
	}
}
