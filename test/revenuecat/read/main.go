package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	rc "github.com/amp-labs/connectors/providers/revenuecat"
	"github.com/amp-labs/connectors/test/revenuecat"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := revenuecat.GetRevenueCatConnector(ctx)

	objects := []string{
		"apps",
		"customers",
		"entitlements",
		"integrations_webhooks",
		"metrics_overview",
		"offerings",
		"products",
	}

	for _, objectName := range objects {
		if err := testReadOnce(ctx, conn, objectName); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
}

func testReadOnce(ctx context.Context, conn *rc.Connector, objectName string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id"),
		PageSize:   20,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("revenuecat read failed (object=%s): %w", objectName, err)
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
