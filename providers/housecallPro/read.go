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

const defaultPageSize = "200"
const readSortDirection = "desc"

// Price book list payloads use "uuid"; other objects use "id".
var readIDFieldByObject = datautils.NewDefaultMap( //nolint:gochecknoglobals
	datautils.Map[string, readhelper.IdFieldQuery]{
		"price_book/material_categories": readhelper.NewIdField("uuid"),
		"price_book/price_forms":         readhelper.NewIdField("uuid"),
		"price_book/services":            readhelper.NewIdField("uuid"),
	},
	func(_ string) readhelper.IdFieldQuery {
		return readhelper.NewIdField("id")
	},
)

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
	pageSize := readhelper.PageSizeWithDefaultStr(params, defaultPageSize)
	url.WithQueryParam("per_page", pageSize)
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
		//   "page": 0,
		//   "page_size": 0,
		//   "total_pages": 0,
		//   "total_items": 0,
		//   .....
		// }
		q := jsonquery.New(node)

		page, err := q.IntegerWithDefault("page", 1)
		if err != nil {
			return "", err
		}

		totalPages, err := q.IntegerWithDefault("total_pages", 0)
		if err != nil {
			return "", err
		}

		if totalPages == 0 {
			totalPages, err = q.IntegerWithDefault("total_pages_count", 0)
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
