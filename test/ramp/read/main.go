package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/ramp"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetRampConnector(ctx)

	sinceTime := time.Now().Add(-1 * time.Hour) // last 30 days

	slog.Info("Reading transactions (last hour)...")

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "transactions",
		Fields: connectors.Fields(
			"id", "amount", "merchant_name", "state", "sync_status", "user_transaction_time",
		),
		Since: sinceTime,
	})
	if err != nil {
		utils.Fail("error reading transactions", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading users...")

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("id", "email", "first_name", "last_name", "role", "status"),
	})
	if err != nil {
		utils.Fail("error reading users", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading departments...")

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "departments",
		Fields:     connectors.Fields("id", "name"),
	})
	if err != nil {
		utils.Fail("error reading departments", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading vendors (last hour)...")

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "vendors",
		Fields:     connectors.Fields("id", "name", "is_active", "total_spend_all_time"),
		Since:      sinceTime,
	})
	if err != nil {
		utils.Fail("error reading vendors", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading reimbursements (last hour)...")

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "reimbursements",
		Fields:     connectors.Fields("id", "amount", "state", "user_id", "sync_status"),
		Since:      sinceTime,
	})
	if err != nil {
		utils.Fail("error reading reimbursements", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
}
