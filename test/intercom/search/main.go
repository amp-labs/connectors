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
	msTest "github.com/amp-labs/connectors/test/intercom"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := msTest.GetIntercomConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "conversations",
		Fields: connectors.Fields(
			"id",
			"state",
			"type",
		),
		Since: time.Unix(1726674883, 0),
	})
	if err != nil {
		utils.Fail("error reading from Intercom", "error", err)
	}

	fmt.Println("Reading conversations..")
	utils.DumpJSON(res, os.Stdout)

	if len(res.NextPage) == 0 {
		return
	}

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "conversations",
		Fields: connectors.Fields(
			"id",
			"state",
			"type",
		),
		NextPage: res.NextPage,
		Since:    time.Unix(1726674883, 0),
	})
	if err != nil {
		utils.Fail("error reading from Intercom", "error", err)
	}

	fmt.Println("Reading conversations (SecondPage)..")
	utils.DumpJSON(res, os.Stdout)
}
