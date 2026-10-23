package clarify

import (
	"context"
	"net/http"
	"strconv"

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

	// Datetime filter operators: `Is after` is exclusive and `Is on or before` inclusive,
	// matching ReadParams (Since exclusive, Until inclusive).
	// https://developer.clarify.ai/docs/api-basics/filtering#operators-by-field-type
	updatedSinceParam = "filter[_updated_at][Is after]"
	updatedUntilParam = "filter[_updated_at][Is on or before]"
)

// Operator names contain spaces, which the docs encode as %20 rather than `+`.
// https://developer.clarify.ai/docs/api-basics/filtering#filter-with-an-operator
var spaceEncoding = map[string]string{"+": "%20"} // nolint:gochecknoglobals

// buildReadRequest lists records with page[limit]/page[offset]. Records (built-in and custom
// objects) are filtered on `_updated_at` by the API; users are filtered in parseReadResponse
// because GET /users accepts no filter.
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
			readURL.WithQueryParam(updatedSinceParam, datautils.Time.FormatRFC3339inUTCWithMilliseconds(params.Since))
		}

		if !params.Until.IsZero() {
			readURL.WithQueryParam(updatedUntilParam, datautils.Time.FormatRFC3339inUTCWithMilliseconds(params.Until))
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
	nextPage := makeNextPageFunc(request.URL)

	filter := readhelper.MakeIdentityFilterFunc(nextPage)
	if !isRecordObject(params.ObjectName) {
		filter = makeUpdatedAtFilter(nextPage)
	}

	return common.ParseResultFiltered(params, response,
		common.MakeRecordsFunc("data"),
		filter,
		readhelper.MakeMarshaledDataFuncWithId(flattenRecord, readhelper.NewIdField("id")),
		params.Fields,
	)
}
