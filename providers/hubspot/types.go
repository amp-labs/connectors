package hubspot

import "github.com/amp-labs/connectors/common"

type SearchParams struct {
	// The name of the object we are reading, e.g. "Account"
	ObjectName string // required
	// NextPage is an opaque token that can be used to get the next page of results.
	NextPage common.NextPageToken // optional, only set this if you want to read the next page of results
	// SortBy is the field to sort by in the direction specified by SortDirection.
	SortBy []SortBy // optional
	// FilterBy is the filter to apply to the search
	FilterGroups []FilterGroup // optional
	// Fields is the list of fields to return in the result.
	Fields []string // optional
}

type SortBy struct {
	// The name of the field to sort by.
	PropertyName string `json:"propertyName,omitempty"`
	// The direction to sort by.
	Direction SortDirection `json:"direction,omitempty"`
}

type FilterGroup struct {
	Filters []Filter `json:"filters,omitempty"`
}

type Filter struct {
	FieldName string             `json:"propertyName,omitempty"`
	Operator  FilterOperatorType `json:"operator,omitempty"`
	Value     string             `json:"value,omitempty"`
}

type (
	SortDirection      string
	FilterOperatorType string
)

const (
	SortDirectionAsc  SortDirection = "ASCENDING"
	SortDirectionDesc SortDirection = "DESCENDING"

	FilterOperatorTypeEQ           FilterOperatorType = "EQ"
	FilterOperatorTypeNEQ          FilterOperatorType = "NEQ"
	FilterOperatorTypeGT           FilterOperatorType = "GT"
	FilterOperatorTypeGTE          FilterOperatorType = "GTE"
	FilterOperatorTypeLT           FilterOperatorType = "LT"
	FilterOperatorTypeLTE          FilterOperatorType = "LTE"
	FilterOperatorBetween          FilterOperatorType = "BETWEEN"
	FilterOperatorIN               FilterOperatorType = "IN"
	FilterOperatorNIN              FilterOperatorType = "NIN"
	FilterPropertyHasProperty      FilterOperatorType = "HAS_PROPERTY"
	FilterPropertyNotHasProperty   FilterOperatorType = "NOT_HAS_PROPERTY"
	FilterPropertyContainsToken    FilterOperatorType = "CONTAINS_TOKEN"
	FilterPropertyNotContainsToken FilterOperatorType = "NOT_CONTAINS_TOKEN"
)

// ObjectField is used to define fields that exist on a hubspot object.
type ObjectField string

const (
	ObjectFieldHsObjectId         ObjectField = "hs_object_id"
	ObjectFieldHsLastModifiedDate ObjectField = "hs_lastmodifieddate"
	ObjectFieldLastModifiedDate   ObjectField = "lastmodifieddate"
)

type ObjectType string

const (
	ObjectTypeContact ObjectType = "contacts"
)
