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

	slog.Info("=== Reading first order for update (tags) ===")

	readRes, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "orders",
		Fields:     connectors.Fields("order"),
		PageSize:   1,
	})
	if err != nil {
		slog.Error("Failed to read orders", "error", err)
		os.Exit(1)
	}

	if len(readRes.Data) == 0 {
		slog.Info("No orders returned; skipping write test")
		return
	}

	orderID, _ := readRes.Data[0].Raw["order"].(string)
	if orderID == "" {
		slog.Error("Order row missing order id")
		os.Exit(1)
	}

	slog.Info("=== Updating order tags (integration marker) ===", "order", orderID)

	writeRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "orders",
		RecordId:   orderID,
		RecordData: map[string]any{
			"tags": map[string]any{"ampersand-integration-test": "1"},
		},
	})
	if err != nil {
		slog.Error("Failed to update order", "error", err)
		os.Exit(1)
	}

	utils.DumpJSON(writeRes, os.Stdout)
	slog.Info("Order write completed")
}
