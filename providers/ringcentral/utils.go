package ringcentral

import (
	"fmt"
	"reflect"

	"github.com/amp-labs/connectors/common"
)

type ObjectsOperationURLs struct {
	ReadPath             string `json:"read_path"`
	WritePath            string `json:"write_path"`
	RecordsField         string `json:"records_field"`
	UsesCursorPagination bool   `json:"uses_cursor_pagination"`
	UsesOffsetPagination bool   `json:"uses_offset_pagination"`
	UsesSyncToken        bool   `json:"uses_sync_token"`
}

func GetFieldByJSONTag(resp *Response, jsonTag string) ([]map[string]any, error) {
	v := reflect.ValueOf(resp).Elem()
	t := v.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		tag := field.Tag.Get("json")

		// Match the JSON tag
		if tag == jsonTag {
			fieldValue := v.Field(i)

			if fieldValue.Kind() != reflect.Slice {
				return nil, fmt.Errorf("field with tag '%s' is not a slice", jsonTag) //nolint: err113
			}

			result, ok := fieldValue.Interface().([]map[string]any)
			if !ok {
				return nil, fmt.Errorf("field with tag '%s' is not of type []map[string]any", jsonTag) //nolint: err113
			}

			return result, nil
		}
	}

	return nil, fmt.Errorf("field with json tag '%s' not found", jsonTag) // nolint: err113
}

func inferValue(value any) common.ValueType {
	v := reflect.ValueOf(value)

	switch v.Kind() { //nolint: exhaustive
	case reflect.String:
		return common.ValueTypeString
	case reflect.Float64:
		return common.ValueTypeFloat
	case reflect.Bool:
		return common.ValueTypeBoolean
	case reflect.Slice:
		return common.ValueTypeOther
	case reflect.Map:
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}
