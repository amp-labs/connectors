package api3

import (
	"github.com/amp-labs/connectors/common/naming"
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

type parameters struct {
	displayPostProcessing DisplayNameProcessor
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
