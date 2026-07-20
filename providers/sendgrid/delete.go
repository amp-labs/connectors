package sendgrid

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/providers/sendgrid/metadata"
)

// Delete-supported objects.
//
//nolint:gochecknoglobals
var supportedDeleteObjects = datautils.NewStringSet(
	objectLists,
	objectTemplates,
	objectASMGroups,
)

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if !supportedDeleteObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	endpointURL, err := urlbuilder.New(
		c.ProviderInfo().BaseURL,
		restAPIVersion,
		strings.TrimSpace(path),
		params.RecordId,
	)
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
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if !httpkit.Status2xx(response.Code) {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	return &common.DeleteResult{Success: true}, nil
}
