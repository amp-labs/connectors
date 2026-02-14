package salesfinity

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion      = "v1"
	defaultPageSize = "100" // https://docs.salesfinity.ai/api-reference/endpoint/call-log
)

var objectTimeField = datautils.NewDefaultMap( //nolint:gochecknoglobals
	datautils.Map[string, string]{
		"call-log": "updatedAt",
	},
	func(key string) string { return "" },
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	pageSize := readhelper.PageSizeWithDefaultStr(params, defaultPageSize)
	url.WithQueryParam("limit", pageSize)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(ctx context.Context, params common.ReadParams,
	request *http.Request, resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc("data"),
		makeFilterFunc(params, request),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}

func makeFilterFunc(params common.ReadParams, request *http.Request) common.RecordsFilterFunc {
	timeField := objectTimeField.Get(params.ObjectName)
	url, _ := urlbuilder.FromRawURL(request.URL)
	nextPageFunc := makeNextRecordsURL(url)

	if timeField == "" {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	return readhelper.MakeTimeFilterFunc(
		readhelper.ChronologicalOrder,
		readhelper.NewTimeBoundary(),
		timeField,
		time.RFC3339,
		nextPageFunc,
	)
}

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextPageNum, err := jsonquery.New(node, "pagination").TextWithDefault("next", "")
		if err != nil {
			return "", err
		}

		if nextPageNum == "" {
			return "", nil
		}

		reqLink.WithQueryParam("page", nextPageNum)

		return reqLink.String(), nil
	}
}
