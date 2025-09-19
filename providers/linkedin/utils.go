package linkedin

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

func makeNextRecord() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return "", nil
	}
}
