package auphan

import (
	"maps"
	"time"

	"github.com/amp-labs/connectors/common"
)

// earliestReportDate is where a read without Since starts. Every dated report was read live from
// this date to today in one call. It is a wall-clock date, so it is not converted to the store's zone.
var earliestReportDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) //nolint:gochecknoglobals

// reportFilters builds the "filters" object of a report request: the report's always-sent filters,
// then date_from and date_to.
func (c *Connector) reportFilters(params common.ReadParams, report reportPayload) map[string]any {
	filters := map[string]any{}
	maps.Copy(filters, report.filters)

	// A report without a date range takes no dates at all.
	if report.dateLayout == "" {
		return filters
	}

	// Auphan reads a missing date_from as the start of today, so without Since the read starts at
	// earliestReportDate to get the whole history. Without Until, Auphan reads through today.
	filters["date_from"] = earliestReportDate.Format(report.dateLayout)
	if !params.Since.IsZero() {
		filters["date_from"] = c.storeTime(params.Since).Format(report.dateLayout)
	}

	if !params.Until.IsZero() {
		filters["date_to"] = c.storeTime(params.Until).Format(report.dateLayout)
	}

	return filters
}
