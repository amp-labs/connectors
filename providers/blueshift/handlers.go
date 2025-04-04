package blueshift

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/blueshift/metadata"
)

var writeVersion = "v1" //nolint:gochecknoglobals

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err = c.RootClient.URL(path)
	if err != nil {
		return nil, err
	}

	if supportPagination.Has(params.ObjectName) {
		url.WithQueryParam(pageSizeKey, pageSize)
		url.WithQueryParam(pageKey, pageNumber)
	}

	if params.NextPage != "" {
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	path, err := metadata.Schemas.LookupRawURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	baseURL, err := c.RootClient.URL("v" + path)
	if err != nil {
		return nil, err
	}

	if nestedObjects.Has(params.ObjectName) {
		return c.parseNestedResponse(response, params, baseURL.String())
	}

	return common.ParseResult(
		response,
		getRecords(params.ObjectName, c.Module()),
		makeNextRecordsURL(baseURL.String()),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		url    *urlbuilder.URL
		err    error
		method = http.MethodPost
	)

	if writeObjectWithSuffix.Has(params.ObjectName) {
		params.ObjectName = fmt.Sprintf("%s.json", params.ObjectName) //nolint:perfsprint
	}

	url, err = c.RootClient.URL(writeVersion, params.ObjectName)
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
	node, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: false,
		}, nil
	}

	rawID, err := jsonquery.New(node).IntegerOptional("id")
	if err != nil || rawID == nil {
		return &common.WriteResult{ //nolint:nilerr
			Success: true,
		}, nil
	}

	recordID := strconv.FormatInt(*rawID, 10)

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     nil,
	}, nil
}

func (c *Connector) parseNestedResponse(response *common.JSONHTTPResponse, params common.ReadParams, baseURL string) (*common.ReadResult, error) { //nolint:lll
	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	templatesNode, err := jsonquery.New(body).ObjectRequired("templates")
	if err != nil {
		return nil, err
	}

	jsonResponse, err := common.ParseJSONResponse(
		&http.Response{
			StatusCode: response.Code,
			Header:     response.Headers,
		},
		templatesNode.Source(),
	)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		jsonResponse,
		getRecords(params.ObjectName, c.Module()),
		makeNextRecordsURL(baseURL),
		common.GetMarshaledData,
		params.Fields,
	)
}
