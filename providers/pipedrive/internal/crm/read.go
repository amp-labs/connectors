package crm

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	readAPIVersion = "api/v2"
	data           = "data"
)

func (a *Adapter) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := a.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	resp, err := a.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath(data),
		nextRecordsURL(url),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (a *Adapter) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	url, err := a.getAPIURL(readAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if !params.Since.IsZero() {
		url.WithQueryParam("updated_since", params.Since.Format(time.RFC3339))
	}

	if !params.Until.IsZero() {
		url.WithQueryParam("updated_until", params.Since.Format(time.RFC3339))
	}

	return url, nil
}
