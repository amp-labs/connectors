package acculynx

import (
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

type paginationStyle int

const (
	paginationOffsetRecord paginationStyle = iota
	paginationOffsetPage
	paginationPageNumber
	paginationNone
)

// objectReadSpec captures the per-object pagination style and incremental
// timeKey. timeKey is the response field used for connector-side Since/Until
// filtering; per repo convention only the "updated_at" semantic field qualifies
// (modifiedDate on AccuLynx) — never createdDate.
type objectReadSpec struct {
	pagination paginationStyle
	timeKey    string
}

//nolint:gochecknoglobals
var objectReadSpecs = datautils.NewDefaultMap(map[string]objectReadSpec{
	// jobs is the only endpoint with a provider-side ModifiedDate filter; we
	// still apply connector-side filtering on top to enforce time bounds precisely.
	"jobs":                           {pagination: paginationOffsetRecord, timeKey: "modifiedDate"},
	"jobs/custom-fields":             {pagination: paginationOffsetRecord, timeKey: "modifiedDate"},
	"contacts/custom-fields":         {pagination: paginationOffsetRecord, timeKey: "modifiedDate"},
	"estimates/sections":             {pagination: paginationNone, timeKey: "modifiedDate"},
	"company-settings/custom-fields": {pagination: paginationOffsetRecord, timeKey: "modifiedDate"},

	"calendars":             {pagination: paginationOffsetRecord},
	"users":                 {pagination: paginationOffsetRecord},
	"supplements":           {pagination: paginationOffsetRecord},
	"supplements/items":     {pagination: paginationOffsetRecord},
	"supplements/notations": {pagination: paginationOffsetRecord},
	"jobs/estimates":        {pagination: paginationOffsetRecord},
	"jobs/history":          {pagination: paginationOffsetRecord, timeKey: "date"},
	"jobs/representatives":  {pagination: paginationOffsetRecord},
	"company-settings/job-file-settings/document-folders":    {pagination: paginationOffsetRecord},
	"company-settings/job-file-settings/insurance-companies": {pagination: paginationOffsetRecord},
	"company-settings/job-file-settings/job-categories":      {pagination: paginationOffsetRecord},
	"company-settings/job-file-settings/photo-video-tags":    {pagination: paginationOffsetRecord},
	"company-settings/job-file-settings/trade-types":         {pagination: paginationOffsetRecord},
	"company-settings/job-file-settings/work-types":          {pagination: paginationOffsetRecord},
	"company-settings/leads/lead-sources":                    {pagination: paginationOffsetRecord},

	// pageStartIndex is the AccuLynx-side offset for these endpoints, advanced
	// by the per-page record count (see advancePagination). Verified via
	// metadata/queryParamStats.json — pageNumber is not an accepted param on
	// any /api/v2 list endpoint, so previously routing these through
	// paginationPageNumber caused the connector to loop on page 1 (AccuLynx
	// silently ignores pageNumber and returns the first page every time).
	"estimates":              {pagination: paginationOffsetPage},
	"calendars/appointments": {pagination: paginationOffsetPage},
	"contacts":               {pagination: paginationOffsetPage},
	"contacts/contact-types": {pagination: paginationOffsetPage},
	"jobs/invoices":          {pagination: paginationOffsetPage},

	"acculynx/countries":        {pagination: paginationNone},
	"acculynx/units-of-measure": {pagination: paginationNone},
	"contacts/email-addresses":  {pagination: paginationNone},
	"contacts/phone-numbers":    {pagination: paginationNone},
	"jobs/contacts":             {pagination: paginationNone},
	"jobs/milestone-history":    {pagination: paginationNone},
	"company-settings/job-file-settings/workflow-milestones": {pagination: paginationNone},
	"company-settings/location-settings/account-types":       {pagination: paginationNone},
}, func(string) objectReadSpec {
	return objectReadSpec{pagination: paginationOffsetRecord}
})

// makeFilterFunc returns an identity filter when the object exposes no usable
// timestamp or the caller supplied no time bounds; otherwise it returns a
// connector-side time filter using the object's modifiedDate field.
func (c *Connector) makeFilterFunc(params common.ReadParams, reqURL *urlbuilder.URL) common.RecordsFilterFunc {
	nextPage := c.makeNextPage(params.ObjectName, reqURL)

	spec := objectReadSpecs.Get(params.ObjectName)
	if spec.timeKey == "" {
		return readhelper.MakeIdentityFilterFunc(nextPage)
	}

	if params.Since.IsZero() && params.Until.IsZero() {
		return readhelper.MakeIdentityFilterFunc(nextPage)
	}

	return readhelper.MakeTimeFilterFunc(
		readhelper.Unordered,
		readhelper.NewTimeBoundary(),
		spec.timeKey,
		time.RFC3339,
		nextPage,
	)
}

// applyJobsIncrementalFilter adds the provider-side ModifiedDate filter to /jobs
// when Since or Until is set. AccuLynx accepts dates in YYYY-MM-DD format with
// day-level granularity; the connector-side filter narrows further at the
// timestamp level.
//
// Reference: https://apidocs.acculynx.com/reference/getjobs
func applyJobsIncrementalFilter(url *urlbuilder.URL, params common.ReadParams) {
	if params.ObjectName != objectJobs {
		return
	}

	if params.Since.IsZero() && params.Until.IsZero() {
		return
	}

	since, until := pairedDateWindow(params.Since, params.Until)
	url.WithQueryParam("dateFilterType", "ModifiedDate")
	url.WithQueryParam("startDate", since.Format(time.DateOnly))
	url.WithQueryParam("endDate", until.Format(time.DateOnly))
}

// applyHistoryDateWindow pushes Since/Until into AccuLynx's server-side
// startDate/endDate filter for /jobs/{id}/history. Without this, an unbounded
// read of a long-lived job's history can require tens of thousands of paged
// requests; with it, the server returns only records inside the requested
// window.
//
// AccuLynx requires YYYY-MM-DD format; passing time-of-day returns HTTP 400.
func applyHistoryDateWindow(url *urlbuilder.URL, params common.ReadParams) {
	if params.ObjectName != "jobs/history" {
		return
	}

	if params.Since.IsZero() && params.Until.IsZero() {
		return
	}

	since, until := pairedDateWindow(params.Since, params.Until)
	url.WithQueryParam("startDate", since.Format(time.DateOnly))
	url.WithQueryParam("endDate", until.Format(time.DateOnly))
}

// pairedDateWindow normalizes a Since/Until pair so both bounds are always
// present. AccuLynx returns HTTP 400 ("Start Date and End Date do not have
// the same format") on /jobs and /jobs/{id}/history when only one of
// startDate / endDate is sent. Missing bounds default to:
//   - Until → time.Now().UTC()
//   - Since → Unix epoch (predates AccuLynx; safely covers all history)
//
// The connector-side time filter still enforces precise bounds on the
// response, so widening upstream when a bound is missing is safe.
func pairedDateWindow(since, until time.Time) (time.Time, time.Time) {
	if until.IsZero() {
		until = time.Now().UTC()
	}

	if since.IsZero() {
		since = time.Unix(0, 0).UTC()
	}

	return since, until
}
