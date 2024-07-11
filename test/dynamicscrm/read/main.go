package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/dynamicscrm"
	connTest "github.com/amp-labs/connectors/test/dynamicscrm"
	"github.com/amp-labs/connectors/test/utils"
)

var (
	objectName = "contacts"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	filePath := os.Getenv("MS_CRM_CRED_FILE")
	if filePath == "" {
		filePath = "./ms-crm-creds.json"
	}

	conn := connTest.GetMSDynamics365CRMConnector(ctx, filePath)
	defer utils.Close(conn)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields: []string{
			"fullname", "emailaddress1", "fax",
		},
	})
	if err != nil {
		utils.Fail("error reading from microsoft CRM", "error", err)
	}

	fmt.Println("Reading contacts..")
	utils.DumpJSON(res, os.Stdout)

	if res.Rows > dynamicscrm.DefaultPageSize {
		utils.Fail(fmt.Sprintf("expected max %v rows", dynamicscrm.DefaultPageSize))
	}
}
