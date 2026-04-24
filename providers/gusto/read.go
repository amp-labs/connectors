package gusto

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/amp-labs/connectors/providers/gusto/metadata"
	"github.com/spyzhov/ajson"
)

const (
	// Pagination:
	//   - Gusto default per-page is 25 (per docs), but no hard maximum is documented.
	//   - 100 works reliably per live tests and is what we use as both default and cap.
	//   - We cap because our "records < per" end-of-pages check would mis-detect the
	//     last page if Gusto silently caps server-side below the requested value.
	defaultPageSize      = "100"
	maxPageSize          = 100
	pageParam            = "page"
	perParam             = "per"
	companyIDPlaceholder = "{company_id}"

	// Gusto rate limit: 200 req/min. 4 concurrent child requests per employee
	// page is conservative and well within that limit.
	maxConcurrentEmployeeFetch = 4
)

// Object name constants — all objects readable via this connector.
const (
	objectAdmins             = "admins"
	objectCompanies          = "companies"
	objectCompanyBenefits    = "company_benefits"
	objectContractorPayments = "contractor_payments"
	objectContractors        = "contractors"
	objectCustomFields       = "custom_fields"
	objectDepartments        = "departments"
	objectEarningTypes       = "earning_types"
	objectEmployees          = "employees"
	objectLocations          = "locations"
	objectPayPeriods         = "pay_periods"
	objectPaySchedules       = "pay_schedules"
	objectPayrolls           = "payrolls"
	// Employee-scoped: reading these requires fetching all employees first,
	// then fanning out one request per employee to the child endpoint.
	objectEmployeeBenefits  = "employee_benefits"
	objectGarnishments      = "garnishments"
	objectHomeAddresses     = "home_addresses"
	objectJobs              = "jobs"
	objectTimeOffActivities = "time_off_activities"
	objectWorkAddresses     = "work_addresses"
)

// ErrMissingCompanyID is returned when the connector is constructed without the companyId metadata.
var ErrMissingCompanyID = errors.New("gusto: companyId metadata is required")

// employeeScopedObjects are objects whose URL paths contain {employee_id}.
// Reading them requires first listing all employees in the company, then
// fanning out one request per employee to the child endpoint. The object
// name itself is used as the URL segment appended after /v1/employees/{uuid}/.
var employeeScopedObjects = datautils.NewSet( //nolint:gochecknoglobals
	objectEmployeeBenefits,
	objectGarnishments,
	objectHomeAddresses,
	objectJobs,
	objectTimeOffActivities,
	objectWorkAddresses,
)

// supportedReadObjects is the complete set of objects this connector can read.
// compensations is excluded: it requires /jobs/{job_id}/compensations, a two-level
// parent lookup (employees → jobs → compensations) with no precedent in this repo.
var supportedReadObjects = datautils.NewSet( //nolint:gochecknoglobals
	objectAdmins,
	objectCompanies,
	objectCompanyBenefits,
	objectContractorPayments,
	objectContractors,
	objectCustomFields,
	objectDepartments,
	objectEarningTypes,
	objectEmployees,
	objectLocations,
	objectPayPeriods,
	objectPaySchedules,
	objectPayrolls,
	objectEmployeeBenefits,
	objectGarnishments,
	objectHomeAddresses,
	objectJobs,
	objectTimeOffActivities,
	objectWorkAddresses,
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedReadObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	apiURL, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	if employeeScopedObjects.Has(params.ObjectName) {
		// Employee-scoped objects start by listing all employees.
		// parseReadResponse fans out to the child endpoint per employee UUID.
		return c.buildEmployeeListURL(params)
	}

	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	if strings.Contains(path, companyIDPlaceholder) {
		if c.companyID == "" {
			return nil, ErrMissingCompanyID
		}

		path = strings.ReplaceAll(path, companyIDPlaceholder, c.companyID)
	}

	apiURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	apiURL.WithQueryParam(perParam, pageSize(params))
	apiURL.WithQueryParam(pageParam, "1")

	return apiURL, nil
}

// buildEmployeeListURL returns the paginated employees-list URL, reusing the
// employees path from the schema so URL construction is consistent.
func (c *Connector) buildEmployeeListURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if c.companyID == "" {
		return nil, ErrMissingCompanyID
	}

	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), objectEmployees)
	if err != nil {
		return nil, err
	}

	path = strings.ReplaceAll(path, companyIDPlaceholder, c.companyID)

	apiURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	apiURL.WithQueryParam(perParam, pageSize(params))
	apiURL.WithQueryParam(pageParam, "1")

	return apiURL, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	if employeeScopedObjects.Has(params.ObjectName) {
		return c.parseEmployeeScopedResponse(ctx, params, request.URL, resp)
	}

	return common.ParseResultFiltered(
		params,
		resp,
		c.recordsFunc(params.ObjectName),
		readhelper.MakeIdentityFilterFunc(nextPageFromPageCounter(request.URL)),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}

// pageSize returns the page size to request, respecting the user's
// params.PageSize but capping at maxPageSize. Capping protects the
// completion check in nextPageFromPageCounter: if Gusto silently
// caps the page size server-side below what we requested, the
// returned record count would be less than per and we would
// incorrectly stop paginating.
func pageSize(params common.ReadParams) string {
	if params.PageSize <= 0 || params.PageSize > maxPageSize {
		return strconv.Itoa(maxPageSize)
	}

	return strconv.Itoa(params.PageSize)
}

// recordsFunc returns the records extractor for the given object.
// All Gusto list endpoints return a bare JSON array at the root (no
// wrapper object), so schemas.json uses "responseKey": "" for every
// object. common.MakeRecordsFunc("") handles this via jsonquery's
// SelfReference semantics — the empty key means "the node itself".
func (c *Connector) recordsFunc(objectName string) common.NodeRecordsFunc {
	return common.MakeRecordsFunc(
		metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), objectName),
	)
}

// parseEmployeeScopedResponse handles reads for objects whose paths require
// {employee_id}. It extracts employee UUIDs from the employee-list response,
// fans out one request per employee to the child endpoint, and returns the
// flattened records. Pagination advances through the employee list.
func (c *Connector) parseEmployeeScopedResponse(
	ctx context.Context,
	params common.ReadParams,
	reqURL *url.URL,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	body, ok := resp.Body()
	if !ok {
		return &common.ReadResult{Done: true}, nil
	}

	nodes, err := c.recordsFunc(objectEmployees)(body)
	if err != nil {
		return nil, err
	}

	uuids := make([]string, 0, len(nodes))

	for _, node := range nodes {
		uuid, uuidErr := jsonquery.New(node).StringRequired("uuid")
		if uuidErr != nil {
			continue
		}

		uuids = append(uuids, uuid)
	}

	childSuffix := params.ObjectName
	allRecords := make([][]common.ReadResultRow, len(uuids))

	jobs := make([]simultaneously.Job, len(uuids))
	for i, empUUID := range uuids {
		idx, uuid := i, empUUID

		jobs[idx] = func(ctx context.Context) error {
			rows, fetchErr := c.fetchEmployeeChildren(ctx, uuid, childSuffix, params)
			if fetchErr != nil {
				return fmt.Errorf("fetching %s for employee %s: %w", childSuffix, uuid, fetchErr)
			}

			allRecords[idx] = rows

			return nil
		}
	}

	if err = simultaneously.DoCtx(ctx, maxConcurrentEmployeeFetch, jobs...); err != nil {
		return nil, err
	}

	var data []common.ReadResultRow
	for _, rows := range allRecords {
		data = append(data, rows...)
	}

	nextPage, err := nextPageFromPageCounter(reqURL)(body)
	if err != nil {
		return nil, err
	}

	return &common.ReadResult{
		Rows:     int64(len(data)),
		Data:     data,
		NextPage: common.NextPageToken(nextPage),
		Done:     nextPage == "",
	}, nil
}

// fetchEmployeeChildren calls /v1/employees/{uuid}/{childSuffix} and returns all
// records as ReadResultRows. No child-level pagination is applied — a single
// request with per=100 matches the established pattern for nested reads in this
// repo (Pipedrive, Granola, NetSuite, Gmail).
func (c *Connector) fetchEmployeeChildren(
	ctx context.Context,
	employeeUUID, childSuffix string,
	params common.ReadParams,
) ([]common.ReadResultRow, error) {
	apiURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, "v1", "employees", employeeUUID, childSuffix)
	if err != nil {
		return nil, err
	}

	apiURL.WithQueryParam(perParam, defaultPageSize)
	apiURL.WithQueryParam(pageParam, "1")

	resp, err := c.JSONHTTPClient().Get(ctx, apiURL.String())
	if err != nil {
		return nil, err
	}

	result, err := common.ParseResultFiltered(
		params, resp,
		c.recordsFunc(params.ObjectName),
		readhelper.MakeIdentityFilterFunc(nextPageFromPageCounter(nil)),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// nextPageFromPageCounter increments the page query param when the current page
// was full (records >= per). Returns "" when the page is partial (no more data).
func nextPageFromPageCounter(previousRequestURL *url.URL) common.NextPageFunc {
	return func(root *ajson.Node) (string, error) {
		if previousRequestURL == nil || root == nil || !root.IsArray() {
			return "", nil
		}

		records, err := root.GetArray()
		if err != nil {
			return "", err
		}

		per, err := strconv.Atoi(previousRequestURL.Query().Get(perParam))
		if err != nil || per <= 0 {
			per, _ = strconv.Atoi(defaultPageSize)
		}

		if len(records) < per {
			return "", nil
		}

		currentPage, err := strconv.Atoi(previousRequestURL.Query().Get(pageParam))
		if err != nil || currentPage <= 0 {
			currentPage = 1
		}

		cloned, err := url.Parse(previousRequestURL.String())
		if err != nil {
			return "", err
		}

		next, err := urlbuilder.FromRawURL(cloned)
		if err != nil {
			return "", err
		}

		next.WithQueryParam(pageParam, strconv.Itoa(currentPage+1))

		return next.String(), nil
	}
}
