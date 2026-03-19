package connectWise

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = "1000"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	url, err := c.getURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("pageSize", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

	conditions := make([]string, 0)
	if !params.Since.IsZero() {
		// Example:
		// 	LastUpdated = [2016-08-20T18:04:26Z]
		condition := fmt.Sprintf("LastUpdated >= [%v]", datautils.Time.FormatRFC3339inUTC(params.Since))
		conditions = append(conditions, condition)
	}

	if !params.Until.IsZero() {
		condition := fmt.Sprintf("LastUpdated <= [%v]", datautils.Time.FormatRFC3339inUTC(params.Until))
		conditions = append(conditions, condition)
	}

	conditionsQuery := strings.Join(conditions, " AND ")
	if conditionsQuery != "" {
		url.WithQueryParam("conditions", conditionsQuery)
	}

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(resp,
		// "" is used, because root level of JSON is right away an array.
		common.ExtractRecordsFromPath(""),
		nextRecordsURL(resp),
		common.GetMarshaledData,
		params.Fields,
	)
}

func nextRecordsURL(resp *common.JSONHTTPResponse) common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {
		return httpkit.HeaderLink(resp, "next"), nil
	}
}
