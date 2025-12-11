package sellsy

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

// createReadOperation determines the HTTP method and request payload for reading an object.
// For search endpoints, it includes time-based filters when supported.
// For other endpoints, it defaults to a GET without a payload. These endpoints have no filtering options.
//
// Behavior:
//   - Uses POST with a JSON payload for endpoints ending in `/search`.
//   - Uses GET without a payload for all other endpoints.
//
// Background:
//
//	Incremental reads in Sellsy are inconsistent across endpoints. Each
//	"search" resource may support different filter fields (e.g., "created",
//	"updated", "updated_at"), or expose different timestamp fields in
//	their responses. Some objects do not support timestamp filtering at all.
//
// Common cases:
//   - Some objects can be filtered by "updated" and include that field in response.
//   - Some can only be filtered by "created" but return "updated" (requiring client-side filtering).
//   - Others expose no filterable timestamp field, though they may include "created" in responses.
//   - A few lack both filterable parameters and usable timestamps entirely.
func createReadOperation(
	url *urlbuilder.URL, params common.ReadParams,
) (method string, payload []byte, err error) {
	if !strings.HasSuffix(url.Path(), "/search") {
		return http.MethodGet, nil, nil
	}

	filter := readFilter{}
	if !params.Since.IsZero() {
		filter.Start = datautils.Time.FormatRFC3339inUTC(params.Since)
	}

	if !params.Until.IsZero() {
		filter.End = datautils.Time.FormatRFC3339inUTC(params.Until)
	}

	objectTimeSpec := objectsFilterParam.Get(params.ObjectName)

	searchPayload := &readSearchPayload{
		Filters: newReadSearchFilters(objectTimeSpec.RequestBy, filter),
	}

	payload, err = json.Marshal(searchPayload)
	if err != nil {
		return "", nil, err
	}

	return http.MethodPost, payload, nil
}

// makeFilterFunc returns a filtering function that respects the Since and Until parameters specified in ReadParams.
//
// If the object does not expose a suitable timestamp field for filtering,
// the function returns an identity filter (passing all records unchanged).
//
// Note:
//
//	Even if the HTTP method (GET or POST) does not support server-side time filtering,
//	the client still applies local filtering to exclude records outside the specified time window.
func makeFilterFunc(params common.ReadParams, request *http.Request) common.RecordsFilterFunc {
	timeSpec := objectsFilterParam.Get(params.ObjectName)
	if timeSpec.ResponseAt == "" {
		// No timestamp field available; return all records unfiltered.
		return readhelper.MakeIdentityFilterFunc(makeNextRecordsURL(request.URL))
	}

	// Client side filtering.
	return readhelper.MakeTimeFilterFunc(
		readhelper.ChronologicalOrder,
		readhelper.NewTimeBoundary(),
		timeSpec.ResponseAt, time.RFC3339,
		makeNextRecordsURL(request.URL))
}

// objectsFilterParam defines which timestamp fields and filters apply to
// each Sellsy object for incremental reads.
//
// Structure:
//   - RequestBy: the time filter parameter accepted by the Sellsy API (if any).
//   - ResponseAt: the timestamp field to use when advancing the incremental read.
//
// Notes:
//   - Some objects can be queried by "created" but should advance based on "updated".
//   - Some support no query filters but still expose a timestamp usable for client-side filtering.
//   - If both are unavailable, RequestBy = filterNone and ResponseAt = "".
//
// References:
//
//	Each object entry includes a link to its API documentation for clarity.
var objectsFilterParam = datautils.NewDefaultMap(map[string]timeFieldSpec{ // nolint:gochecknoglobals
	// -----------------------------------------------------------------------------
	// (1) Objects with UPDATED field
	// -----------------------------------------------------------------------------
	// comments: https://docs.sellsy.com/api/v2/#operation/search-comments
	"comments": {RequestBy: filterByUpdated, ResponseAt: "updated"},
	// contacts: https://docs.sellsy.com/api/v2/#operation/search-contacts
	"contacts": {RequestBy: filterByUpdated, ResponseAt: "updated"},
	// phone-calls: https://docs.sellsy.com/api/v2/#operation/search-phone-calls
	"phone-calls": {RequestBy: filterByCreated, ResponseAt: "updated"},
	// tasks: https://docs.sellsy.com/api/v2/#operation/search-tasks
	"tasks": {RequestBy: filterByCreated, ResponseAt: "updated"},
	// calendar-events: https://docs.sellsy.com/api/v2/#operation/search-calendar-events
	"calendar-events": {RequestBy: filterByDate, ResponseAt: "updated"},
	// custom-activities: https://docs.sellsy.com/api/v2/#operation/post-custom-activities-search
	"custom-activities": {RequestBy: filterByDate, ResponseAt: "updated"},
	// custom-activity-types: https://docs.sellsy.com/api/v2/#operation/get-custom-activity-types
	"custom-activity-types": {RequestBy: filterNone, ResponseAt: "updated"},
	// webhooks: https://docs.sellsy.com/api/v2/#operation/search-webhooks
	"webhooks": {RequestBy: filterNone, ResponseAt: "updated"},
	// -----------------------------------------------------------------------------
	// (2) Objects with CREATED field
	// -----------------------------------------------------------------------------
	// deposit-invoices: https://docs.sellsy.com/api/v2/#operation/search-deposit-invoices
	"deposit-invoices": {RequestBy: filterByCreated, ResponseAt: "created"},
	// documents/models: https://docs.sellsy.com/api/v2/#operation/search-models
	"documents/models": {RequestBy: filterByCreated, ResponseAt: "created"},
	// estimates: https://docs.sellsy.com/api/v2/#operation/search-estimates
	"estimates": {RequestBy: filterByCreated, ResponseAt: "created"},
	// invoices: https://docs.sellsy.com/api/v2/#operation/search-invoices
	"invoices": {RequestBy: filterByCreated, ResponseAt: "created"},
	// opportunities: https://docs.sellsy.com/api/v2/#operation/search-opportunities
	"opportunities": {RequestBy: filterByCreated, ResponseAt: "created"},
	// credit-notes: https://docs.sellsy.com/api/v2/#operation/search-credit-notes
	"credit-notes": {RequestBy: filterNone, ResponseAt: "created"},
	// notifications: https://docs.sellsy.com/api/v2/#operation/search-notifications
	"notifications": {RequestBy: filterNone, ResponseAt: "created"},
	// orders: https://docs.sellsy.com/api/v2/#operation/search-orders
	"orders": {RequestBy: filterNone, ResponseAt: "created"},
	// staffs: https://docs.sellsy.com/api/v2/#operation/search-staffs
	"staffs": {RequestBy: filterNone, ResponseAt: "created"},
	// subscriptions: https://docs.sellsy.com/api/v2/#operation/search-subscriptions
	"subscriptions": {RequestBy: filterNone, ResponseAt: "created"},
	// -----------------------------------------------------------------------------
	// (3) Objects with UPDATED_AT field
	// -----------------------------------------------------------------------------
	// companies: https://docs.sellsy.com/api/v2/#operation/search-companies
	"companies": {RequestBy: filterByUpdatedAt, ResponseAt: "updated_at"},
	// individuals: https://docs.sellsy.com/api/v2/#operation/search-individuals
	"individuals": {RequestBy: filterByUpdatedAt, ResponseAt: "updated_at"},
	// -----------------------------------------------------------------------------
	// (4) Objects with CREATED_AT field:
	// -----------------------------------------------------------------------------
	// pur-invoice: https://docs.sellsy.com/api/v2/#operation/search-ocr-pur-invoice
	"pur-invoice": {RequestBy: filterNone, ResponseAt: "created_at"},
}, func(objectName string) timeFieldSpec {
	// Default case: no time-based filter or response field available.
	return timeFieldSpec{
		RequestBy:  filterNone,
		ResponseAt: "",
	}
})

// timeFieldSpec defines which Sellsy API parameters and response fields are
// associated with time-based filtering for incremental reads.
//
// Example: request records created after a certain time, but advance incrementally based on their "updated" timestamp.
//
//	timeFieldSpec{
//	    RequestBy:  filterByCreated,
//	    ResponseAt: "updated",
//	}
type timeFieldSpec struct {
	// RequestBy specifies the query parameter accepted by the API (if any).
	RequestBy filterParameterType
	// ResponseAt specifies the timestamp field available in the response used
	// to advance the incremental cursor.
	ResponseAt string
}

// readSearchPayload represents the JSON payload for a Sellsy search request.
type readSearchPayload struct {
	Filters readSearchFilters `json:"filters"`
}

// readSearchFilters defines the supported time-based filters accepted in Sellsy
// search request payloads.
//
// Note: These filters control query parameters only — they are not guaranteed to
// directly correspond to timestamp fields in the response. For example, an endpoint
// might accept a `date` filter but return `created` or `updated` fields instead.
//
// Sellsy's API is inconsistent in how it names and expects these filters across
// different objects. Using an unsupported parameter results in a 400 Bad Request.
type readSearchFilters struct {
	// Date is used by endpoints that filter results by a general "date" range.
	// Despite the name, this parameter effectively filters by creation time.
	// Some endpoints expose both `created` and `updated` timestamps in responses,
	// but still expect `date` as the filter parameter.
	Date *readFilter `json:"date,omitempty"`

	// Created filters results by record creation time.
	// Common for resources that explicitly expose `created` in their payloads
	// but do not support an `updated` timestamp.
	// However, it doesn't mean response won't have `updated` field!
	Created *readFilter `json:"created,omitempty"`

	// Updated filters results using the `updated` timestamp field.
	// This parameter is supported by only a few endpoints.
	Updated *readFilter `json:"updated,omitempty"`

	// UpdatedAt exists because some endpoints expect `updated_at` instead of `updated`
	// as their time filter parameter. Functionally identical to `updated`, it reflects
	// Sellsy’s inconsistent naming conventions across objects.
	UpdatedAt *readFilter `json:"updated_at,omitempty"`
}

type readFilter struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

// newReadSearchFilters constructs a readSearchFilters struct with the appropriate time filter based on filter type.
func newReadSearchFilters(parameter filterParameterType, filter readFilter) readSearchFilters {
	filters := readSearchFilters{}

	switch parameter {
	case filterByDate:
		filters.Date = &filter
	case filterByCreated:
		filters.Created = &filter
	case filterByUpdated:
		filters.Updated = &filter
	case filterByUpdatedAt:
		filters.UpdatedAt = &filter
	case filterNone:
		fallthrough
	default:
		return filters // no-op, empty filters.
	}

	return filters
}

type filterParameterType int

const (
	filterNone filterParameterType = iota
	filterByDate
	filterByCreated
	filterByUpdated
	filterByUpdatedAt
)
