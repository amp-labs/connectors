package crm

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{}, //nolint:gochecknoglobals
	func(key string) string {
		return key
	},
)

type metadataFields struct {
	Data []fieldResults `json:"data"`
}

type records struct {
	Data []map[string]any `json:"data"`
}

type fieldResults struct {
	Code       string    `json:"field_code"`
	Name       string    `json:"field_name"`
	FieldType  string    `json:"field_type"` //nolint:tagliatelle
	IsCustom   bool      `json:"is_custom_field"`
	IsOptional bool      `json:"is_optional_response_field"`
	Options    []options `json:"options"`
}

// options represents the set of values one can use for enum, sets data Types.
// this oly works for objects: notes, activities, organizations, deals, products, persons.
type options struct {
	ID    any    `json:"id,omitempty"` // this can be an int,bool,string
	Label string `json:"label,omitempty"`
	Color string `json:"color,omitempty"`
	AltId string `json:"alt_id,omitempty"`
}
