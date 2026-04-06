package fastspring

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// FastSpring delete API references:
// - Delete product: https://developer.fastspring.com/reference/delete-a-product
// - Cancel subscription: https://developer.fastspring.com/reference/cancel-a-subscription

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	if err := validateDeleteParams(params); err != nil {
		return nil, err
	}

	baseURL := c.ProviderInfo().BaseURL

	var (
		deleteURL *urlbuilder.URL
		err       error
	)

	switch params.ObjectName {
	case objectProducts:
		deleteURL, err = urlbuilder.New(baseURL, objectProducts, params.RecordId)
	case objectSubscriptions:
		deleteURL, err = urlbuilder.New(baseURL, objectSubscriptions, params.RecordId)
	default:
		return nil, common.ErrOperationNotSupportedForObject
	}

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

func validateDeleteParams(params common.DeleteParams) error {
	switch params.ObjectName {
	case objectProducts, objectSubscriptions:
		return nil
	default:
		return common.ErrOperationNotSupportedForObject
	}
}

func (c *Connector) parseDeleteResponse(
	_ context.Context,
	_ common.DeleteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	switch response.Code {
	case http.StatusOK, http.StatusAccepted, http.StatusNoContent:
		return &common.DeleteResult{Success: true}, nil
	default:
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}
}
