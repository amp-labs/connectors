package drift

import (
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

			rcd, err := jsonquery.Convertor.ObjectToMap(node)
			if err != nil {
				return nil, err
			}

			return []map[string]any{rcd}, nil

		default:
			return common.ExtractRecordsFromPath(fld)(node)
		}
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {
		links, err := jsonquery.New(n).ObjectOptional("links")
		if err != nil {
			return "", err
		}

		next, err := jsonquery.New(links).StringOptional("next")
		if err != nil {
			return "", err
		}

		if next == nil {
			return "", nil
		}

		return *next, nil
	}
}
