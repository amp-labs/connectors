package restlet

import (
	"context"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
)

// Search implements the SearchConnector interface by sending a search action
// to the RESTlet with field-based filters (instead of date-range filters used by Read).
func (a *Adapter) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	payload, err := buildSearchPayload(params)
	if err != nil {
		return nil, err
	}

	resp, err := a.JSONHTTPClient().Post(ctx, a.restletURL, payload)
	if err != nil {
		return nil, err
	}

	return parseSearchResults(resp)
}

func buildSearchPayload(params *common.SearchParams) (searchRequest, error) {
	pageIndex := 0

	if len(params.NextPage) != 0 {
		idx, err := strconv.Atoi(params.NextPage.String())
		if err != nil {
			return searchRequest{}, fmt.Errorf("invalid nextPage token: %w", err)
		}

		pageIndex = idx
	}

	columns := params.Fields.List()

	// Map FieldFilters to NetSuite search filters.
	// NetSuite requires explicit "AND" between multiple filter expressions.
	var filters []any

	for i, ff := range params.Filter.FieldFilters {
		if i > 0 {
			filters = append(filters, "AND")
		}

		nsOp, ok := filterOperatorMap[ff.Operator]
		if !ok {
			return searchRequest{}, fmt.Errorf("unsupported filter operator: %s", ff.Operator)
		}

		filters = append(filters, []any{ff.FieldName, nsOp, ff.Value})
	}

	pageSize := defaultPageSize
	if params.Limit > 0 {
		pageSize = int(params.Limit)
	}

	return searchRequest{
		Action:    "search",
		Type:      params.ObjectName,
		Columns:   columns,
		Filters:   filters,
		PageSize:  pageSize,
		PageIndex: pageIndex,
		Limit:     pageSize,
		Sort: []sortSpec{
			{Column: "internalid", Direction: "ASC"},
		},
	}, nil
}
