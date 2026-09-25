package document

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
