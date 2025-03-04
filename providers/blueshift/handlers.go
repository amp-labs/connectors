package blueshift

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/blueshift/metadata"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, path)
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
