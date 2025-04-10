package heyreach

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "public", params.ObjectName)
	if err != nil {
		return nil, err
	}

	body, err := constructRequestBody(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(body))
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	var (
		offset int
		err    error
	)

	if params.NextPage.String() != "" {
		offset, err = strconv.Atoi(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("items"),
		makeNextRecord(offset),
		common.GetMarshaledData,
		params.Fields,
	)
}

// constructRequestBody builds the request payload for reading data using a POST method.
// Unlike traditional GET requests where pagination is usually passed as query parameters,
// this approach uses a POST request, so pagination values (like "offset", "limit") are included
// directly in the JSON body.
func constructRequestBody(config common.ReadParams) ([]byte, error) {
	body := map[string]any{}

	if len(config.NextPage) != 0 {
		offset, err := strconv.Atoi(config.NextPage.String())
		if err != nil {
			return nil, err
		}

		body["offset"] = offset
	}

	body["limit"] = DefaultPageSize

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		url    *urlbuilder.URL
		err    error
		method = http.MethodPost
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "public", params.ObjectName)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(body).IntegerWithDefault("id", 0)
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: strconv.Itoa(int(recordID)),
		Errors:   nil,
		Data:     resp,
	}, nil
}
