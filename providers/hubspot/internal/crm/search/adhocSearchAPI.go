package search

import (
	"context"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

// searchCRM is intended for objects outside HubSpot's ObjectAPI.
// For objects within ObjectAPI, refer to the Search method.
//
// Case-by-case explanation:
// * Lists
//   - Provider API endpoint for search:
//     https://developers.hubspot.com/docs/guides/api/crm/lists/overview#search-for-a-list
//   - "/search" always returns an array of items, unlike the usual "read" operation.
//     Therefore, the "retrieve" API endpoint is not used:
//     https://developers.hubspot.com/docs/guides/api/crm/lists/overview#retrieve-lists
func (s Strategy) searchViaNonstandardSearchAPI(
	ctx context.Context, params *common.SearchParams,
) (*common.ReadResult, error) {
	if len(params.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	if len(params.Fields) == 0 {
		return nil, common.ErrMissingFields
	}

	url, err := s.getSearchURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	payload, err := makeNonstandardSearchPayload(params)
	if err != nil {
		return nil, err
	}

	rsp, err := s.clientCRM.Post(ctx, url.String(), payload)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		common.ExtractOptionalRecordsFromPath(params.ObjectName),
		core.GetNextRecordsURLCRM,
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNonstandardSearchPayload(params *common.SearchParams) (nonstandardSearchPayload, error) {
	offset := 0

	if len(params.NextPage) != 0 {
		var err error

		offset, err = strconv.Atoi(params.NextPage.String())
		if err != nil {
			return nonstandardSearchPayload{}, fmt.Errorf("%w: %w", common.ErrNextPageInvalid, err)
		}
	}

	pageSize := core.DefaultPageSizeInt
	if params.Limit != 0 {
		pageSize = params.Limit
	}

	return nonstandardSearchPayload{
		Offset: offset,
		Count:  pageSize,
	}, nil
}

type nonstandardSearchPayload struct {
	Offset int   `json:"offset"`
	Count  int64 `json:"count"`
}
