package hubspot

import (
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

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
	Fields datautils.Set[string] // optional
	// AssociatedObjects is a list of associated objects to fetch along with the main object.
	AssociatedObjects []string // optional
}

func (p SearchParams) ValidateParams() error {
	if len(p.ObjectName) == 0 {
		return common.ErrMissingObjects
	}

	if len(p.Fields) == 0 {
		return common.ErrMissingFields
	}

	return nil
}

type searchCRMParams struct {
	SearchParams

	PageSize int64
}

func (p searchCRMParams) payload() (searchCRMPayload, error) {
	offset := 0

	if len(p.NextPage) != 0 {
		var err error

		offset, err = strconv.Atoi(p.NextPage.String())
		if err != nil {
			return searchCRMPayload{}, fmt.Errorf("%w: %w", common.ErrNextPageInvalid, err)
		}
	}

	pageSize := core.DefaultPageSizeInt
	if p.PageSize != 0 {
		pageSize = p.PageSize
	}

	return searchCRMPayload{
		Offset: offset,
		Count:  pageSize,
	}, nil
}

type searchCRMPayload struct {
	Offset int   `json:"offset"`
	Count  int64 `json:"count"`
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

type Filters []Filter

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
	ObjectFieldId                 ObjectField = "id"
	ObjectFieldProperties         ObjectField = "properties"
)

type ObjectType string

const (
	ObjectTypeContact ObjectType = "contacts"
)

type hubspotHeaderKey string

const (
	xHubspotRequestTimestamp hubspotHeaderKey = "X-Hubspot-Request-Timestamp"
	xHubspotSignatureV3      hubspotHeaderKey = "X-Hubspot-Signature-V3"
)
