package livestorm

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/livestorm/metadata"
	"github.com/spyzhov/ajson"
)

const (
	// Default page size for list endpoints (page[size]). https://developers.livestorm.co/
	defaultPageSize = "100"

	apiVersion                = "v1"
	objectEvents              = "events"
	objectSessionChatMessages = "session_chat_messages"
	objectJobs                = "jobs"
)

// nolint:gochecknoglobals
var readSupportedObjects = datautils.NewStringSet(
	"events",
	"people",
	"people_attributes",
	"session_chat_messages",
	"jobs",
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if !readSupportedObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	u, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.api+json")

	return req, nil
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	switch params.ObjectName {
	case "":
		return nil, common.ErrMissingObjects
	case objectSessionChatMessages:
		return c.buildSessionChatMessagesReadURL(params)
	case objectJobs:
		return c.buildJobReadURL(params)
	default:
		return c.buildGenericReadURL(params)
	}
}

func (c *Connector) buildSessionChatMessagesReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	sessionID := strings.TrimSpace(params.Filter)
	if sessionID == "" {
		return nil, ErrSessionIDRequired
	}

	endpointURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "sessions", sessionID, "chat_messages")
	if err != nil {
		return nil, err
	}

	endpointURL.WithQueryParam("page[number]", "0")
	endpointURL.WithQueryParam("page[size]", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

	return endpointURL, nil
}

func (c *Connector) buildJobReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	jobID := strings.TrimSpace(params.Filter)
	if jobID == "" {
		return nil, ErrJobIDRequired
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "jobs", jobID)
}

func (c *Connector) buildGenericReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	endpointURL, err := buildVersionedPathURL(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	endpointURL.WithQueryParam("page[number]", "0")
	endpointURL.WithQueryParam("page[size]", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

	if params.ObjectName == objectEvents {
		applyEventTimeFilters(endpointURL, params)
	}

	return endpointURL, nil
}

func buildVersionedPathURL(baseURL string, path string) (*urlbuilder.URL, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return urlbuilder.New(baseURL, apiVersion)
	}

	return urlbuilder.New(baseURL, apiVersion, path)
}

func applyEventTimeFilters(u *urlbuilder.URL, params common.ReadParams) {
	if !params.Since.IsZero() {
		u.WithQueryParam("filter[updated_since]", strconv.FormatInt(params.Since.Unix(), 10))
	}

	if !params.Until.IsZero() {
		u.WithQueryParam("filter[updated_until]", strconv.FormatInt(params.Until.Unix(), 10))
	}
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		extractJSONAPIResourceNodes,
		nextPageLivestorm(request),
		readhelper.MakeMarshaledDataFuncWithId(flattenJSONAPIResourceForFields, readhelper.NewIdField("id")),
		params.Fields,
	)
}

func nextPageLivestorm(previous *http.Request) common.NextPageFunc {
	return func(root *ajson.Node) (string, error) {
		if previous == nil || previous.URL == nil {
			return "", nil
		}

		nextPage, err := readNextPageFromMeta(root)
		if err != nil {
			return "", err
		}

		if nextPage != "" {
			return nextPage, nil
		}

		return buildNextPageFromCounters(previous, root)
	}
}

func readNextPageFromMeta(root *ajson.Node) (string, error) {
	nextStr, err := jsonquery.New(root, "meta").StringOptional("next_page")
	if err != nil {
		return "", err
	}

	if nextStr == nil || *nextStr == "" {
		return "", nil
	}

	return *nextStr, nil
}

func buildNextPageFromCounters(previous *http.Request, root *ajson.Node) (string, error) {
	pageCount, current, err := readPageCounters(root)
	if err != nil {
		return "", err
	}

	if pageCount == nil || current == nil {
		return "", nil
	}

	pc := *pageCount
	cur := *current

	if pc <= 0 || cur+1 >= pc {
		return "", nil
	}

	u, err := urlbuilder.FromRawURL(previous.URL)
	if err != nil {
		return "", err
	}

	u.WithQueryParam("page[number]", strconv.FormatInt(cur+1, 10))

	return u.String(), nil
}

func readPageCounters(root *ajson.Node) (*int64, *int64, error) {
	pageCount, err := jsonquery.New(root, "meta").IntegerOptional("page_count")
	if err != nil {
		return nil, nil, err
	}

	current, err := jsonquery.New(root, "meta").IntegerOptional("current_page")
	if err != nil {
		return nil, nil, err
	}

	return pageCount, current, nil
}
