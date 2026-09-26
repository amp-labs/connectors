package document

import "github.com/getkin/kin-openapi/openapi3"

// ArrayLocator is a procedure that decides if field name is related to the object name.
// Below you can find the common cases.
type ArrayLocator func(objectName, fieldName string) bool

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

// OperationFilter callback that filters REST operations based on endpoint parameters.
// Return true if the operation should be kept, false to omit.
type OperationFilter func(objectName string, operation *openapi3.Operation) bool
