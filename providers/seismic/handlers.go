package seismic

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, c.modulePath(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	if !params.Since.IsZero() {
		url.WithQueryParam("modifiedAtStartTime", params.Since.Format(time.RFC3339))
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) modulePath() string {
	module := supportedModules[c.Module()]

	return module.Path()
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.GetRecordsUnderJSONPath(""),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
