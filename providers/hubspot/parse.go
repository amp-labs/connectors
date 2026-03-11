package hubspot

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

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
				return nil, core.ErrMissingId
			}

			result := common.ReadResultRow{
				Raw: record,
				Id:  id,
			}

			if len(fields) != 0 {
				recordProperties, ok := record["properties"].(map[string]any)
				if !ok {
					return nil, core.ErrNotObject
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
			err := c.crmAdapter.AssociationsFiller.FillAssociations(ctx, objName, &data, associatedObjects)
			if err != nil {
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
