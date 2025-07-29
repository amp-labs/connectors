package suiteql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	// SuiteQL timestamp format for TO_TIMESTAMP function.
	suiteQLTimestampFormat = "2006-01-02 15:04:05.000000000"
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

		url.WithQueryParam("limit", "5")
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
		makeNextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// The response is a JSON object with a "links" property.
		// The "links" property is an array of objects with a "rel" property and a "href" property.
		// We need to find the "next" link and return the "href" property.
		links, err := jsonquery.New(node).ArrayRequired("links")
		if err != nil {
			return "", err
		}

		for _, link := range links {
			rel, err := jsonquery.New(link).StringOptional("rel")
			if err != nil {
				return "", err
			}

			if rel != nil && *rel == "next" {
				href, err := jsonquery.New(link).StringRequired("href")
				if err != nil {
					return "", err
				}

				return href, nil
			}
		}

		return "", nil
	}
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
