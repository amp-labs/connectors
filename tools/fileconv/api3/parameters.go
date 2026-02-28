package api3

import (
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
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

// DefaultObjectLocator always returns false and logs the unresolved mapping.
// It should be replaced with a custom locator when ambiguity must be resolved.
func DefaultObjectLocator(objectName, fieldName string) bool {
	slog.Error("don't know which field holds an array for an object", "object", objectName, "fieldName", fieldName)

	return false
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

// SlashesToSpaceSeparated replaces URL slashes with spaces.
func SlashesToSpaceSeparated(displayName string) string {
	return strings.ReplaceAll(displayName, "/", " ")
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

// PropertyFlattener is used to inherit fields from nested object moving them to the top level.
// Ex:
//
//	{
//		"a":1,
//		"b":2,
//		"grouping": {
//			"c":3,
//			"d":4,
//		},
//		"e":5
//	}
//
// If we return true on "grouping" fieldName then it will be flattened with the resulting
// list of fields becoming "a", "b", "c", "d", "e".
type PropertyFlattener func(objectName, fieldName string) bool

// SingleItemDuplicatesResolver processes each endpoint individually.
func SingleItemDuplicatesResolver(mapping func(string) string) DuplicatesResolver {
	return func(collidingEndpoints [][]string) map[string]string {
		result := make(map[string]string)

		for _, endpoints := range collidingEndpoints {
			for _, endpoint := range endpoints {
				value := mapping(endpoint)
				// Object name can never start with a slash
				value, _ = strings.CutPrefix(value, "/")
				result[endpoint] = value
			}
		}

		return result
	}
}

// DuplicatesResolver processes groups of endpoints which collide among each other.
// Returns a registry mapping each endpoint URL path to a unique, non-colliding object name.
type DuplicatesResolver func(collidingEndpoints [][]string) map[string]string // endpoint to objectName

type parameters struct {
	displayPostProcessing DisplayNameProcessor
	operationMethodFilter ReadOperationMethodFilter
	propertyFlattener     PropertyFlattener
	mediaType             string
	autoSelectArrayItem   *bool
	duplicatesResolver    DuplicatesResolver
	versionPrefix         string
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

	if params.propertyFlattener == nil {
		params.propertyFlattener = func(objectName, fieldName string) bool {
			return false
		}
	}

	if len(params.mediaType) == 0 {
		params.mediaType = "application/json"
	}

	if params.autoSelectArrayItem == nil {
		// By default, auto selection is off.
		params.autoSelectArrayItem = goutils.Pointer(false)
	}

	if params.duplicatesResolver == nil {
		// Object name will be set to URL path.
		params.duplicatesResolver = SingleItemDuplicatesResolver(goutils.Identity)
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

// WithPropertyFlattening allows nested fields to be moved to the top level.
// There are some APIs that hold fields of interest under grouping object, the nested object.
// This configuration flattens response schema fields.
// Please, have a look at PropertyFlattener documentation.
func WithPropertyFlattening(propertyFlattener PropertyFlattener) Option {
	return func(params *parameters) {
		params.propertyFlattener = propertyFlattener
	}
}

// WithArrayItemAutoSelection enables automatic selection of the array field in API responses
// if it is the only array type present.
// Default: Disabled.
//
// Use Case: This is helpful when APIs have inconsistent response field names, making it
// tedious to map each object name to its array field. If the response contains only one
// array property and each array represents the API resource schema, this option should be selected.
func WithArrayItemAutoSelection() Option {
	return func(params *parameters) {
		params.autoSelectArrayItem = goutils.Pointer(true)
	}
}

// WithDuplicatesResolver enables custom object name creation based on endpoint paths.
// If the last part of the URI conflicts with other endpoints, they are treated as duplicates,
// and the resolver will be invoked to handle the collision.
// When this option is not specified, the default resolver uses the full URL to identify an object,
// ensuring no collisions.
func WithDuplicatesResolver(duplicatesResolver DuplicatesResolver) Option {
	return func(params *parameters) {
		params.duplicatesResolver = duplicatesResolver
	}
}

// WithVersionPrefix sets a version prefix to strip from API paths
// when generating object names.
//
// For example, if an API path is "/v1/customers/list", using
// WithVersionPrefix("/v1/") will result in "customers/list" as
// the object name.
func WithVersionPrefix(prefix string) Option {
	return func(params *parameters) {
		params.versionPrefix = prefix
	}
}
