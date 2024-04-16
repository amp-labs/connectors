package main

import (
	"context"
	"fmt"
	"github.com/amp-labs/connectors/msdsales"
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
			"fullname", "emailaddress1", "fax",
		},
	})
	if err != nil {
		utils.Fail("error reading from microsoft sales", "error", err)
	}

	fmt.Println("Reading contacts..")
	utils.DumpJSON(res, os.Stdout)

	if res.Rows > msdsales.DefaultPageSize {
		utils.Fail(fmt.Sprintf("expected max %v rows", msdsales.DefaultPageSize))
	}
}
