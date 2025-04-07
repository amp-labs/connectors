package dixa

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func constructRecords(objectName string) common.RecordsFunc {
	switch objectName {
	case businessHours:
		return func(node *ajson.Node) ([]map[string]any, error) {
			schedules, err := jsonquery.New(node, data).ArrayOptional("schedules")
			if err != nil {
				return nil, err
			}

			return jsonquery.Convertor.ArrayToMap(schedules)
		}
	default:
		return common.GetRecordsUnderJSONPath(data)
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		meta, err := jsonquery.New(node).ObjectOptional("meta")
		if err != nil {
			return "", err
		}

		nextCrs, err := jsonquery.New(meta).StringOptional("next")
		if err != nil {
			return "", err
		}

		if nextCrs == nil {
			return "", err
		}

		return *nextCrs, nil
	}
}
