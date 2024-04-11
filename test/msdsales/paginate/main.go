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

	var customPageSize int64 = 6

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     []string{"fullname"},
		PageSize:   customPageSize,
	})
	if err != nil {
		utils.Fail("error reading from microsoft sales", "error", err)
	}

	fmt.Println("FirstPage contacts..")
	utils.DumpJSON(res, os.Stdout)

	if len(res.NextPage) == 0 {
		utils.Fail("there was no second page cursor")
	}

	if res.Rows != customPageSize {
		utils.Fail(
			fmt.Sprintf("first page has record count mismatch, given %v, expected %v", res.Rows, customPageSize))
	}

	res, err = conn.Read(ctx, common.ReadParams{
		NextPage: string(res.NextPage), // TODO should Read params have NextPageToken type?
	})
	if err != nil {
		utils.Fail("error reading from microsoft sales", "error", err)
	}

	fmt.Println("SecondPage contacts..")
	utils.DumpJSON(res, os.Stdout)

	if len(res.NextPage) != 0 || !res.Done {
		utils.Fail("there are more records on next page, but expected to be on last page")
	}
}
