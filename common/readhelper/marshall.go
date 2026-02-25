package readhelper

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// RowMarshallProcessor modifies the results returned by a MarshalFromNodeFunc.
type RowMarshallProcessor func([]common.ReadResultRow) error

// ChainedMarshaller wraps a base marshaller with one or more row processors.
func ChainedMarshaller(
	base common.MarshalFromNodeFunc,
	processors ...RowMarshallProcessor,
) common.MarshalFromNodeFunc {
	return func(nodes []*ajson.Node, fields []string) ([]common.ReadResultRow, error) {
		data, err := base(nodes, fields)
		if err != nil {
			return nil, err
		}

		for _, p := range processors {
			if err := p(data); err != nil {
				return nil, err
			}
		}

		return data, nil
	}
}
