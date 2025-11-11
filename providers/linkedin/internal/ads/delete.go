package ads

import (
	"context"
	"fmt"
	"net/http"
	u "net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/linkedin/internal/shared"
)

func (c *Adapter) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := c.constructURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(u.QueryEscape(params.RecordId))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("LinkedIn-Version", shared.LinkedInVersion) // nolint:canonicalheader
	req.Header.Add("X-Restli-Protocol-Version", shared.ProtocolVersion)

	return req, nil
}

func (c *Adapter) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if resp.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, resp.Code)
	}

	// A successful delete returns 200 OK
	return &common.DeleteResult{
		Success: true,
	}, nil
}
