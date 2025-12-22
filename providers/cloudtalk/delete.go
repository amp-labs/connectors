package cloudtalk

import (
	"context"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	// 1. Get the read path from metadata (e.g. "/contacts/index.json")
	readPath, err := c.getURLPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	// 2. Determine base resource path (e.g. "/contacts")
	baseResource := strings.TrimSuffix(readPath, "/index.json")

	// 3. Construct Delete URL
	// 3. Construct Delete URL
	// Pattern: /contacts/delete/123.json
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, baseResource, "delete", params.RecordId+".json")
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	req *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	// We assume success if no error occurred
	return &common.DeleteResult{
		Success: true,
	}, nil
}
