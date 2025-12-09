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
	// JustCall pagination: max 100 per page for most endpoints
	// https://developer.justcall.io/reference/call_list_v21
	defaultPerPage          = "100"
	whatsappPerPage         = "50" // WhatsApp endpoints have lower limit
	callsAIPerPage          = "20" // Calls AI max is 20
	salesDialerCampaignPage = "50" // Sales Dialer campaigns max is 50
)

// objectsWithoutPagination lists objects that don't support per_page parameter.
var objectsWithoutPagination = map[string]bool{
	"webhooks": true,
}

// objectsWithLowerPageLimit lists objects with lower per_page limits.
var objectsWithLowerPageLimit = map[string]string{
	"whatsapp/messages":       whatsappPerPage,
	"calls_ai":                callsAIPerPage,
	"sales_dialer/campaigns":  salesDialerCampaignPage,
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
		perPage := defaultPerPage
		if limit, ok := objectsWithLowerPageLimit[params.ObjectName]; ok {
			perPage = limit
		}

		url.WithQueryParam("per_page", perPage)
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
