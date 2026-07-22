package sendgrid

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/sendgrid/metadata"
	"github.com/spyzhov/ajson"
)

const (
	restAPIVersion  = "v3"
	defaultPageSize = "100"

	pageSizeKey = "page_size"
	limitKey    = "limit"
	offsetKey   = "offset"
)

// nolint:gochecknoglobals
var supportedReadObjects = datautils.NewStringSet(
	objectContacts,
	objectLists,
	objectSegments,
	objectSinglesends,
	objectTemplates,
	objectFieldDefinitions,
	objectVerifiedSenders,
	objectSenders,
	objectBounces,
	objectBlocks,
	objectSpamReports,
	objectUnsubscribes,
	objectInvalidEmails,
	objectASMGroups,
	objectCategories,
	objectSubusers,
	objectEventWebhookSettings,
	objectParseWebhookSettings,
)

// Marketing list endpoints that paginate with page_size + page_token and return _metadata.next.
// Docs: https://www.twilio.com/docs/sendgrid/api-reference/lists/get-all-lists
//
//nolint:gochecknoglobals
var pageTokenObjects = datautils.NewStringSet(
	objectLists,
	objectSinglesends,
	objectTemplates,
)

// Suppression / stats list endpoints that paginate with limit + offset.
// Docs: https://www.twilio.com/docs/sendgrid/api-reference/bounces-api/retrieve-all-bounces
//
//nolint:gochecknoglobals
var offsetObjects = datautils.NewStringSet(
	objectBounces,
	objectBlocks,
	objectSpamReports,
	objectUnsubscribes,
	objectInvalidEmails,
	objectCategories,
	objectSubusers,
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedReadObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	endpointURL, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	endpointURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, strings.TrimSpace(path))
	if err != nil {
		return nil, err
	}

	pageSize := readhelper.PageSizeWithDefaultStr(params, defaultPageSize)

	switch {
	case pageTokenObjects.Has(params.ObjectName):
		endpointURL.WithQueryParam(pageSizeKey, pageSize)
	case offsetObjects.Has(params.ObjectName):
		endpointURL.WithQueryParam(limitKey, pageSize)
		endpointURL.WithQueryParam(offsetKey, "0")
	}

	return endpointURL, nil
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	recordsKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)

	return common.ParseResult(
		resp,
		extractRecordsOptional(recordsKey),
		makeNextPageFunc(params.ObjectName, request.URL, recordsKey),
		readhelper.MakeMarshaledDataFuncWithId(nil, idFieldForObject(params.ObjectName)),
		params.Fields,
	)
}

func extractRecordsOptional(recordsKey string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(recordsKey)
	}
}

func makeNextPageFunc(objectName string, requestURL *url.URL, recordsKey string) common.NextPageFunc {
	switch {
	case pageTokenObjects.Has(objectName):
		return nextPageFromMetadataNext
	case offsetObjects.Has(objectName):
		return nextPageFromOffset(requestURL, recordsKey)
	default:
		return noNextPage
	}
}

func nextPageFromMetadataNext(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "_metadata").StrWithDefault("next", "")
}

func nextPageFromOffset(requestURL *url.URL, recordsKey string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if requestURL == nil {
			return "", nil
		}

		records, err := jsonquery.New(node).ArrayOptional(recordsKey)
		if err != nil {
			return "", err
		}

		limitStr := requestURL.Query().Get(limitKey)
		if limitStr == "" {
			limitStr = defaultPageSize
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || len(records) < limit {
			return "", nil //nolint:nilerr
		}

		offsetStr := requestURL.Query().Get(offsetKey)
		if offsetStr == "" {
			offsetStr = "0"
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return "", nil //nolint:nilerr
		}

		endpointURL, err := urlbuilder.FromRawURL(requestURL)
		if err != nil {
			return "", err
		}

		endpointURL.WithQueryParam(offsetKey, strconv.Itoa(offset+limit))

		return endpointURL.String(), nil
	}
}

func noNextPage(_ *ajson.Node) (string, error) {
	return "", nil
}

func idFieldForObject(objectName string) readhelper.IdFieldQuery {
	switch objectName {
	case objectBounces, objectBlocks, objectSpamReports, objectUnsubscribes, objectInvalidEmails:
		return readhelper.NewIdField("email")
	case objectCategories:
		return readhelper.NewIdField("category")
	case objectParseWebhookSettings:
		return readhelper.NewIdField("hostname")
	default:
		return readhelper.NewIdField("id")
	}
}
