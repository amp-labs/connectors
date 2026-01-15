package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	cl "github.com/amp-labs/connectors/providers/callrail"
	"github.com/amp-labs/connectors/test/callrail"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := callrail.NewConnector(ctx)
	conn.GetPostAuthInfo(ctx)

	if err := testReadCompanies(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadUsers(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadTrackers(ctx, conn); err != nil {
		slog.Info(err.Error())
	}
}

func testReadCompanies(ctx context.Context, conn *cl.Connector) error {
	params := common.ReadParams{
		ObjectName: "companies",
		Fields:     connectors.Fields("id", "name", "created_at"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadUsers(ctx context.Context, conn *cl.Connector) error {
	params := common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("email", "created_at", "role"),
		Since:      time.Now().Add(-10 * time.Hour),
		NextPage:   "2",
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadTrackers(ctx context.Context, conn *cl.Connector) error {
	params := common.ReadParams{
		ObjectName: "trackers",
		Fields:     connectors.Fields("name", "status", "type"),
		NextPage:   "",
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
