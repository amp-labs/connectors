package acculynx

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// AccuLynx delete API references:
//   - Delete AR Owner from job:     https://apidocs.acculynx.com/reference/deleteAROwnerFromJob
//   - Delete Sales Owner from job:  https://apidocs.acculynx.com/reference/deleteSalesOwnerFromJob
//
// AccuLynx exposes no DELETE for top-level /contacts or /jobs. The only DELETE
// endpoints in the V2 API remove a representative slot from an existing job;
// RecordId here is the jobId, ObjectName selects which representative slot
// (ar-owner or sales-owner) to clear.

const (
	deletableJobsARRepresentative    = "jobs/representatives/ar-owner"
	deletableJobsSalesRepresentative = "jobs/representatives/sales-owner"
)

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.buildDeleteURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
}

// buildDeleteURL maps a deletable ObjectName + RecordId (= jobId) to its
// fully-qualified AccuLynx URL. Returns ErrOperationNotSupportedForObject for
// any object outside the small set AccuLynx allows deletes on.
func (c *Connector) buildDeleteURL(params common.DeleteParams) (*urlbuilder.URL, error) {
	baseURL := c.ProviderInfo().BaseURL

	switch params.ObjectName {
	case deletableJobsARRepresentative:
		return urlbuilder.New(baseURL, apiVersionPrefix,
			"jobs", params.RecordId, "representatives", "ar-owner")
	case deletableJobsSalesRepresentative:
		return urlbuilder.New(baseURL, apiVersionPrefix,
			"jobs", params.RecordId, "representatives", "sales-owner")
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
	if response.Code != http.StatusOK &&
		response.Code != http.StatusNoContent &&
		response.Code != http.StatusAccepted {
		return nil, common.ErrRequestFailed
	}

	return &common.DeleteResult{Success: true}, nil
}
