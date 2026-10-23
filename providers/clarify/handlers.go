package clarify

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	apiVersion = "v1"

	// List endpoints accept page[limit] up to 500 and return 422 above it.
	// https://developer.clarify.ai/docs/api-basics/pagination
	maxPageSize = 500

	pageLimitParam  = "page[limit]"
	pageOffsetParam = "page[offset]"

	// Datetime filter operators.
	// https://developer.clarify.ai/docs/api-basics/filtering#operators-by-field-type
	// Records carry `_updated_at` in milliseconds, but the filter compares whole seconds: live, a record at
	// 19:20:32.797Z matched both `Is after 19:20:32.797Z` and `Is on or after 19:20:32.798Z`, and one at
	// 06:41:04.124Z missed both `Is on or before 06:41:04.124Z` and `Is before 06:41:04.125Z`.
	// So the request asks for every whole second the window touches, [floor(Since), floor(Until)+1s),
	// and parseReadResponse drops what falls outside (Since, Until] at millisecond precision.
	updatedSinceParam = "filter[_updated_at][Is on or after]"
	updatedUntilParam = "filter[_updated_at][Is before]"
)

// Operator names contain spaces, which the docs encode as %20 rather than `+`.
// https://developer.clarify.ai/docs/api-basics/filtering#filter-with-an-operator
var spaceEncoding = map[string]string{"+": "%20"} // nolint:gochecknoglobals

// buildReadRequest lists records with page[limit]/page[offset]. Records (built-in and custom
// objects) are narrowed on `_updated_at` by the API to whole seconds; parseReadResponse applies the
// exact window to them and to users, whose endpoint (GET /users) accepts no filter.
// https://developer.clarify.ai/docs/api-basics/pagination
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	readURL, err := c.readURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, readURL, nil)
}

func (c *Connector) readURL(params common.ReadParams) (string, error) {
	// The next page is the first request with an advanced page[offset]; see makeNextPageFunc.
	if params.NextPage != "" {
		return params.NextPage.String(), nil
	}

	readURL, err := c.listURL(params.ObjectName)
	if err != nil {
		return "", err
	}

	readURL.WithQueryParam(pageLimitParam,
		strconv.Itoa(min(readhelper.PageSizeWithDefault(params, maxPageSize), maxPageSize)))
	readURL.WithQueryParam(pageOffsetParam, "0")

	if isRecordObject(params.ObjectName) {
		if !params.Since.IsZero() {
			readURL.WithQueryParam(updatedSinceParam,
				datautils.Time.FormatRFC3339inUTC(params.Since.Truncate(time.Second)))
		}

		if !params.Until.IsZero() {
			readURL.WithQueryParam(updatedUntilParam,
				datautils.Time.FormatRFC3339inUTC(params.Until.Truncate(time.Second).Add(time.Second)))
		}
	}

	readURL.AddEncodingExceptions(spaceEncoding)

	return readURL.String(), nil
}

func (c *Connector) listURL(objectName string) (*urlbuilder.URL, error) {
	if objectName == objectNameUsers {
		return c.workspaceURL("users")
	}

	return c.workspaceURL("objects", objectName, "resources")
}

func (c *Connector) workspaceURL(path ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, append([]string{apiVersion, "workspaces", c.Workspace}, path...)...)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResultFiltered(params, response,
		common.MakeRecordsFunc("data"),
		makeUpdatedAtFilter(makeNextPageFunc(request.URL)),
		readhelper.MakeMarshaledDataFuncWithId(flattenRecord, readhelper.NewIdField("id")),
		params.Fields,
	)
}
