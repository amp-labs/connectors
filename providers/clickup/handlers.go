package clickup

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/clickup/metadata"
)

var apiVersion = "api/v2" //nolint:gochecknoglobals

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	// For first page, construct the URL
	url, err = c.RootClient.URL(apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("page", "0")

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)
	requestURL := request.URL

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(responseFieldName),
		nextRecordsURL(requestURL),
		common.GetMarshaledData,
		params.Fields,
	)
}
