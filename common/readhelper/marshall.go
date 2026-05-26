package readhelper

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// RowPostProcessor modifies the results returned by a MarshalFromNodeFunc.
type RowPostProcessor func([]common.ReadResultRow) error

// ChainedMarshaller wraps a base marshaller with one or more row processors.
func ChainedMarshaller(
	base common.MarshalFromNodeFunc,
	processors ...RowPostProcessor,
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
