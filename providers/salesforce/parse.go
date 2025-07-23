package salesforce

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	records, err := jsonquery.New(node).ArrayRequired("records")
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(records)
}

// getNextRecordsURL returns the URL for the next page of results.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("nextRecordsUrl", "")
}

// getSalesforceDataMarshaller returns a marshaller that fills Associations in ReadResultRow for Salesforce.
func getSalesforceDataMarshaller(associatedObjects []string) func([]map[string]any, []string) ([]common.ReadResultRow, error) {
	// This is a common.MarshalFunc.
	return func(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
		data := make([]common.ReadResultRow, len(records))

		// Go through each record, attach associations (if any) to the record and
		// convert the record to a common.ReadResultRow.
		for i, record := range records {
			recordMap := common.ToStringMap(record)
			associations := make(map[string][]common.Association)

			// Go through each associated object (from ReadParams), and extract the associations from the record.
			for _, assoc := range associatedObjects {
				// In Salesforce, the associated object is a key in the record map.
				// For example, "Contacts" will be a key in the record map, with an array of associated contacts.
				key, ok := recordMap.GetCaseInsensitive(assoc)
				if !ok {
					continue
				}

				assocList := extractAssociationsFromRecord(key)
				if len(assocList) > 0 {
					associations[assoc] = assocList
				}
			}

			// Extract the ID of the record.
			id, _ := recordMap.GetCaseInsensitive("Id")
			idStr, _ := id.(string)

			data[i] = common.ReadResultRow{
				Fields:       common.ExtractLowercaseFieldsFromRaw(fields, record),
				Raw:          record,
				Associations: associations,
				Id:           idStr,
			}
		}

		return data, nil
	}
}

func extractAssociationsFromRecord(val any) []common.Association {
	var result []common.Association

	assocMap, ok := val.(map[string]any)
	if !ok {
		return result
	}

	// In Salesforce, the associated object is a key in the record map, with an array of associated records.
	// There are other keys in the record map, but we only care about the "records" key for now. The other keys
	// are 'done' and number of records.
	records, ok := assocMap["records"].([]any)
	if !ok {
		return result
	}

	// For each associated record, extract the ID, coax it into a common.Association & add it to the result.
	for _, record := range records {
		if assocRec, ok := record.(map[string]any); ok {
			id, ok := common.ToStringMap(assocRec).GetCaseInsensitive("Id")
			if !ok {
				continue
			}

			idStr, ok := id.(string)
			if !ok {
				continue
			}

			if idStr != "" {
				result = append(result, common.Association{
					ObjectId: idStr,
					Raw:      assocRec,
				})
			}
		}
	}

	return result
}
