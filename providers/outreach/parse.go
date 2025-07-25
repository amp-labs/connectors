package outreach

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func getRecords(node *ajson.Node) ([]*ajson.Node, error) {
	return jsonquery.New(node).ArrayOptional(dataKey)
}

func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "links").StrWithDefault("next", "")
}

func getDataMarshaller(nodeRecordFunc common.RecordTransformer, assc []Associations) common.MarshalFromNodeFunc { //nolint:lll
	return func(records []*ajson.Node, fields []string) ([]common.ReadResultRow, error) {
		data := make([]common.ReadResultRow, len(records))

		for index, nodeRecord := range records {
			raw, err := jsonquery.Convertor.ObjectToMap(nodeRecord)
			if err != nil {
				return nil, err
			}

			record, err := nodeRecordFunc(nodeRecord)
			if err != nil {
				return nil, err
			}

			data[index] = common.ReadResultRow{
				Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
				Raw:    raw,
			}

			// By default the ids are in float64 type
			id, ok := record["id"].(float64)
			if ok {
				data[index].Id = strconv.Itoa(int(id))
			}

			// better approach?
			for _, ass := range assc {
				if ass.ObjectId == strconv.Itoa(int(id)) {
					data[index].Associations = ass.AssociatedObjects
					// should i break?
				}
			}
		}

		return data, nil
	}
}
