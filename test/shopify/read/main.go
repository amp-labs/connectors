package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/shopify"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := shopify.GetShopifyConnector(ctx)

	// Test reading products
	slog.Info("Reading products...")

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "products",
		Fields:     connectors.Fields("id", "title", "handle", "status", "updatedAt"),
	})
	if err != nil {
		utils.Fail("error reading products from Shopify", "error", err)
	}

	slog.Info("Products result", "rows", res.Rows, "done", res.Done)
	utils.DumpJSON(res, os.Stdout)

	// Test reading products with pagination
	slog.Info("Reading products with PageSize=2...")

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "products",
		Fields:     connectors.Fields("id", "title"),
		PageSize:   2,
	})
	if err != nil {
		utils.Fail("error reading products with pagination", "error", err)
	}

	slog.Info("Products page 1", "rows", res.Rows, "done", res.Done, "hasNextPage", res.NextPage != "")
	utils.DumpJSON(res, os.Stdout)

	// Test reading products with Since filter
	slog.Info("Reading products with Since filter...")

	since := time.Now().AddDate(0, 0, -30)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "products",
		Fields:     connectors.Fields("id", "title", "updatedAt"),
		Since:      since,
	})
	if err != nil {
		utils.Fail("error reading products with Since filter", "error", err)
	}

	slog.Info("Products since", "since", since.Format(time.RFC3339), "rows", res.Rows)
	utils.DumpJSON(res, os.Stdout)

	// Test reading orders
	slog.Info("Reading orders...")

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "orders",
		Fields:     connectors.Fields("id", "name", "email", "updatedAt"),
	})
	if err != nil {
		utils.Fail("error reading orders from Shopify", "error", err)
	}

	slog.Info("Orders result", "rows", res.Rows, "done", res.Done)
	utils.DumpJSON(res, os.Stdout)

	// Test reading customers
	slog.Info("Reading customers...")

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "customers",
		Fields:     connectors.Fields("id", "firstName", "lastName", "email"),
	})
	if err != nil {
		utils.Fail("error reading customers from Shopify", "error", err)
	}

	slog.Info("Customers result", "rows", res.Rows, "done", res.Done)
	utils.DumpJSON(res, os.Stdout)
}
