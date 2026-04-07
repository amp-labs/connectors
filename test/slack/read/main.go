package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/slack"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := slack.NewConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "conversations",
		Fields:     connectors.Fields("id", "name", "is_private", "is_archived", "num_members"),
	})

	if err != nil {
		slog.Error(err.Error())
		return
	}

	fmt.Println("Read conversation result...")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("id", "name", "real_name", "is_bot"),
	})

	if err != nil {
		slog.Error(err.Error())
		return
	}

	fmt.Println("Read user result...")
	utils.DumpJSON(res, os.Stdout)

}
