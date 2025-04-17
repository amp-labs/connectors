package heyreach

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	DefaultPageSize = 100
)

// makeNextRecord creates a function that determines the next page token based on the current offset.
func makeNextRecord(offset int) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Extract the data key value from the response.
		value, err := jsonquery.New(node).ArrayRequired("items")
		if err != nil {
			return "", err
		}

		if len(value) == 0 {
			return "", nil
		}

		nextStart := offset + DefaultPageSize

		return strconv.Itoa(nextStart), nil
	}
}
