package main

import (
	"encoding/json"
	"fmt"

	"github.com/amp-labs/connectors/providers/justcall/metadata"
)

func main() {
	result := make(map[string]map[string]interface{})

	for _, mod := range metadata.Schemas.Modules {
		for name, obj := range mod.Objects {
			fields := make(map[string]map[string]interface{})
			for fieldName, field := range obj.Fields {
				fields[fieldName] = map[string]interface{}{
					"DisplayName":  field.DisplayName,
					"IsCustom":     nil,
					"IsRequired":   nil,
					"ProviderType": field.ProviderType,
					"ReadOnly":     nil,
					"ValueType":    field.ValueType,
					"Values":       []string{},
				}
			}
			result[name] = map[string]interface{}{
				"DisplayName": obj.DisplayName,
				"Fields":      fields,
			}
		}
	}

	output := map[string]interface{}{
		"Errors": map[string]string{},
		"Result": result,
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(jsonBytes))
}
