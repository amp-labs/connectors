package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoom"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()
	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZoomConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("email"),
	})
	if err != nil {
		utils.Fail("error reading from Zoom", "error", err)
	}

	fmt.Println("Reading users..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "devices",
		Fields:     connectors.Fields("app_version"),
	})
	if err != nil {
		utils.Fail("error reading from Zoom", "error", err)
	}

	fmt.Println("Reading devices..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "recordings",
		Fields:     connectors.Fields("account_id", "duration", "start_time"),
	})
	if err != nil {
		utils.Fail("error reading from Zoom", "error", err)
	}

	fmt.Println("Reading recordings..")
	utils.DumpJSON(res, os.Stdout)

	// Incremental read: recordings from the last 30 days
	since := time.Now().AddDate(0, 0, -30)
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "recordings",
		Fields:     connectors.Fields("id", "topic", "start_time"),
		Since:      since,
	})
	if err != nil {
		utils.Fail("error reading from Zoom", "error", err)
	}

	fmt.Println("Reading recordings (incremental - last 30 days)..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "users_report",
		Fields:     connectors.Fields("id", "email", "user_name"),
		Since:      since,
	})
	if err != nil {
		utils.Fail("error reading from Zoom", "error", err)
	}

	fmt.Println("Reading users report..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "operation_logs_report",
		Fields:     connectors.Fields("action", "operation_detail", "operator"),
		Since:      since,
	})
	if err != nil {
		utils.Fail("error reading from Zoom", "error", err)
	}

	fmt.Println("Reading operation logs report..")
	utils.DumpJSON(res, os.Stdout)

}
