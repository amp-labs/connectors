package devrev

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/devrev/metadata"
	"github.com/spyzhov/ajson"
)

// Pagination (DevRev list.*):
//   - Typical default page size in API docs: 50.
//   - Many endpoints document a maximum of 100 records per page.
//   - Newer list endpoints enforce a lower maximum: 100 is rejected; use at most 99 or the API returns 400.
const maxPageSize = 99

// objectsWithoutLimitParam lists object names whose list endpoints do not accept
// the limit query parameter (API returns 400 invalid_field for limit).
var objectsWithoutLimitParam = datautils.NewSet( //nolint:gochecknoglobals
	"auth-tokens",
	"dev-orgs.auth-connections",
	"schemas.subtypes",
	"webhooks",
)

// objectsWithModifiedDateFilter lists object names whose list endpoints support
// modified_date.after and modified_date.before query parameter.
var objectsWithModifiedDateFilter = datautils.NewSet( //nolint:gochecknoglobals
	"accounts",
	"conversations",
	"rev-orgs",
	"rev-users",
	"sla-trackers",
)

// objectsWithoutModifiedDate lists object names whose list response records do not
// include modified_date.
var objectsWithoutModifiedDate = datautils.NewSet( //nolint:gochecknoglobals
	"vistas.groups",
)

// objectsWithoutSortBy lists object names whose list endpoints do not accept
// or support the sort_by query parameter.
// doc doesn't describe the default sorting behavior.
var objectsWithoutSortBy = datautils.NewSet( //nolint:gochecknoglobals
	"articles",
	"auth-tokens",
	"code-changes",
	"conversations",
	"directories",
	"org-schedules",
	"question-answers",
	"schemas.subtypes",
	"webhooks",
	"vistas.groups",
)

func pageSize(params common.ReadParams) string {
	if params.PageSize <= 0 {
		return strconv.Itoa(maxPageSize)
	}

	if params.PageSize > maxPageSize {
		return strconv.Itoa(maxPageSize)
	}

	return strconv.Itoa(params.PageSize)
}

// buildReadRequest builds the HTTP request for listing objects.
// Pagination is cursor-based (next_cursor). For objects in objectsWithModifiedDateFilter, Since/Until
// are sent as modified_date.after and modified_date.before; other objects
// are filtered by modified_date on the client.
// See https://devrev.dev/docs.
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		// Next page
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}
	// First page
	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	pageSize := pageSize(params)
	if !objectsWithoutLimitParam.Has(params.ObjectName) {
		url.WithQueryParam("limit", pageSize)
	}

	if !objectsWithoutSortBy.Has(params.ObjectName) {
		// sort_by is string[];
		url.WithQueryParamList("sort_by", []string{"modified_date:desc"})
	}
	// if object supports modified date filter, add the since and until query params
	if objectsWithModifiedDateFilter.Has(params.ObjectName) {
		if !params.Since.IsZero() {
			url.WithQueryParam("modified_date.after", params.Since.UTC().Format(time.RFC3339))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("modified_date.before", params.Until.UTC().Format(time.RFC3339))
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	reqURL, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}
	// Get the response key
	/*{
		"accounts": [
		  ....
		],
		"next_cursor": "cursor_page_2",
		"total": 2
	  }
	*/
	responseKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)
	if responseKey == "" {
		responseKey = params.ObjectName
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc(responseKey),
		makeFilterFunc(params, reqURL),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}

// makeFilterFunc returns the identity filter when the API does time filtering
// (objectsWithModifiedDateFilter) or when rows have no modified_date
// (objectsWithoutModifiedDate). Otherwise it uses a time filter on modified_date;
// objectsWithoutSortBy use Unordered order so Since/Until still apply without
// assuming list sort order or early-stopping pagination.
func makeFilterFunc(params common.ReadParams, reqURL *urlbuilder.URL) common.RecordsFilterFunc {
	nextPageFunc := makeNextRecordsURL(reqURL)

	if objectsWithModifiedDateFilter.Has(params.ObjectName) || objectsWithoutModifiedDate.Has(params.ObjectName) {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	order := readhelper.ReverseOrder
	if objectsWithoutSortBy.Has(params.ObjectName) {
		order = readhelper.Unordered
	}

	return readhelper.MakeTimeFilterFunc(
		order,
		readhelper.NewTimeBoundary(),
		"modified_date",
		time.RFC3339,
		nextPageFunc,
	)
}

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextCursor, err := jsonquery.New(node).TextWithDefault("next_cursor", "")
		if err != nil {
			return "", err
		}

		if nextCursor == "" {
			return "", nil
		}

		reqLink.WithQueryParam("cursor", nextCursor)

		return reqLink.String(), nil
	}
}
