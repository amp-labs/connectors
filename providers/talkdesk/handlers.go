package talkdesk

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) { //nolint: cyclop,lll
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	switch {
	// The updates or creates filter requires passing both 'from' and 'to' values, and fails if only one is provided.
	// If both 'since' and 'until' are provided, we use them.
	// If we only have since, se set the until to now.
	// Similar when given only Until we set the since to 1970-01-01
	// There is no documentation for this, tests led to this.
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
			url.WithQueryParam("updated_at_from", time.Unix(0, 0).UTC().Format(time.RFC3339))
			url.WithQueryParam("updated_at_to", params.Until.UTC().Format(time.RFC3339))
		}

		if filtersByCreation.Has(params.ObjectName) {
			url.WithQueryParam("created_at_from", time.Unix(0, 0).UTC().Format(time.RFC3339))
			url.WithQueryParam("created_at_to", params.Until.UTC().Format(time.RFC3339))
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

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	// example ref: https://docs.talkdesk.com/reference/record-lists
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPatch

		if usesPUTForUpdates.Has(params.ObjectName) {
			method = http.MethodPut
		}
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	if method == http.MethodPatch {
		req.Header.Set("Content-Type", "application/json-patch+json")
	}

	return req, nil
}

func (c *Connector) parseWriteResponse(ctx context.Context, params common.WriteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(body).TextWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}
