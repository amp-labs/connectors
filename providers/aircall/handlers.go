package aircall

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion = "v1"

	// Pagination constants – see Aircall API pagination docs:
	// https://developer.aircall.io/api-references/#pagination
	//   - minimum: 1
	//   - default: 20
	//   - maximum: 50
	aircallMaxPerPage = "50"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// Default pagination
	url.WithQueryParam("per_page", aircallMaxPerPage)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(params.ObjectName),
		makeNextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Aircall returns "meta" object with "next_page_link"
		return jsonquery.New(node, "meta").StrWithDefault("next_page_link", "")
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	// TODO: Implement write support
	return nil, common.ErrNotImplemented
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	// TODO: Implement write support
	return nil, common.ErrNotImplemented
}
