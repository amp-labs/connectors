package confluence

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/spyzhov/ajson"
)

// DefaultPageSize is similar across all endpoints. One example:
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-label/#api-labels-get
const DefaultPageSize = "250"

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

	url.WithQueryParam("limit", DefaultPageSize)

	return url, nil
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		common.ExtractOptionalRecordsFromPath("results"),
		a.makeNextRecordsURL(resp),
		common.GetMarshaledData,
		params.Fields,
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

func (a *Adapter) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	// TODO needs implementation.
	return nil, common.ErrNotImplemented
}

func (a *Adapter) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	// TODO needs implementation.
	return nil, common.ErrNotImplemented
}
