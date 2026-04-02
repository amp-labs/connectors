package bentley

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if objectIncrementalSupport.Has(params.ObjectName) {
		var filters []string

		if !params.Since.IsZero() {
			filters = append(filters, "createdDateTime ge "+params.Since.Format("2006-01-02T15:04:05Z"))
		}

		if !params.Until.IsZero() {
			filters = append(filters, "createdDateTime le "+params.Until.Format("2006-01-02T15:04:05Z"))
		}

		if len(filters) > 0 {
			url.WithQueryParam("$filter", strings.Join(filters, " and "))
		}
	}

	if params.NextPage != "" {
		// Bentley returns a full URL in _links.next.href, so use it directly.
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		getRecords(c.ProviderContext.Module(), params.ObjectName),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		if objectUpdateWithPUT.Has(params.ObjectName) {
			method = http.MethodPut
		} else {
			method = http.MethodPatch
		}

		url.AddPath(params.RecordId)
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	// Some Bentley APIs wrap the response in a key (e.g. {"iTwin": {...}}),
	// others return the object directly. Unwrap if needed.
	node := body

	responseKey := writeResponseKey.Get(params.ObjectName)

	if responseKey != "" {
		nested, err := jsonquery.New(body).ObjectRequired(responseKey)
		if err != nil {
			return nil, err
		}

		node = nested
	}

	recordID, err := jsonquery.New(node).StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     resp,
	}, nil
}
