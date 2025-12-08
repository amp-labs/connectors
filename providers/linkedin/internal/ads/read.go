package ads

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	liinternal "github.com/amp-labs/connectors/providers/linkedin/internal/linkedininternal"
)

func (c *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("LinkedIn-Version", liinternal.LinkedInVersion) // nolint:canonicalheader
	req.Header.Add("X-Restli-Protocol-Version", liinternal.ProtocolVersion)

	return req, nil
}

func (c *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("elements"),
		makeNextRecord(params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}
