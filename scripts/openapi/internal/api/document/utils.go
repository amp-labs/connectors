package document

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

func getQueryParameters(operation *openapi3.Operation) []string {
	queryParams := make([]string, 0, len(operation.Parameters))

	for _, parameter := range operation.Parameters {
		if parameter.Value.In == "query" {
			queryParams = append(queryParams, parameter.Value.Name)
		}
	}

	return queryParams
}

func extractSchema(operation *openapi3.Operation, mime string) *openapi3.Schema {
	if operation == nil {
		return nil
	}

	responses := operation.Responses
	if responses == nil {
		return nil
	}

	// any 2xx response will suffice
	status := responses.Status(http.StatusOK)
	if status == nil {
		return nil
	}

	value := status.Value
	if value == nil {
		return nil
	}

	mediaType := value.Content.Get(mime)
	if mediaType == nil {
		return nil
	}

	schema := mediaType.Schema
	if schema == nil {
		return nil
	}

	schemaValue := schema.Value
	if schemaValue == nil {
		return nil
	}

	return schemaValue
}

type definitionSchemaType int

const (
	schemaTypeUnknown definitionSchemaType = iota
	schemaTypeObject
	schemaTypeArray
)

// This logs any concerns if any.
// The OpenAPI extractor has some expectations that should hold true, otherwise the extraction
// should be rethought to match the edge case.
//
// Operation response has a schema, it must be an object, and it should contain a field
// that will hold Object of interest. That object is what connectors.ReadConnector returns via Read method.
// It is allowed to have an array without it being nested under object.
//
// Returns enum marking the type of schema. This can be used to adjust processing.
func getSchemaType(schema *openapi3.Schema) definitionSchemaType {
	if schema.Type == nil {
		return schemaTypeUnknown
	}

	if len(*schema.Type) != 1 {
		slog.Info("Schema definition has multiple types")
	}

	if slices.Contains(*schema.Type, "object") {
		return schemaTypeObject
	}

	if slices.Contains(*schema.Type, "array") {
		return schemaTypeArray
	}

	slog.Info("Schema definition is neither an object nor an array. " +
		"Expected to be an object containing array of items, or the list itself.")

	return schemaTypeUnknown
}

func createUnprocessableObjectError(endpointPath string) error {
	return fmt.Errorf("%w: object path %v", ErrUnprocessableObject, endpointPath)
}

func isBooleanTruthful(input *bool) bool {
	if input == nil {
		return false
	}

	return *input
}
