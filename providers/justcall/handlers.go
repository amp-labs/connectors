package justcall

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/justcall/metadata"
	"github.com/spyzhov/ajson"
)

const (
	// JustCall pagination: max 100 per page
	// https://developer.justcall.io/reference/call_list_v21
	justcallMaxPerPage = "100"
)

// objectsWithoutPagination lists objects that don't support per_page parameter.
var objectsWithoutPagination = map[string]bool{
	"webhooks": true,
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	url, err := c.buildURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if !objectsWithoutPagination[params.ObjectName] {
		url.WithQueryParam("per_page", justcallMaxPerPage)
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(c.ModuleID, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.BaseURL, path)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("data"),
		makeNextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return jsonquery.New(node).StrWithDefault("next_page_link", "")
	}
}
