package confluence

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/spyzhov/ajson"
)

// DefaultPageSize is similar across all endpoints. One example:
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-label/#api-labels-get
const DefaultPageSize = "250"

type readObjectSpec struct {
	sortingQuery string
	sortOrder    readhelper.TimeOrder
	timeField    string
}

var readObjectsSpec = datautils.NewDefaultMap(map[string]readObjectSpec{ // nolint:gochecknoglobals
	"attachments": {
		sortingQuery: "-created-date",
		sortOrder:    readhelper.ReverseOrder,
		timeField:    "createdAt",
	},
	"blogposts": {
		sortingQuery: "-created-date",
		sortOrder:    readhelper.ReverseOrder,
		timeField:    "createdAt",
	},
	"inline-comments": {
		sortingQuery: "-created-date",
		sortOrder:    readhelper.ReverseOrder,
		timeField:    "resolutionLastModifiedAt",
	},
	"pages": {
		sortingQuery: "-created-date",
		sortOrder:    readhelper.ReverseOrder,
		timeField:    "createdAt",
	},
	"spaces": {
		sortingQuery: "",
		sortOrder:    readhelper.Unordered,
		timeField:    "createdAt",
	},
	"tasks": {
		sortingQuery: "",
		sortOrder:    readhelper.Unordered,
		timeField:    "updatedAt",
	},
}, func(objectName string) readObjectSpec {
	// Default case: no sorting query and no response field available.
	return readObjectSpec{
		sortingQuery: "",
		sortOrder:    readhelper.Unordered,
		timeField:    "",
	}
})

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := a.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (a *Adapter) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := a.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if spec := readObjectsSpec.Get(params.ObjectName); spec.sortingQuery != "" {
		// Sorting example:
		// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-blog-post/#api-blogposts-get
		url.WithQueryParam("sort", spec.sortingQuery)
	}

	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, DefaultPageSize))

	return url, nil
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResultFiltered(params,
		resp,
		common.MakeRecordsFunc("results"),
		a.makeFilterFunc(params, resp),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}

func (a *Adapter) makeFilterFunc(params common.ReadParams, resp *common.JSONHTTPResponse) common.RecordsFilterFunc {
	spec := readObjectsSpec.Get(params.ObjectName)

	nextPageFunc := a.makeNextRecordsURL(resp)
	if spec.timeField == "" {
		// Object has no time field that can be used for connector-side filtering.
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	return readhelper.MakeTimeFilterFunc(
		spec.sortOrder,
		readhelper.NewTimeBoundary(),
		spec.timeField,
		time.RFC3339,
		nextPageFunc,
	)
}

// Next page is communicated via `Link` header under the `next` rel.
// https://developer.atlassian.com/cloud/confluence/rest/v2/intro/#about
func (a *Adapter) makeNextRecordsURL(resp *common.JSONHTTPResponse) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		headerLink := httpkit.HeaderLink(resp, "next")
		if headerLink == "" {
			return "", nil
		}

		url, err := a.getRawModuleURL()
		if err != nil {
			return "", err
		}

		url.AddPath(headerLink)

		return url.String(), nil
	}
}
