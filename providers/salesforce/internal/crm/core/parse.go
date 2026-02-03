package core

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/associations"
	"github.com/spyzhov/ajson"
)

// GetRecords returns the records from the response.
func GetRecords(node *ajson.Node) ([]map[string]any, error) {
	records, err := jsonquery.New(node).ArrayRequired("records")
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(records)
}

// GetNextRecordsURL returns the URL for the next page of results.
func GetNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("nextRecordsUrl", "")
}

// GetDataMarshallerForSearch returns a marshaller that fills Associations in ReadResultRow for Salesforce.
func GetDataMarshallerForSearch(params *common.SearchParams) common.MarshalFunc {
	return getDataMarshaller(params.ObjectName, params.AssociatedObjects)
}

// GetDataMarshallerForRead returns a marshaller that fills Associations in ReadResultRow for Salesforce.
func GetDataMarshallerForRead(params common.ReadParams) common.MarshalFunc {
	return getDataMarshaller(params.ObjectName, params.AssociatedObjects)
}

func getDataMarshaller(objectName string, associatedObjects []string) common.MarshalFunc {
	return func(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
		data := make([]common.ReadResultRow, len(records))

		// Go through each record, attach associations (if any) to the record and
		// convert the record to a common.ReadResultRow.
		for idx, record := range records {
			recordMap := common.ToStringMap(record)

			// Extract the ID of the record.
			id, _ := recordMap.GetCaseInsensitive("Id")
			idStr, _ := id.(string)

			data[idx] = common.ReadResultRow{
				Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
				Raw:    record,
				Id:     idStr,
			}

			asts := associations.ExtractAssociations(recordMap, objectName, associatedObjects)

			if len(asts) > 0 {
				data[idx].Associations = asts
			}
		}

		return data, nil
	}
}
