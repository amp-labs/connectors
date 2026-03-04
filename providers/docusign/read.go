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
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/docusign/metadata"
	"github.com/spyzhov/ajson"
)

var (
	defaultTimeRange = time.Now().AddDate(-2, 0, 0) // 2 years
	defaultPageSize  = 1000

	nextURIKey = "nextUri"
)

// make this work for envelopes first before tackling the other potential objects
// there are established patterns.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	// Check read allowed on object?

	// Apply filters & find page limit

	// Pagination

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	// check against non-2xx
	return common.ParseResult(
		res,
		common.ExtractRecordsFromPath(config.ObjectName),
		getNextRecordURL,
		common.GetMarshaledData, // define marshall function to map the response to ReadRowResult
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if config.NextPage != "" {
		// /restapi/v2.1 is stripped from the nextUri value so we need to add them back for the full path.
		// return urlbuilder.New(c.BaseURL, restapiPrefix, versionPrefix, config.NextPage.String())
		tURL := fmt.Sprintf("%s/%s/%s/%s", c.BaseURL, restapiPrefix, versionPrefix, config.NextPage.String())
		parsed, err := url.Parse(tURL)
		if err != nil {
			return nil, err
		}

		return urlbuilder.FromRawURL(parsed)
		// if err != nil {
		// 	return nil, err
		// }

		// fmt.Printf("%s\n", fPath.String())

		// path, err := urlbuilder.New(c.BaseURL, restapiPrefix, versionPrefix)
		// if err != nil {
		// 	return nil, err
		// }

		// qParams, err := url.ParseQuery(parsed.RawQuery)
		// if err != nil {
		// 	return nil, err
		// }
		// for key, q := range qParams {
		// 	path.WithQueryParamList(key, q)
		// }

		// return path, nil
	}

	path, err := metadata.Schemas.FindURLPath(common.ModuleRoot, config.ObjectName)
	if err != nil {
		return nil, err
	}
	// Can we strip this from schemas.json paths and just build the full accounts+objects path at request time?
	path = strings.ReplaceAll(path, "{accountId}", c.accountId)

	// Add query params
	url, err := urlbuilder.New(c.BaseURL, restapiPrefix, path)
	if err != nil {
		return nil, err
	}

	resolveEnvelopesQueryParams(url, config)

	return url, nil
}

func makeGetRecords(objectName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		// templates needs workaround?
		responseFieldName := metadata.Schemas.LookupArrayFieldName(common.ModuleRoot, objectName)

		return jsonquery.New(node).ArrayOptional(responseFieldName)
	}
}

// todo: check next page / pagination
func getNextRecordURL(node *ajson.Node) (string, error) {
	nextUri, err := jsonquery.New(node).StringOptional(nextURIKey)
	if err != nil {
		return "", err
	}
	return *nextUri, nil
}

// For envelopes, `from_date` is required unless `envelope_ids`, `folder_ids`, or `transaction_ids` is provided.
// In which case, `from_date` will default to the last 2 years (set by their server?).
// From perspective of just proxying, reject request if it doesn't meet query params of Docusign's API,
// let it error (seems bad -> eating up quota), or set the default 2 years (call it out in docs, let users
// specify)?
// Mainly for the timeframe stuff + some validation
// page size also defaults to 1000, maybe don't need explicit handling as that's what next_uri is for
// ISO 8601 format recommended for time (envelopes)
//
// Filter supports DocuSign envelope query parameters as a &-separated key=value string.
// Supported keys: envelope_ids, folder_ids, folder_types, status, from_to_status, search_text
// Example: "envelope_ids=abc,def&status=sent&folder_ids=drafts"
// Reference: https://developers.docusign.com/docs/esign-rest-api/reference/envelopes/envelopes/list/
func resolveEnvelopesQueryParams(url *urlbuilder.URL, config common.ReadParams) {
	// If no `envelope_ids`, `folder_ids`, or `transaction_ids`, set `from_date` to default 2 years (with some wiggle room?)
	// Otherwise, it doesn't need to be set but if it is, it needs to be appended
	// if config.Filter != "" {
	// 	for _, pair := range strings.Split(config.Filter, "&") {
	// 	}
	// }
	// Skip user provided query param fields for now -> need to get reads working...

	// Check Zoom's `mandatoryDateObjects` to see a pattern for
	// handling required query params
	startTime := defaultTimeRange
	if !config.Since.IsZero() {
		startTime = config.Since
	}
	url.WithQueryParam("from_date", startTime.UTC().Format(time.RFC3339))

	// doc says required with `to_date` but we'll see...
	if !config.Until.IsZero() {
		url.WithQueryParam("to_date", config.Until.Format(time.RFC3339))
	}

	count := defaultPageSize
	if config.PageSize > 0 {
		count = config.PageSize
	}
	url.WithQueryParam("count", strconv.Itoa(count))
}
