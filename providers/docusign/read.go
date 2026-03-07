package docusign

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/docusign/metadata"
	"github.com/spyzhov/ajson"
)

var (
	defaultTimeRange = time.Now().AddDate(-2, 0, 0) // 2 years
	maxPageSize      = 1000
	maxUsersPageSize = 100

	nextURIKey = "nextUri"
)

var (
	incrementalObjects = datautils.NewSet(
		"envelopes",
		"bulk_send_batch",
		"templates",
		"users",
	)

	requiredQueryParamsObjects = datautils.NewSet(
		"envelopes",
		// Requires either from_date or batch_ids but doesn't return an error if neither is provided.
		// https://developers.docusign.com/docs/esign-rest-api/reference/bulkenvelopes/bulksend/getbulksendbatches/
		"bulk_send_batch",
	)

	responseKeyOverrides = map[string]string{
		"templates":       "envelopeTemplates",
		"tab_definitions": "tabs",
		"bulk_send_batch": "bulkBatchSummaries",
		"bulk_send_lists": "bulkListSummaries",
		"signing_groups":  "groups",
	}
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	req, err := c.buildReadRequest(ctx, config)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(ctx, req.URL.String())
	if err != nil {
		return nil, err
	}

	return c.parseReadResponse(ctx, config, req, resp)
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	reqURL, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if config.NextPage != "" {
		return urlbuilder.New(config.NextPage.String())
	}

	path, err := metadata.Schemas.FindURLPath(common.ModuleRoot, config.ObjectName)
	if err != nil {
		return nil, err
	}
	path = strings.ReplaceAll(path, "{accountId}", c.accountId)

	url, err := urlbuilder.New(c.BaseURL, restapiPrefix, path)
	if err != nil {
		return nil, err
	}
	addQueryParams(url, config)

	return url, nil
}

func (c *Connector) parseReadResponse(_ context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		makeRecords(params),
		getNextRecordURL(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeRecords(params common.ReadParams) common.RecordsFunc {
	objName := params.ObjectName
	if respKey, ok := responseKeyOverrides[objName]; ok {
		objName = respKey
	}
	return common.ExtractRecordsFromPath(objName)
}

func getNextRecordURL(req *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextUri, err := jsonquery.New(node).StrWithDefault(nextURIKey, "")
		if err != nil || nextUri == "" {
			return "", err
		}

		// /restapi/v2.1 is stripped from nextUri but is needed to construct the full URL.
		// So replace the query params of the original request with the ones in nextUri.
		req.RawQuery = ""
		nextURL, err := urlbuilder.FromRawURL(req)
		if err != nil {
			return "", err
		}

		// Extract the query params
		parsedNextUri, err := url.Parse(nextUri)
		if err != nil {
			return "", err
		}
		for key, param := range parsedNextUri.Query() {
			nextURL.WithQueryParamList(key, param)
		}
		return nextURL.String(), nil
	}
}

func addQueryParams(url *urlbuilder.URL, config common.ReadParams) {
	if incrementalObjects.Has(config.ObjectName) {
		startTime := config.Since
		endTime := config.Until
		count := config.PageSize

		if requiredQueryParamsObjects.Has(config.ObjectName) && startTime.IsZero() {
			startTime = defaultTimeRange
		}

		if !startTime.IsZero() {
			url.WithQueryParam("from_date", startTime.UTC().Format(time.RFC3339))
		}
		if !endTime.IsZero() {
			url.WithQueryParam("to_date", config.Until.Format(time.RFC3339))
		}

		if count <= 0 {
			if config.ObjectName == "users" {
				count = maxUsersPageSize
			} else {
				count = maxPageSize
			}
		}
		url.WithQueryParam("count", strconv.Itoa(count))
	}
}
