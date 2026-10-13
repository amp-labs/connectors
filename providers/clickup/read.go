package clickup

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = "100"

var objectsV3 = datautils.NewSet( // nolint:gochecknoglobals
	"channels",
	"docs",
)

type readObjectSpec struct {
	since string
	until string
}

var incrementalReadSpecs = map[string]readObjectSpec{ // nolint:gochecknoglobals
	"tasks": {
		// https://developer.clickup.com/reference/gettasks#query-params
		// Unix milliseconds.
		since: "date_updated_gt",
		until: "date_updated_lt",
	},
	"time_entries": {
		// https://developer.clickup.com/reference/gettimeentrieswithinadaterange#query-params
		// Unix milliseconds.
		since: "start_date",
		until: "end_date",
	},
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	// For first page, construct the URL
	url, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if objectsV3.Has(params.ObjectName) {
		// Cursor based pagination with page size controls.
		url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))
	} else {
		// Objects V2
		// Offset based pagination without page size controls.
		url.WithQueryParam("page", "0")
	}

	if spec, ok := incrementalReadSpecs[params.ObjectName]; ok {
		if spec.since != "" && !params.Since.IsZero() {
			url.WithQueryParam(spec.since, datautils.Time.UnixMilli(params.Since))
		}

		if spec.until != "" && !params.Until.IsZero() {
			url.WithQueryParam(spec.until, datautils.Time.UnixMilli(params.Until))
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := objectNameReadResponseField.Get(params.ObjectName)

	previousURL, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	identifierField := "id"
	if params.ObjectName == "tags" {
		identifierField = "name"
	}

	var recordTransformer common.RecordTransformer = nil
	if params.ObjectName == "tasks" {
		recordTransformer = recordTransformerWithCustomFields
	}

	return common.ParseResult(
		response,
		common.MakeRecordsFunc(responseFieldName),
		nextRecordsUrl(previousURL),
		readhelper.MakeMarshaledDataFuncWithId(
			recordTransformer,
			readhelper.NewIdField(identifierField),
		),
		params.Fields,
	)
}

var objectNameReadResponseField = datautils.NewDefaultMap(map[string]string{ // nolint:gochecknoglobals
	"channels":         "data",
	"custom_items":     "custom_items",
	"customroles":      "custom_roles",
	"docs":             "docs",
	"folder_templates": "templates",
	"goals":            "goals",
	"groups":           "groups",
	"list_templates":   "templates",
	"spaces":           "spaces",
	"tags":             "data",
	"tasks":            "tasks",
	"teams":            "teams",
	"time_entries":     "data",
}, func(objectName string) string {
	return "data"
})

// nextRecordsUrl returns a pagination function that supports both API V2 and V3.
// If the response contains a non-nil `next_cursor`, cursor-based pagination is
// used (V3). Otherwise, `last_page` determines whether to increment the `page`
// query parameter (V2).
//
// API V2 has limited pagination support:
//   - The page size cannot be configured.
//   - Pagination is offset-based.
//
// API V3 is expected to address these limitations.
//
// Reference:
// https://developer.clickup.com/docs/tasks#large-requests-and-other-considerations
func nextRecordsUrl(previousUrl *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextCursor, err := jsonquery.New(node).StringOptional("next_cursor")
		if err != nil {
			return "", err
		}

		if nextCursor != nil {
			return nextRecordsUrlV3(previousUrl, *nextCursor)
		}

		lastPage, err := jsonquery.New(node).BoolWithDefault("last_page", true)
		if err != nil {
			return "", err
		}

		if hasMore := !lastPage; hasMore {
			return nextRecordsUrlV2(previousUrl)
		}

		return "", nil
	}
}

// nextRecordsUrlV2 returns the URL for the next page by incrementing the `page` query parameter.
func nextRecordsUrlV2(url *urlbuilder.URL) (string, error) {
	nextUrl := *url

	// parse the current page value form previous request url
	currentPage := 0
	currentPageStr, ok := url.GetFirstQueryParam("page")

	if ok && currentPageStr != "" {
		var err error

		currentPage, err = strconv.Atoi(currentPageStr)
		if err != nil {
			return "", nil //nolint:nilerr
		}
	}

	nextPage := currentPage + 1
	nextUrl.WithQueryParam("page", strconv.Itoa(nextPage))

	return nextUrl.String(), nil
}

// nextRecordsUrlV3 returns the URL for the next page by setting the `cursor` query parameter.
func nextRecordsUrlV3(url *urlbuilder.URL, cursor string) (string, error) {
	if cursor == "" {
		// Pagination is done.
		return "", nil
	}

	url.WithQueryParam("cursor", cursor)

	return url.String(), nil
}
