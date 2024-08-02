package outreach

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/spyzhov/ajson"
)

// getNextRecords returns the "next" url for the next page of results,
// If available, else returns an empty string.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	nextPageURL, err := jsonquery.New(node, "links").Str("next", true)
	if err != nil {
		return "", err
	}

	if nextPageURL == nil {
		return "", nil
	}

	return *nextPageURL, nil
}

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	records, err := jsonquery.New(node).Array("data", false)
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(records)
}

func getTotalSize(node *ajson.Node) (int64, error) {
	return jsonquery.New(node).ArraySize("data")
}

// getMarshalledData accepts a list of records and returns a list of structured data ([]ReadResultRow).
func getMarshalledData(records []map[string]interface{}, fields []string) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	for i, record := range records {
		data[i] = common.ReadResultRow{
			Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
		}
	}

	return data, nil
}

// parseData wraps the data in the format required by Outreach API.
func parseData(data any, objectName string, id ...string) (map[string]any, error) {
	var (
		nestedFields = make(map[string]any)
		attributes   = make(map[string]any)
		reqData      = make(map[string]any)
	)

	// Updating requires the id in the request body.
	// Re-adding it to the request.
	if len(id) > 0 {
		iD, err := strconv.Atoi(id[0])
		if err != nil {
			return nil, ErrIdMustInt
		}

		nestedFields[idKey] = iD
	}

	received, ok := data.(map[string]any) //nolint: varnamelen
	if !ok {
		return nil, ErrMustJSON
	}

	// If Relationships key has data, add it on the request.
	value, ok := received[relationshipsKey]
	if ok {
		nestedFields[relationshipsKey] = value
	}

	// Adds attributes key values.
	for k, v := range received {
		if k != relationshipsKey && k != typeKey {
			attributes[k] = v
		}
	}

	// If no type provided, provides a type which should be a singular word of the ObjectName
	// is added.
	_, ok = received[typeKey]
	if !ok {
		objectType := naming.NewSingularString(objectName)
		received[typeKey] = objectType
	}

	nestedFields[attributesKey] = attributes
	nestedFields[typeKey] = received[typeKey]
	reqData[dataKey] = nestedFields

	return reqData, nil
}
