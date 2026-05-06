package marketing

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/hubspot/internal/shared"
)

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := a.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

// When reading objects in Hubspot you must explicitly request the fields.
// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/guide#campaign-properties
//
// Reading campaigns object:
// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/get-campaigns
//   - Incremental reading is not available.
//   - Sorting is applied using "updatedAt" field from newest to oldest.
func (a *Adapter) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := a.getURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("properties", strings.Join(params.Fields.List(), ","))
	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, shared.DefaultPageSize))
	url.WithQueryParam("sort", "-updatedAt") // newest first

	return url, nil
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc("results"),
		makeIncrementalFilterFunc(params),
		readhelper.MakeMarshaledDataFuncWithId(
			common.FlattenNestedFields("properties"),
			readhelper.IdFieldQuery{Field: "id"},
		),
		params.Fields,
	)
}

// makeIncrementalFilterFunc embodies connector-side filtering.
// ReverseOrder is used because we request Campaigns sorted from newest to oldest.
func makeIncrementalFilterFunc(params common.ReadParams) common.RecordsFilterFunc {
	if params.Since.IsZero() && params.Until.IsZero() {
		return readhelper.MakeIdentityFilterFunc(shared.GetNextRecordsURL)
	}

	return readhelper.MakeTimeFilterFunc(
		readhelper.ReverseOrder,
		readhelper.NewTimeBoundary(),
		"updatedAt", time.RFC3339,
		shared.GetNextRecordsURL,
	)
}
