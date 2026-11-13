package auphan

import (
	"maps"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
)

// reportPayload describes what goes into the "filters" of a report's request body.
// Every Auphan report is read with the same call, only the filters differ:
//
//	POST reports/execute
//	{"type": "payments", "columns": [...], "filters": {"selection": "close_date",
//	  "date_from": "2026-11-05 00:00:00", "date_to": "2026-11-05 23:59:59"}}
type reportPayload struct {
	// filters are always sent for this report. They are the filters Auphan insists on
	// but won't fill in by itself, such as "selection" (which date field the report filters by).
	// Leave a filter out if the report doesn't have it: Auphan rejects unknown filters.
	filters map[string]any

	// dateLayout is how date_from and date_to are written, e.g. "2026-11-05 00:00:00" or "2026-11-05".
	// Leave it empty for a report with no date range: it is then read whole, in one call.
	dateLayout string
}

// selection is the payload for a report that picks its date field with "selection".
func selection(field string) reportPayload {
	return reportPayload{filters: map[string]any{"selection": field}, dateLayout: time.DateTime}
}

// with adds filters to a payload.
func (p reportPayload) with(filters map[string]any) reportPayload {
	merged := maps.Clone(p.filters)
	maps.Copy(merged, filters)
	p.filters = merged

	return p
}

// reportObjects lists every object read through the reports endpoint (reports/execute).
// The key is both the object name and the Auphan report type.
var reportObjects = datautils.Map[string, reportPayload]{ //nolint:gochecknoglobals
	"adjustments":     {dateLayout: time.DateTime},
	"attendance":      {filters: map[string]any{"status": []int{4}}, dateLayout: time.DateTime},
	"bank_drop":       {dateLayout: time.DateTime},
	"costs":           {dateLayout: time.DateTime},
	"counts":          {dateLayout: time.DateTime},
	"customer_tabs":   {dateLayout: time.DateTime},
	"customers":       {dateLayout: time.DateTime},
	"employees":       {dateLayout: time.DateTime},
	"gfs_sku":         {},
	"invoice_logs":    {dateLayout: time.DateTime},
	"invoices":        selection("close_date"),
	"message_actions": {dateLayout: time.DateTime},
	"modifiers":       selection("close_date"),
	"payment_txns":    {dateLayout: time.DateTime},
	// payments and sales send defaults that Auphan lists in reports/types.get but does not apply
	// when they are left out (checked live). Without them, payments counts the change given back
	// on cash, and sales counts deleted and moved items and gift cards sold.
	"payments": selection("close_date").with(map[string]any{
		"status": []int{2}, "remove_cash_change": 1, "net_cash_deducts_tips": 1,
	}),
	"petty_cash":          {dateLayout: time.DateTime},
	"price_change":        {dateLayout: time.DateTime},
	"production":          {dateLayout: time.DateTime},
	"purchases":           selection("purchase_date"),
	"receiving":           {dateLayout: time.DateTime},
	"sales":               selection("close_date").with(map[string]any{"status": []int{1}, "item_type": []int{1}}),
	"sales_printer_stats": {dateLayout: time.DateTime},
	"salerules":           {dateLayout: time.DateOnly},
	"screen_stats":        selection("start_time"),
	"shifts":              {dateLayout: time.DateTime},
	"stock_usage":         {dateLayout: time.DateTime},
	"tax":                 selection("close_date"),
	"transfers":           {dateLayout: time.DateTime},
	"usergroups":          {},
	"waitlist":            selection("reservation_dt"),
	"worked_hours":        selection("sales_date"),
}

// listObjects are read from a paginated list endpoint.
var listObjects = map[string]string{ //nolint:gochecknoglobals
	"stores": "settings/stores.list",
}
