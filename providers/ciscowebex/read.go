package ciscowebex

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion      = "v1"
	defaultPageSize = 100
)

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

	limitParam := objectLimitQueryParam.Get(params.ObjectName)
	url.WithQueryParam(limitParam, strconv.Itoa(pageSize))

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

var objectTimeField = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	"groups":        "lastModified",
	"events":        "created",
	"organizations": "created",
}, func(key string) string {
	return ""
})

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseKey := objectResponseField.Get(params.ObjectName)

	return common.ParseResultFiltered(
		params,
		resp,
		makeOptionalRecordsFunc(responseKey),
		makeFilterFunc(params, resp),
		makeMarshaledDataWithIdFunc(),
		params.Fields,
	)
}

func makeOptionalRecordsFunc(jsonPath string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(jsonPath)
	}
}

func makeMarshaledDataWithIdFunc() common.MarshalFromNodeFunc {
	return func(records []*ajson.Node, fields []string) ([]common.ReadResultRow, error) {
		fields = append(fields, "id")

		result, err := common.MakeMarshaledDataFunc(nil)(records, fields)
		if err != nil {
			return nil, err
		}

		for i := range result {
			idAny := result[i].Fields["id"]
			if idAny == nil {
				continue
			}

			if id, ok := idAny.(string); ok {
				result[i].Id = id
			}
		}

		return result, nil
	}
}

func makeFilterFunc(params common.ReadParams, resp *common.JSONHTTPResponse) common.RecordsFilterFunc {
	timeField := objectTimeField.Get(params.ObjectName)

	nextPageFunc := makeNextRecordsURL(resp)
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

func makeNextRecordsURL(resp *common.JSONHTTPResponse) common.NextPageFunc {
	return func(_ *ajson.Node) (string, error) {
		return httpkit.HeaderLink(resp, "next"), nil
	}
}
