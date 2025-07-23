package jsonquery

import (
	"encoding/json"
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

		m, ok := data.(map[string]any)
		if !ok {
			return nil, ErrNotObject
		}

		output = append(output, m)
	}

	return output, nil
}

func (convertor) ArrayToObjects(arr []*ajson.Node) ([]any, error) {
	output := make([]any, 0, len(arr))

	for _, v := range arr {
		data, err := v.Unpack()
		if err != nil {
			return nil, err
		}

		output = append(output, data)
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

func ParseNode[T any](node *ajson.Node) (*T, error) {
	var template T

	raw, err := node.Unpack()
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &template); err != nil {
		return nil, err
	}

	return &template, nil
}
