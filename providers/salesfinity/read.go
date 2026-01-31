package salesfinity

import (
	"context"
	"net/http"
	"strconv"
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
	defaultPageSize = 10 //https://docs.salesfinity.ai/api-reference/endpoint/call-log#:~:text=items%20per%20page%20(-,default%20is%2010%2C,-max%20is%20100
)

var objectTimeField = datautils.NewDefaultMap(datautils.Map[string, string]{}, func(key string) string {

	switch key {
	case "call-log":
		return "updatedAt"
	default:
		return ""
	}

})

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}
	pageSize := defaultPageSize
	if params.PageSize > 0 {
		pageSize = params.PageSize
	}
	url.WithQueryParam("limit", strconv.Itoa(pageSize))
	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(ctx context.Context, params common.ReadParams, request *http.Request, resp *common.JSONHTTPResponse) (*common.ReadResult, error) {
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
		nextPageNum, err := jsonquery.New(node, "pagination").IntegerOptional("next")
		if err != nil {
			return "", err
		}
		if nextPageNum == nil {
			return "", nil
		}

		reqLink.WithQueryParam("page", strconv.FormatInt(*nextPageNum, 10))
		return reqLink.String(), nil
	}
}
