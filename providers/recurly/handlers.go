package recurly

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

var (
	ApiVersionHeader = "application/vnd.recurly.v2021-02-25+json" //nolint:gochecknoglobals
	limit            = "200"                                      //nolint:gochecknoglobals
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var url *urlbuilder.URL

	var err error

	if params.NextPage == "" {
		url, err = buildFirstPageURL(c.ProviderInfo().BaseURL, params)
	} else {
		url, err = buildNextPageURL(c.ProviderInfo().BaseURL, params.NextPage.String())
	}

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	// version is required in the Accept header
	// see: https://recurly.com/developers/api/v2021-02-25/index.html#section/Getting-Started/Versioning
	req.Header.Set("Accept", ApiVersionHeader)

	return req, nil
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
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
