package solarwinds

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const incidents = "incidents"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if !params.Since.IsZero() && params.ObjectName == incidents {
		url.WithQueryParam("updated_from", params.Since.Format(time.RFC3339))
	}

	if !params.Until.IsZero() && params.ObjectName == incidents {
		url.WithQueryParam("updated_to", params.Until.Format(time.RFC3339))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")

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
		common.ExtractOptionalRecordsFromPath(""),
		getNextRecordsURL(resp),
		common.GetMarshaledData,
		params.Fields,
	)
}
