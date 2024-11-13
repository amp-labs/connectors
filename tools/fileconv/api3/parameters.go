package api3

import (
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

// ObjectArrayLocator is a procedure that decides if field name is related to the object name.
// Below you can find the common cases.
type ObjectArrayLocator func(objectName, fieldName string) bool

// IdenticalObjectLocator item schema within response is stored under matching object name.
// Ex: requesting contacts will return payload with {"contacts":[...]}.
func IdenticalObjectLocator(objectName, fieldName string) bool {
	return fieldName == objectName
}

// DataObjectLocator item schema within response is always stored under the data field.
// Ex: requesting contacts or leads or users will return payload with {"data":[...]}.
func DataObjectLocator(objectName, fieldName string) bool {
	return fieldName == "data"
}

// CustomMappingObjectCheck builds ObjectArrayLocator using mapping,
// which knows exceptions and patterns to determine response field name.
//
// Ex:
//
//	CustomMappingObjectCheck(datautils.NewDefaultMap(map[string]string{
//			"orders":	"orders",
//			"carts":	"carts",
//			"coupons":	"coupons",
//		}, func(key string) string { return "data" }))
//
// This can be understood as follows: orders, carts, coupons REST resources will be found under JSON response field
// matching "it's name", while the rest will be located under "data" field.
func CustomMappingObjectCheck(dict datautils.DefaultMap[string, string]) ObjectArrayLocator {
	return func(objectName, fieldName string) bool {
		expected := dict.Get(objectName)

		return fieldName == expected
	}
}

// DisplayNameProcessor allows to format Display Names.
type DisplayNameProcessor func(displayName string) string

// CapitalizeFirstLetterEveryWord makes all words start with capital except some prepositions.
func CapitalizeFirstLetterEveryWord(displayName string) string {
	return naming.CapitalizeFirstLetterEveryWord(displayName)
}

// CamelCaseToSpaceSeparated converts camel case into lower case space separated string.
func CamelCaseToSpaceSeparated(displayName string) string {
	return strcase.ToDelimited(displayName, ' ')
}

// Pluralize will apply pluralization to the display name.
func Pluralize(displayName string) string {
	return naming.NewPluralString(displayName).String()
}

// ReadOperationMethodFilter callback that filters REST operations based on endpoint parameters.
type ReadOperationMethodFilter func(objectName string, operation *openapi3.Operation) bool

// OnlyOptionalQueryParameters operation must include only optional query parameters.
func OnlyOptionalQueryParameters(objectName string, operation *openapi3.Operation) bool {
	for _, parameter := range operation.Parameters {
		if parameter.Value.In == "query" && parameter.Value.Required {
			// Operation should be ignored for metadata extraction.
			return false
		}
	}

	return true
}

type parameters struct {
	displayPostProcessing DisplayNameProcessor
	operationMethodFilter ReadOperationMethodFilter
	mediaType             string
}

type Option = func(params *parameters)

func createParams(opts []Option) *parameters {
	var params parameters
	for _, opt := range opts {
		opt(&params)
	}

	// Default values are setup here.

	if params.displayPostProcessing == nil {
		params.displayPostProcessing = func(displayName string) string {
			return displayName
		}
	}

	if params.operationMethodFilter == nil {
		params.operationMethodFilter = func(objectName string, operation *openapi3.Operation) bool {
			return true
		}
	}

	if len(params.mediaType) == 0 {
		params.mediaType = "application/json"
	}

	return &params
}

// WithDisplayNamePostProcessors will apply processors in the given order.
func WithDisplayNamePostProcessors(processors ...DisplayNameProcessor) Option {
	return func(params *parameters) {
		params.displayPostProcessing = func(displayName string) string {
			for _, processor := range processors {
				displayName = processor(displayName)
			}

			return displayName
		}
	}
}

// WithParameterFilterGetMethod adds custom callback to decide
// if GET operation should be included based on parameters definitions.
func WithParameterFilterGetMethod(parameterFilter ReadOperationMethodFilter) Option {
	return func(params *parameters) {
		params.operationMethodFilter = parameterFilter
	}
}

// WithMediaType picks which media type which should be used when searching schemas in API response.
// By default, schema is expected to be under "application/json" media response.
func WithMediaType(mediaType string) Option {
	return func(params *parameters) {
		params.mediaType = mediaType
	}
}
