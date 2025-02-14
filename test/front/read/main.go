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
	fr "github.com/amp-labs/connectors/providers/front"
	"github.com/amp-labs/connectors/test/front"

	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := front.GetFrontConnector(ctx)

	if err := testReadTeams(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadAccounts(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadContacts(ctx, conn); err != nil {
		slog.Info(err.Error())
	}
}

func testReadTeams(ctx context.Context, conn *fr.Connector) error {
	params := common.ReadParams{
		ObjectName: "teams",
		Fields:     connectors.Fields("id", "name"),
		// NextPage:   "https://fire.chilipiper.com/api/fire-edge/v1/org/workspace?page=1\u0026pageSize=2",
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

func testReadAccounts(ctx context.Context, conn *fr.Connector) error {
	params := common.ReadParams{
		ObjectName: "accounts",
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

func testReadContacts(ctx context.Context, conn *fr.Connector) error {
	params := common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("name", "id"),
		Since:      time.Now().Add(-1000 * time.Hour),
		// NextPage:   "https://fire.chilipiper.com/api/fire-edge/v1/org/workspace/users?page=1\u0026pageSize=2",
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
