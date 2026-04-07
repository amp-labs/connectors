package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/test/freshchat"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := freshchat.NewConnector(ctx)

	sinceTime, err := time.Parse(time.RFC3339, "2025-07-11T21:56:07.503Z")
	if err != nil {
		utils.Fail("parse time: %w", err)
	}

	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "users",
		Since:      sinceTime,
		Fields:     connectors.Fields("first_name", "id", "last_name", "email", "reference_id"),
	})
	if err != nil {
		utils.Fail("error reading users", "error", err)
	}

	fmt.Println("Reading users")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, connectors.ReadParams{
		ObjectName: "agents",
		Since:      sinceTime,
		Fields:     connectors.Fields("groups", "role_id", "timezone", "availability_status"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading.. agents")
	utils.DumpJSON(res, os.Stdout)
}
