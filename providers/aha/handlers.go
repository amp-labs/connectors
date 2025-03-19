package aha

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/aha/metadata"
	"github.com/spyzhov/ajson"
)

const (
	pageSizeKey = "per_page"
	pageSize    = "200"
	pageKey     = "page"
	sinceKey    = "created_since"
	apiVersion  = "v1"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	var (
		url *urlbuilder.URL
		err error
	)

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	addQueryParams(url, params, 1)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseKey := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	baseURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		response,
		common.GetRecordsUnderJSONPath(responseKey),
		makeNextRecordsURL(baseURL, params),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL(baseURL *urlbuilder.URL, params common.ReadParams) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pagination, err := jsonquery.New(node).ObjectRequired("pagination")
		if err != nil {
			return "", nil //nolint:nilerr
		}

		totalPage, err := jsonquery.New(pagination).IntegerRequired("total_pages")
		if err != nil {
			return "", err
		}

		currentPage, err := jsonquery.New(pagination).IntegerRequired("current_page")
		if err != nil {
			return "", err
		}

		if currentPage == totalPage {
			return "", nil
		}

		nextPage := currentPage + 1

		addQueryParams(baseURL, params, nextPage)

		return baseURL.String(), nil
	}
}

func addQueryParams(url *urlbuilder.URL, params common.ReadParams, page int64) {
	url.WithQueryParam(pageSizeKey, pageSize)
	url.WithQueryParam(pageKey, strconv.FormatInt(page, 10))

	if supportSince.Has(params.ObjectName) {
		url.WithQueryParam(sinceKey, datautils.Time.FormatRFC3339inUTC(params.Since))
	}
}
