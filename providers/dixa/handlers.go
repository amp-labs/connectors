package dixa

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const restAPIVersion = "v1"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	objectURL := url.String()

	if params.NextPage != "" {
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL)
		if err != nil {
			return nil, err
		}

		objectURL = url.String() + params.NextPage.String()
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, objectURL, nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		constructRecords(params.ObjectName),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
