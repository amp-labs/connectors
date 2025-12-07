package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/getresponse"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGetResponseConnector(ctx)

	// Test reading campaigns
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("campaignId", "name", "description", "createdOn", "isDefault"),
	})
	if err != nil {
		utils.Fail("error reading campaigns from GetResponse", "error", err)
	}

	slog.Info("Reading campaigns..")
	utils.DumpJSON(res, os.Stdout)

	// Test reading contacts
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("contactId", "email", "name", "createdOn"),
	})
	if err != nil {
		utils.Fail("error reading contacts from GetResponse", "error", err)
	}

	slog.Info("Reading contacts..")
	utils.DumpJSON(res, os.Stdout)

	// Test reading campaigns with filter and sort
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("campaignId", "name", "createdOn"),
		Filter:     "query[isDefault]=true&sort[createdOn]=DESC",
	})
	if err != nil {
		utils.Fail("error reading filtered campaigns from GetResponse", "error", err)
	}

	slog.Info("Reading filtered campaigns (isDefault=true, sorted by createdOn DESC)..")
	utils.DumpJSON(res, os.Stdout)
}
