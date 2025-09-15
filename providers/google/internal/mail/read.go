package mail

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = "500"

// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.drafts/list
// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/list
// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.threads/list
var paginatedObjects = datautils.NewSet("drafts", "messages", "threads") // nolint:gochecknoglobals

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := a.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	// Add pagination query parameters.
	if paginatedObjects.Has(params.ObjectName) {
		url.WithQueryParam("maxResults", defaultPageSize)

		if params.NextPage != "" {
			url.WithQueryParam("pageToken", params.NextPage.String())
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := Schemas.LookupArrayFieldName(a.Module(), params.ObjectName)

	return common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath(responseFieldName),
		makeNextRecordsURL(params),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL(params common.ReadParams) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if !paginatedObjects.Has(params.ObjectName) {
			// There is no next page.
			return "", nil
		}

		return jsonquery.New(node).StrWithDefault("nextPageToken", "")
	}
}
