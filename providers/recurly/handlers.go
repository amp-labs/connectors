package recurly

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var (
	ApiVersionHeader = "application/vnd.recurly.v2021-02-25+json" //nolint:gochecknoglobals
	limit            = "200"                                      //nolint:gochecknoglobals
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var url *urlbuilder.URL

	var err error

	if params.NextPage == "" {
		url, err = buildFirstPageURL(c.ProviderInfo().BaseURL, params)
	} else {
		url, err = buildNextPageURL(c.ProviderInfo().BaseURL, params.NextPage.String())
	}

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	// version is required in the Accept header
	// see: https://recurly.com/developers/api/v2021-02-25/index.html#section/Getting-Started/Versioning
	req.Header.Set("Accept", ApiVersionHeader)

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("data"),
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
		method = http.MethodPut

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

	// version is required in the Accept header
	// see: https://recurly.com/developers/api/v2021-02-25/index.html#section/Getting-Started/Versioning
	req.Header.Set("Accept", ApiVersionHeader)
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

	recordID, err := jsonquery.New(body).StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     resp,
	}, nil
}
