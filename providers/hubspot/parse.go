package hubspot

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// getNextRecordsAfter returns the "after" value for the next page of results.
func getNextRecordsAfter(node *ajson.Node) (string, error) {
	var nextPage string

	if node.HasKey("paging") {
		next, err := parsePagingNext(node)
		if err != nil {
			return "", err
		}

		after, err := next.GetKey("after")
		if err != nil {
			return "", err
		}

		if !after.IsString() {
			return "", ErrNotString
		}

		nextPage = after.MustString()
	}

	return nextPage, nil
}

// getNextRecordsURL returns the URL for the next page of results.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	var nextPage string

	if node.HasKey("paging") {
		next, err := parsePagingNext(node)
		if err != nil {
			return "", err
		}

		link, err := next.GetKey("link")
		if err != nil {
			return "", err
		}

		if !link.IsString() {
			return "", ErrNotString
		}

		nextPage = link.MustString()
	}

	return nextPage, nil
}

// parsePagingNext is a helper to return the paging.next node.
func parsePagingNext(node *ajson.Node) (*ajson.Node, error) {
	paging, err := node.GetKey("paging")
	if err != nil {
		return nil, err
	}

	if !paging.IsObject() {
		return nil, ErrNotObject
	}

	next, err := paging.GetKey("next")
	if err != nil {
		return nil, err
	}

	if !next.IsObject() {
		return nil, ErrNotObject
	}

	return next, nil
}

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]interface{}, error) {
	records, err := node.GetKey("results")
	if err != nil {
		return nil, err
	}

	if !records.IsArray() {
		return nil, ErrNotArray
	}

	arr := records.MustArray()

	out := make([]map[string]interface{}, 0, len(arr))

	for _, v := range arr {
		if !v.IsObject() {
			return nil, ErrNotObject
		}

		data, err := v.Unpack()
		if err != nil {
			return nil, err
		}

		m, ok := data.(map[string]interface{})
		if !ok {
			return nil, ErrNotObject
		}

		out = append(out, m)
	}

	return out, nil
}

// getMarshalledData accepts a list of records and returns a list of structured data ([]ReadResultRow).
func (c *Connector) getMarshalledData(ctx context.Context, objName string, associatedObjects []string) func(records []map[string]interface{}, fields []string) ([]common.ReadResultRow, error) {
	return func(records []map[string]interface{}, fields []string) ([]common.ReadResultRow, error) {
		data := make([]common.ReadResultRow, len(records))

		//nolint:varnamelen
		for i, record := range records {
			recordProperties, ok := record["properties"].(map[string]interface{})
			if !ok {
				return nil, ErrNotObject
			}

			id, ok := record["id"].(string)
			if !ok {
				return nil, errMissingId
			}

			data[i] = common.ReadResultRow{
				Fields: common.ExtractLowercaseFieldsFromRaw(fields, recordProperties),
				Raw:    record,
				Id:     id,
			}
		}

		if len(associatedObjects) > 0 {
			if err := c.fillAssociations(ctx, objName, &data, associatedObjects); err != nil {
				return nil, err
			}
		}

		return data, nil
	}
}

// GetResultId returns the id of a hubspot result row.
// nolint:cyclop
func GetResultId(row *common.ReadResultRow) string {
	if row == nil {
		return ""
	}

	// Attempt to get it from the fields
	if idValue, ok := row.Fields[string(ObjectFieldId)].(string); ok && idValue != "" {
		return idValue
	} else if idValue, ok = row.Fields[string(ObjectFieldHsObjectId)].(string); ok && idValue != "" {
		return idValue
	}

	// Attempt to get it from raw
	if idValue, ok := row.Raw[string(ObjectFieldId)].(string); ok && idValue != "" {
		return idValue
	}

	// Attempt to get the properties map
	propertiesValue, ok := row.Raw[string(ObjectFieldProperties)].(map[string]any)
	if !ok || propertiesValue == nil {
		return ""
	}

	// Attempt to get the ObjectFieldHsObjectId from the properties map
	if hsObjectId, ok := propertiesValue[string(ObjectFieldHsObjectId)].(string); ok && hsObjectId != "" {
		return hsObjectId
	}

	// If everything fails, return an empty string
	return ""
}
