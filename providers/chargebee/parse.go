package chargebee

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func extractRecords(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		records, err := jsonquery.New(node).ArrayRequired("list")
		if err != nil {
			return nil, err
		}

		result := make([]map[string]any, 0, len(records))
		objectResponseKey := objectResponseField.Get(objectName)

		// Loop through each record in the response list to extract the actual object data.
		// Chargebee API responses can have objects nested under a specific key (e.g., "subscription", "customer")
		// https://apidocs.chargebee.com/docs/api/customers?lang=curl#list_customers
		for _, record := range records {
			objectData, err := jsonquery.New(record).ObjectOptional(objectResponseKey)
			if err == nil && objectData != nil {
				recordMap, err := jsonquery.Convertor.ObjectToMap(objectData)
				if err != nil {
					return nil, err
				}

				result = append(result, recordMap)

				continue
			}

			// If no nested objects, use record itself
			recordMap, err := jsonquery.Convertor.ObjectToMap(record)
			if err != nil {
				return nil, err
			}

			result = append(result, recordMap)
		}

		return result, nil
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextOffset, err := jsonquery.New(node).StringOptional("next_offset")
		if err != nil {
			return "", err
		}

		if nextOffset == nil || *nextOffset == "" {
			return "", nil
		}

		return *nextOffset, nil
	}
}
