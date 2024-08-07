package api3

// ObjectCheck is a procedure that decides if field name is related to the object name.
// Below you can find the common cases.
type ObjectCheck func(fieldName, objectName string) bool

// IdenticalObjectCheck item schema within response is stored under matching object name.
// Ex: requesting contacts will return payload with {"contacts":[...]}.
func IdenticalObjectCheck(fieldName, objectName string) bool {
	return fieldName == objectName
}

// DataObjectCheck item schema within response is always stored under the data field.
// Ex: requesting contacts or leads or users will return payload with {"data":[...]}.
func DataObjectCheck(fieldName, objectName string) bool {
	return fieldName == "data"
}
