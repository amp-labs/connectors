package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/revenuecat"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := revenuecat.GetRevenueCatConnector(ctx)

	pageSize := 20
	if v := os.Getenv("REVENUECAT_PAGE_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			pageSize = n
		}
	}

	// Test 1: Read apps (single page)
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "apps",
		Fields:     connectors.Fields("id", "name", "object"),
		PageSize:   pageSize,
	})
	if err != nil {
		utils.Fail("error reading from revenuecat", "object", "apps", "error", err)
	}
	slog.Info("Reading apps..")
	utils.DumpJSON(res, os.Stdout)

	// Test 2: Read customers (single page)
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "customers",
		Fields:     connectors.Fields("id", "object"),
		PageSize:   pageSize,
	})
	if err != nil {
		utils.Fail("error reading from revenuecat", "object", "customers", "error", err)
	}
	slog.Info("Reading customers..")
	utils.DumpJSON(res, os.Stdout)

	// Test 3: Read entitlements (single page)
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "entitlements",
		Fields:     connectors.Fields("id", "display_name", "object"),
		PageSize:   pageSize,
	})
	if err != nil {
		utils.Fail("error reading from revenuecat", "object", "entitlements", "error", err)
	}
	slog.Info("Reading entitlements..")
	utils.DumpJSON(res, os.Stdout)

	// Test 4: Read integrations_webhooks (single page)
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "integrations_webhooks",
		Fields:     connectors.Fields("id", "name", "object"),
		PageSize:   pageSize,
	})
	if err != nil {
		utils.Fail("error reading from revenuecat", "object", "integrations_webhooks", "error", err)
	}
	slog.Info("Reading integrations webhooks..")
	utils.DumpJSON(res, os.Stdout)

	// Test 5: Read metrics_overview (single page)
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "metrics_overview",
		Fields:     connectors.Fields("id", "name", "object"),
		PageSize:   pageSize,
	})
	if err != nil {
		utils.Fail("error reading from revenuecat", "object", "metrics_overview", "error", err)
	}
	slog.Info("Reading metrics overview..")
	utils.DumpJSON(res, os.Stdout)

	// Test 6: Read offerings (single page)
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "offerings",
		Fields:     connectors.Fields("id", "display_name", "object"),
		PageSize:   pageSize,
	})
	if err != nil {
		utils.Fail("error reading from revenuecat", "object", "offerings", "error", err)
	}
	slog.Info("Reading offerings..")
	utils.DumpJSON(res, os.Stdout)

	// Test 7: Read products (single page)
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "products",
		Fields:     connectors.Fields("id", "object"),
		PageSize:   pageSize,
	})
	if err != nil {
		utils.Fail("error reading from revenuecat", "object", "products", "error", err)
	}
	slog.Info("Reading products..")
	utils.DumpJSON(res, os.Stdout)

	// Test 8: Read all products with pagination
	slog.Info("=== Test 8: Reading all products (with pagination) ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "products",
		Fields:     connectors.Fields("id", "object"),
		PageSize:   pageSize,
	})

	// Test 9: Read products with small page size to verify pagination
	slog.Info("=== Test 9: Reading products with PageSize=2 ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "products",
		Fields:     connectors.Fields("id", "object"),
		PageSize:   2,
	})

	slog.Info("Read operation completed successfully.")
}
