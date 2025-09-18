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
func getSalesforceDataMarshaller(assoc []string) func([]map[string]any, []string) ([]common.ReadResultRow, error) {
	// This is a common.MarshalFunc.
	return func(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
		data := make([]common.ReadResultRow, len(records))

		// Go through each record, attach associations (if any) to the record and
		// convert the record to a common.ReadResultRow.
		for idx, record := range records {
			recordMap := common.ToStringMap(record)
			associations := make(map[string][]common.Association)

			// Go through each associated object (from ReadParams), and extract the associations from the record.
			for _, assoc := range assoc {
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

			data[idx] = common.ReadResultRow{
				Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
				Raw:    record,
				Id:     idStr,
			}

			if len(associations) > 0 {
				data[idx].Associations = associations
			}
		}

		return data, nil
	}
}

func extractAssociationsFromRecord(val any) []common.Association {
	var result []common.Association

	assocMap, ok := val.(map[string]any) // nolint:varnamelen
	if !ok {
		return result
	}

	// In Salesforce, the associated object is a key in the record map. It appears as a nested object
	// containing a "records" array with the associated data. Additionally, it includes metadata such as
	// "done" (a boolean indicating if all records have been fetched) and the total record count.
	// There are other keys in the record map, but we only care about the "records" key for now. The other keys
	// are 'done' and number of records.
	records, ok := assocMap["records"].([]any)
	if !ok {
		return result
	}

	// For each associated record, extract the ID, convert it into a common.Association and add it to the result.
	for _, record := range records {
		if assocRec, ok := record.(map[string]any); ok {
			id, _ := common.ToStringMap(assocRec).GetCaseInsensitive("Id")
			idStr, _ := id.(string)

			association := common.Association{
				Raw: assocRec,
			}

			if idStr != "" {
				association.ObjectId = idStr
			}

			result = append(result, association)
		}
	}

	return result
}
