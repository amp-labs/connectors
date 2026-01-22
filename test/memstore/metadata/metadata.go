package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/memstore"
	memstoretest "github.com/amp-labs/connectors/test/memstore"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Set up logging
	utils.SetupLogging()

	// Create context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Get configured connector
	conn := memstoretest.GetMemStoreConnector(ctx)

	slog.Info("=== MemStore Metadata Operations Examples ===")
	fmt.Println()

	// Run test functions
	testListObjectMetadata(ctx, conn)
	testInspectFieldMetadata(ctx, conn)
	testSchemaValidation(ctx, conn)

	slog.Info("All metadata operations completed successfully")
}

// testListObjectMetadata demonstrates getting metadata for all configured objects
func testListObjectMetadata(ctx context.Context, conn *memstore.Connector) {
	slog.Info("Listing metadata for all objects")

	objectNames := []string{"contacts", "companies", "deals"}

	result, err := conn.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		slog.Error("Failed to list object metadata", "error", err)
		return
	}

	slog.Info("Object metadata retrieved", "objectCount", len(result.Result))

	// Print summary for each object
	for objectName, metadata := range result.Result {
		slog.Info("Object metadata",
			"objectName", objectName,
			"displayName", metadata.DisplayName,
			"fieldCount", len(metadata.Fields))

		// Print field names
		fieldNames := make([]string, 0, len(metadata.Fields))
		for fieldName := range metadata.Fields {
			fieldNames = append(fieldNames, fieldName)
		}

		fmt.Printf("  Fields: %v\n", fieldNames)
	}

	fmt.Println()
}

// testInspectFieldMetadata demonstrates detailed inspection of field metadata
func testInspectFieldMetadata(ctx context.Context, conn *memstore.Connector) {
	slog.Info("Inspecting detailed field metadata")

	// Get metadata for contacts
	result, err := conn.ListObjectMetadata(ctx, []string{"contacts", "companies", "deals"})
	if err != nil {
		slog.Error("Failed to get metadata", "error", err)
		return
	}

	// Inspect contacts fields
	slog.Info("=== Contact Fields ===")

	contactMetadata := result.Result["contacts"]
	for fieldName, fieldInfo := range contactMetadata.Fields {
		required := false
		if fieldInfo.IsRequired != nil {
			required = *fieldInfo.IsRequired
		}

		slog.Info("Contact field",
			"fieldName", fieldName,
			"displayName", fieldInfo.DisplayName,
			"valueType", fieldInfo.ValueType,
			"providerType", fieldInfo.ProviderType,
			"required", required)

		// Show enum values if present
		if len(fieldInfo.Values) > 0 {
			enumValues := make([]string, len(fieldInfo.Values))
			for i, val := range fieldInfo.Values {
				enumValues[i] = val.Value
			}

			fmt.Printf("  Enum values: %v\n", enumValues)
		}
	}

	fmt.Println()

	// Inspect companies fields (focusing on enum and types)
	slog.Info("=== Company Fields ===")

	companyMetadata := result.Result["companies"]
	for fieldName, fieldInfo := range companyMetadata.Fields {
		required := false
		if fieldInfo.IsRequired != nil {
			required = *fieldInfo.IsRequired
		}

		slog.Info("Company field",
			"fieldName", fieldName,
			"valueType", fieldInfo.ValueType,
			"providerType", fieldInfo.ProviderType,
			"required", required)

		// Check for enum values (single select)
		if len(fieldInfo.Values) > 0 {
			enumValues := make([]string, len(fieldInfo.Values))
			for i, val := range fieldInfo.Values {
				enumValues[i] = val.Value
			}

			fmt.Printf("  Enum values: %v\n", enumValues)
		}
	}

	fmt.Println()

	// Inspect deals fields
	slog.Info("=== Deal Fields ===")

	dealMetadata := result.Result["deals"]
	for fieldName, fieldInfo := range dealMetadata.Fields {
		required := false
		if fieldInfo.IsRequired != nil {
			required = *fieldInfo.IsRequired
		}

		extraInfo := ""
		if required {
			extraInfo += " (REQUIRED)"
		}

		if len(fieldInfo.Values) > 0 {
			enumValues := make([]string, len(fieldInfo.Values))
			for i, val := range fieldInfo.Values {
				enumValues[i] = val.Value
			}

			extraInfo += fmt.Sprintf(" [enum: %v]", enumValues)
		}

		slog.Info("Deal field",
			"fieldName", fieldName,
			"valueType", fieldInfo.ValueType,
			"providerType", fieldInfo.ProviderType,
			"extra", extraInfo)
	}

	fmt.Println()
}

// testSchemaValidation demonstrates how metadata reflects schema constraints
func testSchemaValidation(ctx context.Context, conn *memstore.Connector) {
	slog.Info("Demonstrating schema validation through metadata")

	result, err := conn.ListObjectMetadata(ctx, []string{"contacts", "companies", "deals"})
	if err != nil {
		slog.Error("Failed to get metadata", "error", err)
		return
	}

	// Analyze contacts schema
	slog.Info("=== Contact Schema Constraints ===")

	contactMetadata := result.Result["contacts"]

	requiredFields := make([]string, 0)
	enumFields := make(map[string][]string)

	for fieldName, fieldInfo := range contactMetadata.Fields {
		if fieldInfo.IsRequired != nil && *fieldInfo.IsRequired {
			requiredFields = append(requiredFields, fieldName)
		}

		if len(fieldInfo.Values) > 0 {
			values := make([]string, len(fieldInfo.Values))
			for i, val := range fieldInfo.Values {
				values[i] = val.Value
			}

			enumFields[fieldName] = values
		}
	}

	slog.Info("Contact constraints",
		"requiredFieldCount", len(requiredFields))
	printJSON("Required Fields", requiredFields)
	printJSON("Enum Fields", enumFields)

	fmt.Println()

	// Analyze companies schema
	slog.Info("=== Company Schema Constraints ===")

	companyMetadata := result.Result["companies"]

	companyConstraints := make(map[string]map[string]any)

	for fieldName, fieldInfo := range companyMetadata.Fields {
		constraints := make(map[string]any)

		constraints["valueType"] = fieldInfo.ValueType
		if fieldInfo.IsRequired != nil && *fieldInfo.IsRequired {
			constraints["required"] = true
		}

		if len(fieldInfo.Values) > 0 {
			enumValues := make([]string, len(fieldInfo.Values))
			for i, val := range fieldInfo.Values {
				enumValues[i] = val.Value
			}

			constraints["enum"] = enumValues
		}

		if len(constraints) > 0 {
			companyConstraints[fieldName] = constraints
		}
	}

	printJSON("Company Field Constraints", companyConstraints)

	fmt.Println()

	// Analyze deals schema
	slog.Info("=== Deal Schema Constraints ===")

	dealMetadata := result.Result["deals"]

	dealConstraints := make(map[string]map[string]any)

	for fieldName, fieldInfo := range dealMetadata.Fields {
		constraints := make(map[string]any)

		constraints["valueType"] = fieldInfo.ValueType
		if fieldInfo.IsRequired != nil && *fieldInfo.IsRequired {
			constraints["required"] = true
		}

		if len(fieldInfo.Values) > 0 {
			enumValues := make([]string, len(fieldInfo.Values))
			for i, val := range fieldInfo.Values {
				enumValues[i] = val.Value
			}

			constraints["enum"] = enumValues
		}

		dealConstraints[fieldName] = constraints
	}

	printJSON("Deal Field Constraints", dealConstraints)

	fmt.Println()

	// Summary
	slog.Info("Schema validation summary",
		"contactRequiredFields", len(requiredFields),
		"contactEnumFields", len(enumFields),
		"companyConstrainedFields", len(companyConstraints),
		"dealFields", len(dealMetadata.Fields))
}

// printJSON prints data as formatted JSON
func printJSON(label string, data any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal JSON", "error", err)
		return
	}

	fmt.Printf("%s:\n%s\n", label, string(jsonData))
}
