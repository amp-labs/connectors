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
	// JustCall pagination: max 100 per page for most endpoints.
	// https://developer.justcall.io/reference/call_list_v21
	defaultPerPage = "100"
	perPage50      = "50"
	perPage20      = "20"

	// JustCall datetime format: yyyy-mm-dd hh:mm:ss or yyyy-mm-dd.
	// https://developer.justcall.io/reference/call_list_v21
	datetimeFormat = "2006-01-02 15:04:05"
)

// objectsWithoutPagination lists objects that don't support per_page parameter.
var objectsWithoutPagination = map[string]bool{ //nolint:gochecknoglobals
	"webhooks": true,
}

// objectsWithLowerPageLimit lists objects with lower per_page limits.
var objectsWithLowerPageLimit = map[string]string{ //nolint:gochecknoglobals
	"messages":               perPage50,
	"whatsapp/messages":      perPage50,
	"campaigns":              perPage50,
	"calls_ai":               perPage20,
	"meetings_ai":            perPage20,
	"sales_dialer/campaigns": perPage50,
}

// objectsWithIncrementalSync lists objects that support from_datetime/to_datetime filtering.
// https://developer.justcall.io/reference/call_list_v21
var objectsWithIncrementalSync = map[string]bool{ //nolint:gochecknoglobals
	"calls":                  true,
	"texts":                  true,
	"calls_ai":               true,
	"meetings_ai":            true,
	"sales_dialer/calls":     true,
	"whatsapp/messages":      true,
	"threads":                true,
	"sales_dialer/campaigns": true,
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

	if hasPagination := !objectsWithoutPagination[params.ObjectName]; hasPagination {
		perPage := defaultPerPage
		if limit, ok := objectsWithLowerPageLimit[params.ObjectName]; ok {
			perPage = limit
		}

		url.WithQueryParam("per_page", perPage)
	}

	// Add incremental sync parameters if supported and provided.
	if objectsWithIncrementalSync[params.ObjectName] {
		if !params.Since.IsZero() {
			url.WithQueryParam("from_datetime", params.Since.Format(datetimeFormat))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("to_datetime", params.Until.Format(datetimeFormat))
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, path)
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
