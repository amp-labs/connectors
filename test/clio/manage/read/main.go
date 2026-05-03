package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	clioTest "github.com/amp-labs/connectors/test/clio"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := clioTest.GetClioManageConnector(ctx)

	// all fields selected
	testscenario.ReadThroughPages(ctx, conn.Manage, common.ReadParams{
		ObjectName: "expense_categories",
		Fields: connectors.Fields(
			"id",
			"etag",
			"name",
			"rate",
			"entry_type",
			"created_at",
			"updated_at",
			"xero_expense_code",
			"accessible_to_user",
			"currency",
			"tax_setting",
			"tax_settings",
			"groups",
			"utbms_code",
		),
		PageSize: 1,
	})

	// sample fields
	testscenario.ReadThroughPages(ctx, conn.Manage, common.ReadParams{
		ObjectName: "activities",
		Fields:     connectors.Fields("id", "updated_at", "type", "activity_description", "user"),
		Since:      time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
	})

}
