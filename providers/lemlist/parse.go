package lemlist

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		schema, fld := responseSchema(objectName)

		switch schema {
		case object:
			if fld != "" {
				rcds, err := jsonquery.New(node).ArrayOptional(fld)
				if err != nil {
					return nil, err
				}

				return jsonquery.Convertor.ArrayToMap(rcds)
			}

			record, err := jsonquery.Convertor.ObjectToMap(node)

			return []map[string]any{record}, err

		default:
			return common.ExtractRecordsFromPath(fld)(node)
		}
	}
}

func nextRecordsURL(objectName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		schema, _ := responseSchema(objectName)
		switch schema {
		case object:
			paginationQuery := jsonquery.New(node, "pagination")

			currentPage, err := paginationQuery.IntegerOptional("currentPage")
			if err != nil {
				return "", err
			}

			totalPage, err := paginationQuery.IntegerOptional("totalPage")
			if err != nil {
				return "", err
			}

			nextPage, err := paginationQuery.IntegerOptional("nextPage")
			if err != nil {
				return "", err
			}

			if nextPage == nil || currentPage == nil || totalPage == nil {
				return "", nil
			}

			if *currentPage < *totalPage {
				return strconv.Itoa(int(*nextPage)), nil
			}
		default:
			return "", nil
		}

		return "", nil
	}
}
