package hunter

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		records, err := jsonquery.New(node, "data").ArrayOptional(objectName)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		meta := jsonquery.New(node, metaField)

		params, err := meta.ObjectOptional(paramsField)
		if err != nil {
			return "", err
		}

		total, err := meta.IntegerOptional(totalField)
		if err != nil {
			return "", err
		}

		offset, err := jsonquery.New(params).IntegerOptional(offsetKey)
		if err != nil {
			return "", err
		}

		if total == nil || offset == nil {
			return "", nil
		}

		if (pageSize + int(*offset)) < int(*total) {
			return strconv.Itoa(pageSize + int(*offset)), nil
		}

		return "", nil
	}
}
