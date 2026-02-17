package talkdesk

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) { //nolint: cyclop,lll
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	switch {
	case !params.Since.IsZero() && !params.Until.IsZero():
		if filtersByUpdates.Has(params.ObjectName) {
			url.WithQueryParam("updated_at_from", params.Since.UTC().Format(time.RFC3339))
			url.WithQueryParam("updated_at_to", params.Until.UTC().Format(time.RFC3339))
		}

		if filtersByCreation.Has(params.ObjectName) {
			url.WithQueryParam("created_at_from", params.Since.UTC().Format(time.RFC3339))
			url.WithQueryParam("created_at_to", params.Until.UTC().Format(time.RFC3339))
		}
	case !params.Since.IsZero():
		if filtersByUpdates.Has(params.ObjectName) {
			url.WithQueryParam("updated_at_from", params.Since.UTC().Format(time.RFC3339))
			url.WithQueryParam("updated_at_to", time.Now().UTC().Format(time.RFC3339))
		}

		if filtersByCreation.Has(params.ObjectName) {
			url.WithQueryParam("created_at_from", params.Since.UTC().Format(time.RFC3339))
			url.WithQueryParam("created_at_to", time.Now().UTC().Format(time.RFC3339))
		}
	case !params.Until.IsZero():
		if filtersByUpdates.Has(params.ObjectName) {
			url.WithQueryParam("updated_at_from", params.Since.UTC().Format(time.RFC3339))
			url.WithQueryParam("updated_at_to", time.Now().UTC().Format(time.RFC3339))
		}

		if filtersByCreation.Has(params.ObjectName) {
			url.WithQueryParam("created_at_from", params.Since.UTC().Format(time.RFC3339))
			url.WithQueryParam("created_at_to", time.Now().UTC().Format(time.RFC3339))
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		getRecords(params.ObjectName),
		nextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
}
