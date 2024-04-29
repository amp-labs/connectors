package microsoftdynamicscrm

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

func getTotalSize(node *ajson.Node) (int64, error) {
	return common.JSONManager.ArrSize(node, "value")
}

func getRecords(node *ajson.Node) ([]map[string]any, error) {
	arr, err := common.JSONManager.GetArr(node, "value")
	if err != nil {
		return nil, err
	}

	return common.JSONManager.ArrToMap(arr)
}

func getNextRecordsURL(node *ajson.Node) (string, error) {
	return common.JSONManager.GetStringWithDefault(node, "@odata.nextLink", "")
}

func getMarshaledData(records []map[string]interface{}, fields []string) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	for i, record := range records {
		data[i] = common.ReadResultRow{
			Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
		}
	}

	return data, nil
}
