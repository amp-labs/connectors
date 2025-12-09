package suiteql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/netsuite/internal/shared"
)

const (
	// SuiteQL timestamp format for TO_TIMESTAMP function. This seems to be a magic
	// date format that is accepted by SuiteQL. We should verify it this actually
	// works for all instances.
	suiteQLTimestampFormat = "2006-01-02 15:04:05.000000000"

	maxRecordsPerPage = 1000
)

// buildReadRequest builds the HTTP request for SuiteQL queries.
// SuiteQL uses SQL-like queries instead of REST endpoints.
func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var urlStr string

	if params.NextPage != "" {
		urlStr = string(params.NextPage)
	} else {
		url, err := urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, "suiteql")
		if err != nil {
			return nil, err
		}

		url.WithQueryParam("limit", strconv.Itoa(maxRecordsPerPage))
		urlStr = url.String()
	}

	jsonBody, err := json.Marshal(makeSuiteQLBody(params))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Prefer", "transient")

	return req, nil
}

// parseReadResponse parses the response from a SuiteQL query.
func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(resp,
		common.ExtractRecordsFromPath("items"),
		shared.GetNextPageURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeSuiteQLBody(params common.ReadParams) suiteQLQueryBody {
	body := suiteQLQueryBody{
		Query: "SELECT * FROM " + params.ObjectName,
	}

	var queries []string

	if !params.Since.IsZero() {
		sinceStr := params.Since.Format(suiteQLTimestampFormat)
		queries = append(queries, fmt.Sprintf("lastModifiedDate >= TO_TIMESTAMP('%s', 'YYYY-MM-DD HH24:MI:SSxFF')", sinceStr))
	}

	if !params.Until.IsZero() {
		untilStr := params.Until.Format(suiteQLTimestampFormat)
		queries = append(queries, fmt.Sprintf("lastModifiedDate <= TO_TIMESTAMP('%s', 'YYYY-MM-DD HH24:MI:SSxFF')", untilStr))
	}

	if len(queries) > 0 {
		body.Query += " WHERE " + strings.Join(queries, " AND ")
	}

	return body
}

type suiteQLQueryBody struct {
	Query string `json:"q"`
}

// nolint:tagliatelle
type suiteQLResponse struct {
	Links        []suiteQLLink    `json:"links"`
	Count        int              `json:"count"`
	HasMore      bool             `json:"hasMore"`
	Items        []map[string]any `json:"items"`
	Offset       int              `json:"offset"`
	TotalResults int              `json:"totalResults"`
}

type suiteQLLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}
