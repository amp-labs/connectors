package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/fastspring"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetFastSpringConnector(ctx)

	slog.Info("=== Reading first subscription for update ===")

	readRes, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "subscriptions",
		Fields:     connectors.Fields("subscription", "product"),
		PageSize:   1,
	})
	if err != nil {
		slog.Error("Failed to read subscriptions", "error", err)
		os.Exit(1)
	}

	if len(readRes.Data) == 0 {
		slog.Info("No subscriptions returned; skipping write test")
		return
	}

	subID, _ := readRes.Data[0].Raw["subscription"].(string)
	product, _ := readRes.Data[0].Raw["product"].(string)
	if subID == "" {
		slog.Error("Subscription row missing subscription id")
		os.Exit(1)
	}

	// Re-send the same product path to exercise POST /subscriptions/{id} without changing catalog.
	if product == "" {
		slog.Info("Subscription missing product field; skipping write")
		return
	}

	slog.Info("=== Updating subscription (no-op product echo) ===", "subscription", subID, "product", product)

	writeRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "subscriptions",
		RecordId:   subID,
		RecordData: map[string]any{
			"product": product,
		},
	})
	if err != nil {
		slog.Error("Failed to update subscription", "error", err)
		os.Exit(1)
	}

	utils.DumpJSON(writeRes, os.Stdout)
	slog.Info("Subscription write completed")
}
