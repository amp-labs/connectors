package jsonquery

import (
	"encoding/json"
	"errors"
	"fmt"

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

func convertMapToAjsonNode(jsonMap map[string]any) (result *ajson.Node, errOut error) {
	defer func() {
		if errOut != nil {
			errOut = fmt.Errorf("%w:%w", ErrConversionFailed, errOut)
		}
	}()

	raw, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return ajson.Unmarshal(raw)
}

// ParseNode attempts to convert the input JSON representation (either *ajson.Node or map[string]any)
// into a concrete Go struct of type T. It handles both raw ajson nodes and generic maps by first
// converting maps into ajson nodes before unmarshalling.
func ParseNode[T any, J jsonType](jsonData J) (*T, error) {
	switch data := any(jsonData).(type) {
	case *ajson.Node:
		return parseJSONNode[T](data)
	case map[string]any:
		node, err := convertMapToAjsonNode(data)
		if err != nil {
			return nil, err
		}

		return parseJSONNode[T](node)
	default:
		return nil, ErrUnknownJSONRepresentation
	}
}

func parseJSONNode[T any](node *ajson.Node) (*T, error) {
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
