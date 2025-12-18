package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/shopify"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := shopify.GetShopifyConnector(ctx)

	// Test 1: Read all products with pagination
	slog.Info("=== Test 1: Reading all products (with pagination) ===")

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "products",
		Fields:     connectors.Fields("id", "title", "handle", "status", "updatedAt"),
	})

	// Test 2: Read products with small page size to verify pagination
	slog.Info("=== Test 2: Reading products with PageSize=2 ===")

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "products",
		Fields:     connectors.Fields("id", "title"),
		PageSize:   2,
	})

	// Test 3: Read products with Since filter (incremental sync)
	slog.Info("=== Test 3: Reading products with Since filter (incremental sync) ===")

	since := time.Now().AddDate(0, 0, -30)
	slog.Info("Filtering products updated since", "since", since.Format(time.RFC3339))

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "products",
		Fields:     connectors.Fields("id", "title", "updatedAt"),
		Since:      since,
	})

	// Test 4: Read all orders
	slog.Info("=== Test 4: Reading all orders ===")

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "orders",
		Fields:     connectors.Fields("id", "name", "email", "updatedAt"),
	})

	// Test 5: Read all customers
	slog.Info("=== Test 5: Reading all customers ===")

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "customers",
		Fields:     connectors.Fields("id", "firstName", "lastName", "email"),
	})

	slog.Info("=== All tests completed successfully ===")
}
