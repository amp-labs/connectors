package common

type ValueType string

const (
	ValueTypeString  = "string"
	ValueTypeBoolean = "boolean"
	ValueTypeFloat   = "float" // float is more preferred than int if provider doesn't differentiate.
	ValueTypeInt     = "int"

	ValueTypeDate     = "date"
	ValueTypeDateTime = "datetime"

	ValueTypeSingleSelect = "singleSelect"
	ValueTypeMultiSelect  = "multiSelect"

	ValueTypeOther = "other"
)
