package jsonquery

import "github.com/spyzhov/ajson"

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
