package fastspring

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/httpkit"
)

// FastSpring delete API references:
// - Delete product: https://developer.fastspring.com/reference/delete-a-product
// - Cancel subscription: https://developer.fastspring.com/reference/cancel-a-subscription
func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	deleteURL, err := buildDeleteURL(c.ProviderInfo().BaseURL, params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func buildDeleteURL(baseURL string, params common.DeleteParams) (*urlbuilder.URL, error) {
	switch params.ObjectName {
	case objectProducts:
		return urlbuilder.New(baseURL, objectProducts, params.RecordId)
	case objectSubscriptions:
		return urlbuilder.New(baseURL, objectSubscriptions, params.RecordId)
	default:
		return nil, common.ErrOperationNotSupportedForObject
	}
}

func (c *Connector) parseDeleteResponse(
	_ context.Context,
	_ common.DeleteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if !httpkit.Status2xx(response.Code) {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	return &common.DeleteResult{Success: true}, nil
}
