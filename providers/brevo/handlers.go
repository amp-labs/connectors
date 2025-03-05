package brevo

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/brevo/metadata"
)

var apiVersion = "v3" //nolint:gochecknoglobals

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	if params.NextPage != "" {
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// first page pagination
	if supportLimitAndOffset.Has(params.ObjectName) {
		url.WithQueryParam("limit", strconv.Itoa(pageSize))
		url.WithQueryParam("offset", "0")
	}

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
		common.GetRecordsUnderJSONPath(responseFieldName),
		nextRecordsURL(requestURL),
		common.GetMarshaledData,
		params.Fields,
	)
}
