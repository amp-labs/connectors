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
	cp "github.com/amp-labs/connectors/providers/chilipiper"
	"github.com/amp-labs/connectors/test/chilipiper"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := chilipiper.GetChiliPiperConnector(ctx)

	if err := testReadTeams(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadWorkspaces(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadWorkspaceUsers(ctx, conn); err != nil {
		slog.Info(err.Error())
	}
}

func testReadWorkspaces(ctx context.Context, conn *cp.Connector) error {
	params := common.ReadParams{
		ObjectName: "workspace",
		Fields:     connectors.Fields("id", "name"),
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

func testReadTeams(ctx context.Context, conn *cp.Connector) error {
	params := common.ReadParams{
		ObjectName: "team",
		Fields:     connectors.Fields("id", "name"),
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

func testReadWorkspaceUsers(ctx context.Context, conn *cp.Connector) error {
	params := common.ReadParams{
		ObjectName: "workspace_users",
		Fields:     connectors.Fields("name", "id"),
		Since:      time.Now().Add(-1000 * time.Hour),
		// NextPage:   "https://fire.chilipiper.com/api/fire-edge/v1/org/workspace?page=2&pageSize=2",
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
