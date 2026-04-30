package livestorm

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
	"github.com/amp-labs/connectors/providers/livestorm/metadata"
	"github.com/spyzhov/ajson"
)

// Default page size for list endpoints (page[size]). https://developers.livestorm.co/
const defaultPageSize = "100"

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

	if params.ObjectName == "" {
		return nil, common.ErrMissingObjects
	}

	switch params.ObjectName {
	case objectSessionChatMessages:
		sessionID := strings.TrimSpace(params.Filter)
		if sessionID == "" {
			return nil, ErrSessionIDRequired
		}

		u, err := urlbuilder.New(c.ProviderInfo().BaseURL, "v1", "sessions", sessionID, "chat_messages")
		if err != nil {
			return nil, err
		}

		u.WithQueryParam("page[number]", "0")
		u.WithQueryParam("page[size]", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

		return u, nil
	case objectJobs:
		jobID := strings.TrimSpace(params.Filter)
		if jobID == "" {
			return nil, ErrJobIDRequired
		}

		return urlbuilder.New(c.ProviderInfo().BaseURL, "v1", "jobs", jobID)
	default:
		path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
		if err != nil {
			return nil, err
		}

		u, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
		if err != nil {
			return nil, err
		}

		u.WithQueryParam("page[number]", "0")
		u.WithQueryParam("page[size]", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

		if params.ObjectName == objectEvents {
			applyEventTimeFilters(u, params)
		}

		return u, nil
	}
}

const (
	objectEvents              = "events"
	objectSessionChatMessages = "session_chat_messages"
	objectJobs                = "jobs"
)

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
		func(root *ajson.Node) ([]map[string]any, error) {
			return extractJSONAPIDataRecords(root)
		},
		nextPageLivestorm(request),
		readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField("id")),
		params.Fields,
	)
}

func nextPageLivestorm(previous *http.Request) common.NextPageFunc {
	return func(root *ajson.Node) (string, error) {
		if previous == nil || previous.URL == nil {
			return "", nil
		}

		nextStr, err := jsonquery.New(root, "meta").StringOptional("next_page")
		if err != nil {
			return "", err
		}

		if nextStr != nil && *nextStr != "" {
			return *nextStr, nil
		}

		pageCount, err := jsonquery.New(root, "meta").IntegerOptional("page_count")
		if err != nil {
			return "", err
		}

		current, err := jsonquery.New(root, "meta").IntegerOptional("current_page")
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

		cloned, err := url.Parse(previous.URL.String())
		if err != nil {
			return "", err
		}

		u, err := urlbuilder.FromRawURL(cloned)
		if err != nil {
			return "", err
		}

		u.WithQueryParam("page[number]", strconv.FormatInt(cur+1, 10))

		return u.String(), nil
	}
}
