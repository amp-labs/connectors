package api3

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

// ObjectCheck is a procedure that decides if field name is related to the object name.
// Below you can find the common cases.
type ObjectCheck func(objectName, fieldName string) bool

// IdenticalObjectCheck item schema within response is stored under matching object name.
// Ex: requesting contacts will return payload with {"contacts":[...]}.
func IdenticalObjectCheck(objectName, fieldName string) bool {
	return fieldName == objectName
}

// DataObjectCheck item schema within response is always stored under the data field.
// Ex: requesting contacts or leads or users will return payload with {"data":[...]}.
func DataObjectCheck(objectName, fieldName string) bool {
	return fieldName == "data"
}

// CustomMappingObjectCheck builds ObjectCheck using mapping,
// which knows exceptions and patterns to determine response field name.
//
// Ex:
//
//	CustomMappingObjectCheck(handy.NewDefaultMap(map[string]string{
//			"orders":	"orders",
//			"carts":	"carts",
//			"coupons":	"coupons",
//		}, func(key string) string { return "data" }))
//
// This can be understood as follows: orders, carts, coupons REST resources will be found under JSON response field
// matching "it's name", while the rest will be located under "data" field.
func CustomMappingObjectCheck(dict handy.DefaultMap[string, string]) ObjectCheck {
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

// ParameterFilterGetMethod callback that filters REST operations based on endpoint parameters.
type ParameterFilterGetMethod func(objectName string, operation *openapi3.Operation) bool

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
	parameterFilter       ParameterFilterGetMethod
}

type Option = func(params *parameters)

func createParams(opts []Option) *parameters {
	var params parameters
	for _, opt := range opts {
		opt(&params)
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
func WithParameterFilterGetMethod(parameterFilter ParameterFilterGetMethod) Option {
	return func(params *parameters) {
		params.parameterFilter = parameterFilter
	}
}
