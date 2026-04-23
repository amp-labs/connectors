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
	co "github.com/amp-labs/connectors/providers/chargeover"
	"github.com/amp-labs/connectors/test/chargeover"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := chargeover.NewConnector(ctx)

	if err := testReadCustomers(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadContacts(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadReports(ctx, conn); err != nil {
		slog.Info(err.Error())
	}
}

func testReadCustomers(ctx context.Context, conn *co.Connector) error {
	params := common.ReadParams{
		ObjectName: "customer",
		Fields:     connectors.Fields("superuser_id", "company", "language_id"),
		Since:      time.Now().Add(-10 * time.Hour),
		Until:      time.Now(),
		// NextPage: "https://amplabs.chargeover.com/api/v3/customer?limit=1\u0026offset=1\u0026where=mod_datetime%3AGTE%3A2026-03-13T01%253A03%253A56%252B03%253A00",
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

func testReadContacts(ctx context.Context, conn *co.Connector) error {
	params := common.ReadParams{
		ObjectName: "user",
		Fields:     connectors.Fields("user_id", "phone", "title"),
		Since:      time.Now().Add(-10 * time.Hour),
		Until:      time.Now(),
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

func testReadReports(ctx context.Context, conn *co.Connector) error {
	params := common.ReadParams{
		ObjectName: "_report",
		Fields:     connectors.Fields("report_id", "name"),
		NextPage:   "",
		Since:      time.Now().Add(-10 * time.Hour),
		Until:      time.Now(),
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
