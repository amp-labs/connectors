package gusto

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Gusto delete API conventions (App Integrations track):
//   - Most deletes are TOP-LEVEL by uuid:  DELETE /v1/{object}/{uuid}
//   - earning_types is COMPANY-SCOPED:     DELETE /v1/companies/{cid}/earning_types/{uuid}
//
// Workflow-style deletes (terminations, rehire) follow a different shape and
// are out of scope here — same reasoning as their POST/PUT counterparts in
// write.go (proxy-only).
//
// API references (slug = delete-v1-{path-with-dashes}):
// https://docs.gusto.com/app-integrations/reference/
//   - delete-v1-employee
//   - delete-v1-jobs-job_id
//   - delete-v1-compensations-compensation_id
//   - delete-v1-home_addresses-home_address_uuid
//   - delete-v1-work_addresses-work_address_uuid
//   - delete-v1-employee_benefits-employee_benefit_id
//   - delete-v1-company_benefits-company_benefit_id
//   - delete-department
//   - delete-v1-companies-company_id-earning_types-earning_type_uuid

// supportedDeleteObjects enumerates objects Gusto exposes a DELETE endpoint
// for. Many Gusto resources have no DELETE (companies, contractors,
// locations, payrolls, pay_schedules, garnishments, admins,
// contractor_payments, custom_fields). For those we reject with
// ErrOperationNotSupportedForObject.
//
//nolint:gochecknoglobals
var supportedDeleteObjects = datautils.NewStringSet(
	objectEmployees,
	objectJobs,
	objectCompensations,
	objectHomeAddresses,
	objectWorkAddresses,
	objectEmployeeBenefits,
	objectCompanyBenefits,
	objectDepartments,
	objectEarningTypes,
)

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	if !supportedDeleteObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.buildDeleteURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

// buildDeleteURL routes earning_types through the company-scoped path and all
// other supported objects through the flat top-level path.
func (c *Connector) buildDeleteURL(params common.DeleteParams) (*urlbuilder.URL, error) {
	baseURL := c.ProviderInfo().BaseURL

	if companyScopedUpdate.Has(params.ObjectName) {
		if c.companyId == "" {
			return nil, ErrMissingCompanyID
		}

		return urlbuilder.New(baseURL, "v1", "companies", c.companyId, params.ObjectName, params.RecordId)
	}

	return urlbuilder.New(baseURL, "v1", params.ObjectName, params.RecordId)
}

func (c *Connector) parseDeleteResponse(
	_ context.Context,
	_ common.DeleteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent && response.Code != http.StatusAccepted {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	return &common.DeleteResult{Success: true}, nil
}
