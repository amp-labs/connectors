package api

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/document"
	"github.com/getkin/kin-openapi/openapi3"
)

type (
	ArrayLocator      = document.ArrayLocator
	OperationFilter   = document.OperationFilter
	PropertyFlattener = document.PropertyFlattener
)

// ArrayLocationAtObjectName item schema within response is stored under matching object name.
// Ex: requesting contacts will return payload with {"contacts":[...]}.
func ArrayLocationAtObjectName(objectName, fieldName string) bool {
	return fieldName == objectName
}

// ArrayLocationAtData item schema within response is always stored under the data field.
// Ex: requesting contacts or leads or users will return payload with {"data":[...]}.
func ArrayLocationAtData(objectName, fieldName string) bool {
	return fieldName == "data"
}

// ArrayLocationFromMap builds ArrayLocator using mapping,
// which knows exceptions and patterns to determine response field name.
//
// Ex:
//
//	ArrayLocationFromMap(datautils.NewDefaultMap(map[string]string{
//			"orders":	"orders",
//			"carts":	"carts",
//			"coupons":	"coupons",
//		}, func(key string) string { return "data" }))
//
// This can be understood as follows: orders, carts, coupons REST resources will be found under JSON response field
// matching "it's name", while the rest will be located under "data" field.
func ArrayLocationFromMap(dict datautils.DefaultMap[string, string]) ArrayLocator {
	return func(objectName, fieldName string) bool {
		expected := dict.Get(objectName)

		return fieldName == expected
	}
}

type listParams struct {
	mediaType           string
	propertyFlattener   PropertyFlattener
	autoSelectArrayItem *bool
	locator             ArrayLocator
	operationFilter     OperationFilter
}

type ListOption func(params *listParams)

func createParams(opts []ListOption) *listParams {
	var params listParams
	for _, opt := range opts {
		opt(&params)
	}

	// Default values are setup here.
	if params.locator == nil {
		params.locator = func(objectName, fieldName string) bool {
			return false
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
		params.autoSelectArrayItem = new(false)
	}

	if params.operationFilter == nil {
		params.operationFilter = func(objectName string, operation *openapi3.Operation) bool {
			return true
		}
	}

	return &params
}

type listOptions struct{}

var List = listOptions{} // nolint:gochecknoglobals

func (listOptions) WithArrayLocator(locator ArrayLocator) ListOption {
	return func(params *listParams) {
		params.locator = locator
	}
}

// WithMediaType picks which media type which should be used when searching schemas in API response.
// By default, schema is expected to be under "application/json" media response.
func (listOptions) WithMediaType(mediaType string) ListOption {
	return func(params *listParams) {
		params.mediaType = mediaType
	}
}

// WithPropertyFlattener allows nested fields to be moved to the top level.
// There are some APIs that hold fields of interest under grouping object, the nested object.
// This configuration flattens response schema fields.
// Please, have a look at PropertyFlattener documentation.
func (listOptions) WithPropertyFlattener(propertyFlattener PropertyFlattener) ListOption {
	return func(params *listParams) {
		params.propertyFlattener = propertyFlattener
	}
}

// WithAutoSelectArrayItems enables automatic selection of the array field in API responses
// if it is the only array type present.
// Default: Disabled.
//
// Use Case: This is helpful when APIs have inconsistent response field names, making it
// tedious to map each object name to its array field. If the response contains only one
// array property and each array represents the API resource schema, this option should be selected.
func (listOptions) WithAutoSelectArrayItems() ListOption {
	return func(params *listParams) {
		params.autoSelectArrayItem = new(true)
	}
}

func (o listOptions) WithOnlyOptionalQueryParameters() ListOption {
	return o.WithOperationFilter(func(objectName string, operation *openapi3.Operation) bool {
		for _, parameter := range operation.Parameters {
			if parameter.Value.In == "query" && parameter.Value.Required {
				// Operation should be ignored for metadata extraction.
				return false
			}
		}

		return true
	})
}

func (listOptions) WithOperationFilter(filter OperationFilter) ListOption {
	return func(params *listParams) {
		params.operationFilter = filter
	}
}
