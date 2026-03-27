package housecallpro

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/housecallPro/metadata"
	"github.com/spyzhov/ajson"
)

const (
	readSortDirection = "desc"
	readPerPageCap    = 100
	readPageSizeCap   = 200
)

// Price book list payloads generally use "uuid"; other objects use "id".
var readIDFieldByObject = datautils.NewDefaultMap( //nolint:gochecknoglobals
	datautils.Map[string, readhelper.IdFieldQuery]{
		"price_book/material_categories": readhelper.NewIdField("uuid"),
		"price_book/services":            readhelper.NewIdField("uuid"),
		"price_book/materials":           readhelper.NewIdField("uuid"),
	},
	func(_ string) readhelper.IdFieldQuery {
		return readhelper.NewIdField("id")
	},
)

// Most list endpoints take page size as page_size; checklists and routes use per_page.
var readListPageSize = datautils.NewDefaultMap( //nolint:gochecknoglobals
	map[string]listPageSize{
		"checklists": {param: "per_page", defaultStr: strconv.Itoa(readPerPageCap), max: readPerPageCap},
		"routes":     {param: "per_page", defaultStr: strconv.Itoa(readPerPageCap), max: readPerPageCap},
	},
	func(string) listPageSize {
		return listPageSize{param: "page_size", defaultStr: strconv.Itoa(readPageSizeCap), max: readPageSizeCap}
	},
)

type listPageSize struct {
	param      string // "page_size" or "per_page"
	defaultStr string
	max        int // maximum allowed PageSize for this object (inclusive)
}

func (s listPageSize) queryValue(params common.ReadParams) string {
	if params.PageSize <= 0 {
		return s.defaultStr
	}

	if params.PageSize > s.max {
		return s.defaultStr
	}

	return strconv.Itoa(params.PageSize)
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	path, err := metadata.Schemas.FindURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	ps := readListPageSize.Get(params.ObjectName)
	url.WithQueryParam(ps.param, ps.queryValue(params))

	spec := objectReadSpecs.Get(params.ObjectName)
	if spec.timeKey != "" {
		url.WithQueryParam("sort_by", spec.timeKey)
		url.WithQueryParam("sort_direction", readSortDirection)
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

	responseKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)
	if responseKey == "" {
		responseKey = params.ObjectName
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc(responseKey),
		makeFilterFunc(params, reqURL),
		readhelper.MakeMarshaledDataFuncWithId(nil, readIDFieldByObject.Get(params.ObjectName)),
		params.Fields,
	)
}

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Paginated list responses include page metadata at the root
		//  {
		//   "page": 1,
		//   "page_size": 1,
		//   "total_pages": 1,
		//   "total_items": 1,
		//   .....
		// }
		rootQuery := jsonquery.New(node)

		page, err := rootQuery.IntegerWithDefault("page", 1)
		if err != nil {
			return "", err
		}

		totalPages, err := rootQuery.IntegerWithDefault("total_pages", 0)
		if err != nil {
			return "", err
		}

		if totalPages == 0 {
			totalPages, err = rootQuery.IntegerWithDefault("total_pages_count", 0)
			if err != nil {
				return "", err
			}
		}

		if totalPages == 0 || page >= totalPages {
			return "", nil
		}

		reqLink.WithQueryParam("page", strconv.FormatInt(page+1, 10))

		return reqLink.String(), nil
	}
}
