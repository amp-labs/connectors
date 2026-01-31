package salesfinity

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}
	// The Salesfinity API exposes contact lists in two ways: "contact-lists" for Create
	// and "contact-lists/csv" for reading/deleting the same lists in CSV form.
	// Delete is done via the CSV endpoint, so we append "csv" to the path when the object is "contact-lists".
	if params.ObjectName == "contact-lists" {
		url.AddPath("csv")
	}
	url.AddPath(params.RecordId)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create delete request: %w", err)
	}

	return req, nil
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if resp.Code != http.StatusNoContent && resp.Code != http.StatusOK {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, resp.Code)
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
