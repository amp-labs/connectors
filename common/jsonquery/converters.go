package jsonquery

import (
	"errors"

	"github.com/spyzhov/ajson"
)

// Convertor provides common conversion methods from ajson to go types.
var Convertor = convertor{} //nolint:gochecknoglobals

type convertor struct{}

func (convertor) ArrayToMap(arr []*ajson.Node) ([]map[string]any, error) {
	output := make([]map[string]any, 0, len(arr))

	for _, v := range arr {
		if !v.IsObject() {
			return nil, ErrNotObject
		}

		data, err := v.Unpack()
		if err != nil {
			return nil, err
		}

		m, ok := data.(map[string]interface{})
		if !ok {
			return nil, ErrNotObject
		}

		output = append(output, m)
	}

	return output, nil
}

func (convertor) ObjectToMap(node *ajson.Node) (map[string]any, error) {
	data, err := node.GetObject()
	if err != nil {
		return nil, err
	}

	result := make(map[string]any)
	for k, v := range data {
		result[k], err = v.Unpack()
		if err != nil {
			return nil, errors.Join(err, ErrUnpacking)
		}
	}

	return result, nil
}
