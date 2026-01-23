package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	fmt.Println("Reading meetings..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "meeting_summaries",
		Fields:     connectors.Fields("topic"),
	})
	if err != nil {
		utils.Fail("error reading from Zoom", "error", err)
	}

	fmt.Println("Reading meetings..")
	utils.DumpJSON(res, os.Stdout)

}
