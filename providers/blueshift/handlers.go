package blueshift

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/blueshift/metadata"
)

const writeVersion = "v1"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	path, err := metadata.Schemas.LookupObjectURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	fullPath := fmt.Sprintf("%s%s", "v", path)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, fullPath)
	if err != nil {
		return nil, err
	}

	log.Printf("URL: %s", url.String())

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
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	path, err := metadata.Schemas.LookupObjectURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	fullPath := fmt.Sprintf("%s%s", "v", path)

	baseURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, fullPath)
	if err != nil {
		return nil, err
	}

	if nestedObjects.Has(params.ObjectName) {
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
			getRecords(params.ObjectName),
			makeNextRecordsURL(baseURL.String()),
			common.GetMarshaledData,
			params.Fields,
		)
	}

	return common.ParseResult(
		response,
		getRecords(params.ObjectName),
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

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, writeVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	jsonData, errr := json.Marshal(params.RecordData)

	if errr != nil {
		return nil, errr
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	node, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: false,
		}, nil
	}

	rawID, err := jsonquery.New(node).IntegerOptional("id")
	if err != nil { //nolint:nilerr
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	if rawID == nil {
		return &common.WriteResult{
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
