package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	housecallpro "github.com/amp-labs/connectors/test/housecallPro"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := housecallpro.GetConnector(ctx)

	const pageSize = 10

	// Object names match providers/housecallPro/metadata/schemas.json.
	for _, tc := range []struct {
		object string
		fields []string
	}{
		// {"customers", []string{"id", "first_name", "email"}},
		// {"employees", []string{"id", "first_name", "email"}},
		// {"estimates", []string{"id", "estimate_number", "updated_at"}},
		// {"jobs", []string{"id", "updated_at", "work_status"}},
		// {"job_fields/job_types", []string{"id", "name"}},
		// {"leads", []string{"id", "status"}},
		// {"lead_sources", []string{"id", "name"}},
		// {"price_book/material_categories", []string{"uuid", "name", "updated_at"}},
		// {"price_book/price_forms", []string{"id", "name"}},
		{"price_book/services", []string{"uuid", "name"}},
		// {"service_zones", []string{"id", "name"}},
		// {"events", []string{"id", "name", "updated_at"}},
		// {"tags", []string{"id", "name"}},
		// {"invoices", []string{"id", "invoice_number", "status"}},
		// {"routes", []string{"id", "name"}},
	} {
		slog.Info("reading object", "object", tc.object)
		testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
			ObjectName: tc.object,
			Fields:     connectors.Fields(tc.fields...),
			PageSize:   pageSize,
		})
	}
}
