package docusign

import (
	"context"
	"fmt"
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
	defaultPageSize  = 1000

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
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	reqURL, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, reqURL.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		res,
		common.ExtractRecordsFromPath(config.ObjectName),
		getNextRecordURL(c.BaseURL),
		common.GetMarshaledData,
		config.Fields,
	)
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

func makeGetRecords(objectName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		// templates needs workaround?
		responseFieldName := metadata.Schemas.LookupArrayFieldName(common.ModuleRoot, objectName)

		return jsonquery.New(node).ArrayOptional(responseFieldName)
	}
}

// cleanup reconstructing the url...not sure I like the manual reconstruction.
// try to implement new pattern buildReadRequest which returns a req and wrap it in
// the old read.
// todo: set up tests for pagination with envelopes -> other objects -> refactor/cleanup
func getNextRecordURL(baseURL string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextUri, err := jsonquery.New(node).StrWithDefault(nextURIKey, "")
		if err != nil || nextUri == "" {
			return "", err
		}

		// Preserve the query parameters from nextUri

		// /restapi/v2.1 is stripped from the nextUri value so we need to add them back for the full path.
		fullURL := fmt.Sprintf("%s/%s/%s%s", baseURL, restapiPrefix, versionPrefix, nextUri)
		fullURLParsed, err := url.Parse(fullURL)
		if err != nil {
			return "", err
		}

		nextURL, err := urlbuilder.FromRawURL(fullURLParsed)
		if err != nil {
			return "", err
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

		if count < 0 {
			count = defaultPageSize
		}
		url.WithQueryParam("count", strconv.Itoa(count))
	}
}
