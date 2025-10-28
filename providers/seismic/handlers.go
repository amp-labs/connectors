package seismic

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const apiVersion = "v2"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ModuleInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if !params.Since.IsZero() {
		url.WithQueryParam("modifiedAtStartTime", params.Since.Format(time.RFC3339))
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
		common.ExtractRecordsFromPath(""),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
