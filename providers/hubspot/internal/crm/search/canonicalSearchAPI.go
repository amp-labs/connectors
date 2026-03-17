package search

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/associations"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

// https://developers.hubspot.com/docs/api/crm/search
func (s Strategy) searchViaObjectAPI(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := s.getObjectsAPISearchURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	rsp, err := s.clientCRM.Post(ctx, url.String(), makeFilterBody(params))
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		core.GetRecords,
		core.GetNextRecordsAfter,
		associations.CreateDataMarshallerWithAssociations(
			ctx, s.associationsFiller, params.ObjectName, params.AssociatedObjects,
		),
		params.Fields,
	)
}

func makeFilterBody(params *common.SearchParams) *CanonicalSearchPayload {
	filters := make([]Filter, 0, len(params.Filter.FieldFilters))

	// Put every filter under a single AND clause.
	for _, f := range params.Filter.FieldFilters {
		if f.Operator == common.FilterOperatorEQ {
			filters = append(filters, Filter{
				FieldName: f.FieldName,
				Operator:  FilterOperatorTypeEQ,
				Value:     f.Value,
			})
		}
	}

	payload := &CanonicalSearchPayload{
		FilterGroups: []FilterGroup{{
			// Hubspot API allows "OR" of multiple "AND" clauses.
			// But we include only one "AND" set.
			Filters: filters,
		}},
		Limit: searchPageSize,
		After: params.NextPage,
	}

	if params.Limit > 0 {
		payload.Limit = params.Limit
	}

	if params.Fields != nil {
		payload.Properties = params.Fields.List()
	}

	return payload
}

// CanonicalSearchPayload represents the body for a search request for ObjectAPI.
// See: https://developers.hubspot.com/docs/api-reference/search/guide#filter-search-results
//
// It contains:
//   - Limit: maximum number of results to return.
//   - FilterGroups: a list of filter groups, each representing an AND clause;
//     multiple groups are combined with an OR.
//   - After: paging token for fetching the next page of results.
//   - Properties: specific object fields to return.
type CanonicalSearchPayload struct {
	Limit        int64                `json:"limit,omitempty"`
	FilterGroups []FilterGroup        `json:"filterGroups,omitempty"`
	After        common.NextPageToken `json:"after,omitempty"`
	// Sorts        any                  `json:"sorts,omitempty"`
	Properties []string `json:"properties,omitempty"`
}

type FilterGroup struct {
	Filters []Filter `json:"filters,omitempty"`
}

type Filter struct {
	FieldName string             `json:"propertyName,omitempty"`
	Operator  FilterOperatorType `json:"operator,omitempty"`
	Value     any                `json:"value,omitempty"`
}

// FilterOperatorType defines operations allowed for search filtering.
// See full list here: https://developers.hubspot.com/docs/api-reference/search/guide#filter-search-results
type FilterOperatorType string

const FilterOperatorTypeEQ FilterOperatorType = "EQ"
