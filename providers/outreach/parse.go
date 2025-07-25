package outreach

import (
	"errors"
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
	assocMap := make(map[string]map[string][]common.Association)
	for _, a := range assc {
		assocMap[a.ObjectId] = a.AssociatedObjects
	}

	return func(records []*ajson.Node, fields []string) ([]common.ReadResultRow, error) {
		data := make([]common.ReadResultRow, len(records))

		for idx, nodeRecord := range records {
			raw, err := jsonquery.Convertor.ObjectToMap(nodeRecord)
			if err != nil {
				return nil, err
			}

			record, err := nodeRecordFunc(nodeRecord)
			if err != nil {
				return nil, err
			}

			// RecordId of Outreach Objects are float64 by default.
			idF, ok := record["id"].(float64)
			if !ok {
				return nil, errors.New("failed to convert the object id to expected type") //nolint: err113
			}

			idStr := strconv.Itoa(int(idF))

			data[idx] = common.ReadResultRow{
				Id:           idStr,
				Fields:       common.ExtractLowercaseFieldsFromRaw(fields, record),
				Raw:          raw,
				Associations: assocMap[idStr],
			}
		}

		return data, nil
	}
}
