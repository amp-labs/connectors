package hubspot

import (
	"github.com/amp-labs/connectors/common"
)

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
