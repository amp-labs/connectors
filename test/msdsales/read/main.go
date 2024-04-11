package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	msTest "github.com/amp-labs/connectors/test/msdsales"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	filePath := os.Getenv("MS_SALES_CRED_FILE")
	if filePath == "" {
		filePath = "./ms-sales-creds.json"
	}

	conn := msTest.GetMSDynamics365SalesConnector(ctx, filePath)
	defer utils.Close(conn)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields: []string{
			"fullname", "emailaddress1", "fax", "familystatuscode",
		},
		PageSize: 2,
	})
	if err != nil {
		utils.Fail("error reading from microsoft sales", "error", err)
	}

	fmt.Println("Reading contacts..")
	utils.DumpJSON(res, os.Stdout)

	if res.Rows != 2 && !res.Done {
		utils.Fail("expected exactly 2 resource and more pages")
	}
}
