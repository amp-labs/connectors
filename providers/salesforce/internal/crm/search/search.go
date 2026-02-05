package search

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

func (s Strategy) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	// Additional parameter validation.
	if params.Limit != 0 {
		return nil, common.ErrPaginationControl
	}

	url, err := s.buildSearchURL(params)
	if err != nil {
		return nil, err
	}

	rsp, err := s.clientCRM.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		core.GetRecords,
		core.GetNextRecordsURL,
		core.GetDataMarshallerForSearch(params),
		params.Fields,
	)
}

func (s Strategy) buildSearchURL(params *common.SearchParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		return s.getModuleURL(params.NextPage.String())
	}

	// If NextPage is not set, then we're reading the first page of results.
	// We need to construct the SOQL query and then make the request.
	url, err := s.getQueryURL()
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("q", makeSOQL(params).String())

	return url, nil
}
