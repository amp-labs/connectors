package auphan

import (
	"errors"
	"fmt"
	"maps"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// RawFilterTypeReportFilters is the raw filter type report reads accept: an object of extra Auphan
// report filters, merged into the request's "filters". Valid ids per report come from reports/types.get,
// e.g. {"status": [2], "net_cash_deducts_tips": 0, "selection": "open_date"}.
// date_from and date_to are not allowed, because the connector sets them from Since and Until.
const RawFilterTypeReportFilters = "reportFilters"

var (
	errFilterNotObject   = errors.New("must be an object of Auphan report filters")
	errFilterDateRange   = errors.New("must not set date_from or date_to, use Since and Until")
	errFilterOnlyReports = errors.New("report filters only apply to report objects")
)

// ReportFilters is Auphan's raw filter: extra filters for reports/execute.
type ReportFilters map[string]any

// RawFilterType returns RawFilterTypeReportFilters.
func (ReportFilters) RawFilterType() string {
	return RawFilterTypeReportFilters
}

// FromFilter accepts a JSON object without date_from or date_to.
func (ReportFilters) FromFilter(filter any) (ReportFilters, error) {
	filters, ok := filter.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w, got %T", errFilterNotObject, filter)
	}

	if _, ok := filters["date_from"]; ok {
		return nil, errFilterDateRange
	}

	if _, ok := filters["date_to"]; ok {
		return nil, errFilterDateRange
	}

	return filters, nil
}

var _ connectors.RawFilterConnector = (*Connector)(nil)

// ValidateRawFilter accepts raw filters of type RawFilterTypeReportFilters.
func (*Connector) ValidateRawFilter(filter common.RawFilter) error {
	_, err := common.RawFilterAs[ReportFilters](filter)

	return err
}

// reportFilters builds a report's "filters": the date window, the report's date selection, and any raw filters.
func reportFilters(params common.ReadParams, window *reportWindow) (map[string]any, error) {
	filters := map[string]any{}

	if selection := reportObjects[params.ObjectName]; selection != "" {
		filters["selection"] = selection
	}

	if params.RawFilter != nil {
		extra, err := common.RawFilterAs[ReportFilters](*params.RawFilter)
		if err != nil {
			return nil, err
		}

		maps.Copy(filters, extra)
	}

	filters["date_from"] = window.from.Format(reportTimeLayout)
	filters["date_to"] = window.to.Format(reportTimeLayout)

	return filters, nil
}
