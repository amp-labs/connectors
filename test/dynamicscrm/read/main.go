package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/dynamicscrm"
	"github.com/amp-labs/connectors/test/utils"
)

var objectName = "contacts"

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetMSDynamics365CRMConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("fullname", "emailaddress1", "fax", "creditlimit_base"),
	})
	if err != nil {
		utils.Fail("error reading from microsoft CRM", "error", err)
	}

	fmt.Println("Reading contacts..")
	utils.DumpJSON(res, os.Stdout)
}
