package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/fastspring"
	connTest "github.com/amp-labs/connectors/test/fastspring"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	os.Exit(mainFn())
}

func mainFn() int {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetFastSpringConnector(ctx)

	slog.Info("=== products (create -> delete) ===")

	// Product path must look like a slug (e.g. "premium-laptop"); dots / odd shapes often fail validation.
	// See: https://developer.fastspring.com/reference/create-or-update-products
	productPath := fmt.Sprintf("amp-integration-%s", strings.ReplaceAll(gofakeit.UUID(), "-", ""))
	slog.Info("Creating product", "product", productPath)

	createRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"product": productPath,
			"display": map[string]any{
				"en": fmt.Sprintf("Amp integration %s", gofakeit.Word()),
			},
			"format": "digital",
			"description": map[string]any{
				"summary": map[string]any{
					"en": "Temporary connector integration test product.",
				},
			},
			"pricing": map[string]any{
				"price": map[string]any{
					"USD": 1.00,
				},
			},
		},
	})
	if err != nil {
		slog.Error("Failed to create product", "error", err)
		return 1
	}
	utils.DumpJSON(createRes, os.Stdout)

	// Prefer API response id; FastSpring may omit `product` in some responses — delete uses the path.
	recordID := createRes.RecordId
	if recordID == "" {
		recordID = productPath
	}

	defer func() {
		if recordID == "" {
			return
		}
		if err := deleteProduct(ctx, conn, recordID); err != nil {
			slog.Warn("Cleanup delete failed", "product", recordID, "error", err)
		}
	}()

	if err := deleteProduct(ctx, conn, recordID); err != nil {
		slog.Error("Failed to delete product", "error", err, "product", recordID)
		return 1
	}
	recordID = ""

	slog.Info("FastSpring products write-delete test completed successfully")
	return 0
}

func deleteProduct(ctx context.Context, conn *fastspring.Connector, productPath string) error {
	slog.Info("Deleting product", "product", productPath)

	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "products",
		RecordId:   productPath,
	})
	if err != nil {
		return err
	}
	utils.DumpJSON(res, os.Stdout)

	if !res.Success {
		return fmt.Errorf("delete reported Success=false")
	}

	return nil
}
