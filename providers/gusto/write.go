package gusto

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// Gusto write API conventions:
//   - CREATE: nested under parent
//       POST /v1/companies/{company_id}/{object}     (company-scoped)
//       POST /v1/employees/{employee_id}/{object}    (employee-scoped)
//       POST /v1/jobs/{job_id}/{object}              (job-scoped — compensations only)
//   - UPDATE: top-level by UUID
//       PUT /v1/{object}/{uuid}
//     Every PUT requires a `version` field in the body for optimistic
//     concurrency control. Callers must include it; we do not synthesize it.
//   - DELETE: not exposed by Gusto for the objects covered here. Most
//     resources prefer "deactivate" semantics (e.g., `terminations`,
//     `inactive=true`) over hard delete. Live testing will surface any
//     gaps; until then this connector is Write-only.
//
// Workflow operations (payroll calculate/submit/cancel, time-off
// approve/reject, etc.) are out of scope here — they don't fit the
// (objectName, recordId, recordData) shape of WriteConnector. Match the
// codebase precedent: Stripe charges/captures and QuickBooks invoice
// void/send are also proxy-only. Gusto's Proxy support is already enabled
// in providers/gusto.go.
//
// API references (slug = {method}-v1-{path-with-dashes}):
// https://docs.gusto.com/app-integrations/reference/
//   - put-v1-employees
//   - put-v1-locations
//   - put-v1-companies
//   - put-v1-compensations-compensation_id
//   - post-v1-employees
//   - post-v1-companies-company_id-locations
//   - post-v1-companies-company_uuid-contractors
//   - post-v1-employees-employee_id-employee_benefits
//   - post-v1-employees-employee_id-home_addresses
//   - post-v1-employees-employee_id-work_addresses
//   - post-v1-companies-company_id-earning_types
//   - post-v1-companies-company_id-company_benefits

// Path-template parent-id keys carried in RecordData for nested creates.
// Callers POSTing employee-scoped or job-scoped objects must include the
// matching key in their payload; the connector pulls it into the URL path
// and removes it from the body before sending.
const (
	parentIDKeyEmployeeID = "employee_id"
	parentIDKeyJobID      = "job_id"
)

var errMissingParentID = errors.New("gusto: missing parent id in record data")

// objectCompensations is write-only: read-side excludes it (would require
// employees → jobs → compensations 2-level traversal). Writes target it
// directly via /v1/jobs/{job_id}/compensations + PUT /v1/compensations/{uuid}.
const objectCompensations = "compensations"

// Write-supported objects, classified by parent scope at create time.
//
//nolint:gochecknoglobals
var (
	// companyScopedCreate creates under POST /v1/companies/{company_id}/{object}.
	// company_id comes from the connector struct (set at construction time).
	//
	// objectCustomFields is intentionally absent: Gusto's public API exposes
	// only GET endpoints for custom fields (definitions + per-employee values).
	// Definitions are managed through Gusto's admin UI, not the API. Verified
	// against the Gusto docs sitemap (App Integrations + Embedded Payroll
	// tracks). See providers/gusto/custom.go for the read-side handling.
	companyScopedCreate = datautils.NewStringSet(
		objectEmployees,
		objectLocations,
		objectDepartments,
		objectContractors,
		objectPayrolls,
		objectPaySchedules,
		objectEarningTypes,
		objectCompanyBenefits,
		objectAdmins,
		objectContractorPayments,
	)

	// employeeScopedCreate creates under POST /v1/employees/{employee_id}/{object}.
	// employee_id is extracted from RecordData.
	employeeScopedCreate = datautils.NewStringSet(
		objectJobs,
		objectEmployeeBenefits,
		objectGarnishments,
		objectHomeAddresses,
		objectWorkAddresses,
	)

	// jobScopedCreate creates under POST /v1/jobs/{job_id}/{object}.
	// job_id is extracted from RecordData.
	jobScopedCreate = datautils.NewStringSet(
		objectCompensations,
	)

	// updateOnly objects expose only PUT (top-level entities not created via the API).
	updateOnly = datautils.NewStringSet(
		objectCompanies,
	)

	// createOnly objects expose only POST. Most one-shot Gusto resources fall
	// here; Gusto either has no PUT for them or the operation is undocumented.
	createOnly = datautils.NewStringSet(
		objectAdmins,
		objectContractorPayments,
	)

	// companyScopedUpdate objects use a nested URL for PUT —
	// PUT /v1/companies/{company_id}/{object}/{record_id} — instead of the
	// flat PUT /v1/{object}/{record_id} pattern used by employees, locations,
	// jobs, etc. Confirmed via sitemap entries:
	//   put-v1-companies-company_id-earning_types-earning_type_uuid
	//   put-v1-companies-company_id-pay_schedules-pay_schedule_id
	//   put-v1-companies-company_id-payrolls
	companyScopedUpdate = datautils.NewStringSet(
		objectEarningTypes,
		objectPaySchedules,
		objectPayrolls,
	)

	// allWriteSupported is the union of every set above. Used by
	// validateWriteParams for a single membership check rather than 4 ORed
	// .Has() calls.
	allWriteSupported = datautils.MergeSets(
		companyScopedCreate,
		employeeScopedCreate,
		jobScopedCreate,
		updateOnly,
	)
)

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := validateWriteParams(params); err != nil {
		return nil, err
	}

	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	url, method, err := c.buildWriteURL(params, record)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// validateWriteParams rejects unsupported objects and operations early.
func validateWriteParams(params common.WriteParams) error {
	if !allWriteSupported.Has(params.ObjectName) {
		return common.ErrOperationNotSupportedForObject
	}

	if params.IsCreate() && updateOnly.Has(params.ObjectName) {
		return common.ErrOperationNotSupportedForObject
	}

	if params.IsUpdate() && createOnly.Has(params.ObjectName) {
		return common.ErrOperationNotSupportedForObject
	}

	return nil
}

// buildWriteURL routes to one of:
//   - PUT  /v1/{object}/{recordId}                                (top-level update)
//   - PUT  /v1/companies/{company_id}/{object}/{recordId}         (company-scoped update)
//   - POST /v1/companies/{company_id}/{object}                    (company-scoped create)
//   - POST /v1/employees/{employee_id}/{object}                   (employee-scoped create)
//   - POST /v1/jobs/{job_id}/{object}                             (job-scoped create)
//
// For nested creates, the parent ID is pulled from `record` (and removed so it
// is not echoed back into the request body).
func (c *Connector) buildWriteURL(
	params common.WriteParams, record map[string]any,
) (*urlbuilder.URL, string, error) {
	if params.IsUpdate() {
		return c.buildUpdateURL(params)
	}

	return c.buildCreateURL(params, record)
}

// buildUpdateURL handles PUT routing — flat top-level for most objects, nested
// under company for the few Gusto exposes that way (earning_types,
// pay_schedules, payrolls).
func (c *Connector) buildUpdateURL(params common.WriteParams) (*urlbuilder.URL, string, error) {
	baseURL := c.ProviderInfo().BaseURL

	if companyScopedUpdate.Has(params.ObjectName) {
		if c.companyID == "" {
			return nil, "", ErrMissingCompanyID
		}

		u, err := urlbuilder.New(baseURL, "v1", "companies", c.companyID, params.ObjectName, params.RecordId)

		return u, http.MethodPut, err
	}

	u, err := urlbuilder.New(baseURL, "v1", params.ObjectName, params.RecordId)

	return u, http.MethodPut, err
}

// buildCreateURL handles POST routing across the three create scopes.
func (c *Connector) buildCreateURL(
	params common.WriteParams, record map[string]any,
) (*urlbuilder.URL, string, error) {
	baseURL := c.ProviderInfo().BaseURL

	switch {
	case companyScopedCreate.Has(params.ObjectName):
		if c.companyID == "" {
			return nil, "", ErrMissingCompanyID
		}

		u, err := urlbuilder.New(baseURL, "v1", "companies", c.companyID, params.ObjectName)

		return u, http.MethodPost, err

	case employeeScopedCreate.Has(params.ObjectName):
		employeeID, ok := stringFromRecord(record, parentIDKeyEmployeeID)
		if !ok {
			return nil, "", fmt.Errorf("%w: %s", errMissingParentID, parentIDKeyEmployeeID)
		}

		u, err := urlbuilder.New(baseURL, "v1", "employees", employeeID, params.ObjectName)

		return u, http.MethodPost, err

	case jobScopedCreate.Has(params.ObjectName):
		jobID, ok := stringFromRecord(record, parentIDKeyJobID)
		if !ok {
			return nil, "", fmt.Errorf("%w: %s", errMissingParentID, parentIDKeyJobID)
		}

		u, err := urlbuilder.New(baseURL, "v1", "jobs", jobID, params.ObjectName)

		return u, http.MethodPost, err

	default:
		return nil, "", common.ErrOperationNotSupportedForObject
	}
}

// stringFromRecord pulls a parent-id from RecordData and removes it from the
// map so it is not echoed back into the request body. Returns false if the
// key is missing or empty.
func stringFromRecord(record map[string]any, key string) (string, bool) {
	raw, ok := record[key]
	if !ok {
		return "", false
	}

	str, ok := raw.(string)
	if !ok || str == "" {
		return "", false
	}

	delete(record, key)

	return str, true
}

// parseWriteResponse extracts the record's UUID (Gusto's primary key) and
// returns the full response body as Data. PUTs that return 200 with the
// updated object echo through the same path. On 204 (no body), the caller's
// RecordId is echoed for correlation.
func (c *Connector) parseWriteResponse(
	_ context.Context,
	params common.WriteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success:  true,
			RecordId: params.RecordId,
		}, nil
	}

	recordID, err := jsonquery.New(body).StrWithDefault("uuid", params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}
