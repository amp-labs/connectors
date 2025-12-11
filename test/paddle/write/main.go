package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/paddle"
	"github.com/amp-labs/connectors/test/paddle"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := paddle.GetPaddleConnector(ctx)

	customerID, err := testCreatingCustomer(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdatingCustomer(ctx, conn, customerID)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingCustomer(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "customers",
		RecordData: map[string]any{
			"name":  gofakeit.Name(),
			"email": gofakeit.Email(),
		},
	}

	slog.Info("Creating customer...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return res.RecordId, nil
}

func testUpdatingCustomer(ctx context.Context, conn *cc.Connector, customerID string) error {
	params := common.WriteParams{
		ObjectName: "customers",
		RecordId:   customerID,
		RecordData: map[string]any{
			"name": gofakeit.Name(),
		},
	}

	slog.Info("Updating customers...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
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
