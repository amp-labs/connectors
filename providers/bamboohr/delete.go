package bamboohr

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/bamboohr/metadata"
)

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if !deleteSupportedObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	endpointURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path, params.RecordId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpointURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) parseDeleteResponse(
	_ context.Context,
	_ common.DeleteParams,
	_ *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	return &common.DeleteResult{Success: true}, nil
}
