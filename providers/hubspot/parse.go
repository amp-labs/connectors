package hubspot

import (
	"context"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

/*
Pagination format:

{
  "results": [...],
  "paging": {
    "next": {
      "after": "394",
      "link": "https://api.hubapi.com/crm/v3/objects/contacts?limit=100&properties=listId%2Cname&after=394"
    }
  }
}
*/

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
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	records, err := node.GetKey("results")
	if err != nil {
		return nil, err
	}

	if !records.IsArray() {
		return nil, ErrNotArray
	}

	arr := records.MustArray()

	out := make([]map[string]any, 0, len(arr))

	for _, v := range arr {
		if !v.IsObject() {
			return nil, ErrNotObject
		}

		data, err := v.Unpack()
		if err != nil {
			return nil, err
		}

		m, ok := data.(map[string]any)
		if !ok {
			return nil, ErrNotObject
		}

		out = append(out, m)
	}

	return out, nil
}

// getDataMarshaller returns a function that accepts a list of records and fields
// and returns a list of structured data ([]ReadResultRow).
//
//nolint:gocognit
func (c *Connector) getDataMarshaller(
	ctx context.Context,
	objName string,
	associatedObjects []string,
) func(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
	return func(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
		data := make([]common.ReadResultRow, len(records))

		//nolint:varnamelen
		for i, record := range records {
			id, ok := record["id"].(string)
			if !ok {
				return nil, errMissingId
			}

			result := common.ReadResultRow{
				Raw: record,
				Id:  id,
			}

			if len(fields) != 0 {
				recordProperties, ok := record["properties"].(map[string]any)
				if !ok {
					return nil, ErrNotObject
				}

				result.Fields = common.ExtractLowercaseFieldsFromRaw(fields, recordProperties)

				// Some fields like "id" exist at the top level of the record,
				// not inside the "properties" object. Add those if requested.
				for _, field := range fields {
					lowercaseField := strings.ToLower(field)
					if _, exists := result.Fields[lowercaseField]; !exists {
						if value, ok := record[lowercaseField]; ok {
							result.Fields[lowercaseField] = value
						}
					}
				}
			}

			data[i] = result
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

func getNextRecordsURLCRM(node *ajson.Node) (string, error) {
	hasMore, err := jsonquery.New(node).BoolWithDefault("hasMore", false)
	if err != nil {
		return "", err
	}

	if !hasMore {
		// Next page doesn't exist
		return "", nil
	}

	offset, err := jsonquery.New(node).IntegerWithDefault("offset", 0)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(offset, 10), nil
}
