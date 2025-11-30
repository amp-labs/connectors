package aircall

import (
	"context"
	"net/http"
	"strconv"

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
	aircallMaxPerPage     = "50"
	aircallMaxPageSizeInt = 50
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

	// Pagination
	pageSize := aircallMaxPerPage

	if params.PageSize > 0 {
		// Aircall maximum is 50
		if params.PageSize > aircallMaxPageSizeInt {
			pageSize = "50"
		} else {
			pageSize = strconv.Itoa(params.PageSize)
		}
	}

	url.WithQueryParam("per_page", pageSize)

	// Incremental sync: Add date range filters if provided
	// Aircall API uses Unix timestamps for 'from' and 'to' parameters
	// https://developer.aircall.io/api-references/#list-all-calls
	// Note: Not all objects support filtering by date (e.g. teams, tags).
	if supportsDateFiltering(params.ObjectName) {
		if !params.Since.IsZero() {
			url.WithQueryParam("from", strconv.FormatInt(params.Since.Unix(), 10))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("to", strconv.FormatInt(params.Until.Unix(), 10))
		}
	}

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

func supportsDateFiltering(objectName string) bool {
	switch objectName {
	case "calls", "users", "contacts", "numbers":
		return true
	case "teams", "tags":
		return false
	default:
		// Default to false to be safe
		return false
	}
}
