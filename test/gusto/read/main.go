package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/gusto"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)

	slog.Info("=== Reading employees ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "employees",
		Fields:     connectors.Fields("uuid", "first_name", "last_name", "email"),
	})

	slog.Info("=== Reading locations ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "locations",
		Fields:     connectors.Fields("uuid", "street_1", "city", "state"),
	})

	slog.Info("=== Reading departments ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "departments",
		Fields:     connectors.Fields("uuid", "title"),
	})

	slog.Info("=== Reading pay_schedules ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "pay_schedules",
		Fields:     connectors.Fields("uuid", "frequency"),
	})

	slog.Info("=== Reading contractors ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "contractors",
		Fields:     connectors.Fields("uuid", "first_name", "last_name"),
	})

	slog.Info("=== Reading jobs (employee-scoped) ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "jobs",
		Fields:     connectors.Fields("uuid", "title", "employee_uuid", "rate", "payment_unit"),
	})

	slog.Info("=== Reading garnishments (employee-scoped) ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "garnishments",
		Fields:     connectors.Fields("uuid", "employee_uuid", "description", "amount"),
	})

	slog.Info("=== Reading employee_benefits (employee-scoped) ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "employee_benefits",
		Fields:     connectors.Fields("uuid", "employee_uuid", "company_benefit_uuid", "active"),
	})

	slog.Info("=== Reading home_addresses (employee-scoped) ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "home_addresses",
		Fields:     connectors.Fields("uuid", "employee_uuid", "city", "state"),
	})
}
