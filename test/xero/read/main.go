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
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/xero"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := xero.GetXeroConnector(ctx)

	_, err := conn.GetPostAuthInfo(ctx)
	if err != nil {
		utils.Fail(err.Error())
	}

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("ContactID", "ContactNumber", "AccountNumber"),
		Since:      time.Date(2025, 6, 01, 0, 0, 0, 0, time.UTC),
	})

	if err != nil {
		utils.Fail("error reading from Xero", "error", err)
	}

	slog.Info("Reading contacts..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "contactGroups",
		Fields:     connectors.Fields("Name", "ContactGroupID", "Status"),
	})

	if err != nil {
		utils.Fail("error reading from Xero", "error", err)
	}

	slog.Info("Reading contactGroups..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "purchaseOrders",
		Fields:     connectors.Fields("PurchaseOrderID", "PurchaseOrderNumber", "FirstName"),
	})

	if err != nil {
		utils.Fail("error reading from Xero", "error", err)
	}

	slog.Info("Reading purchaseOrders..")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("Read operation completed successfully.")
}
