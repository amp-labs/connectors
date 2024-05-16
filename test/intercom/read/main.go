package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/intercom"
	msTest "github.com/amp-labs/connectors/test/intercom"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	filePath := os.Getenv("INTERCOM_CRED_FILE")
	if filePath == "" {
		filePath = "./intercom-creds.json"
	}

	conn := msTest.GetIntercomConnector(ctx, filePath)
	defer utils.Close(conn)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "conversations",
		Fields: []string{
			"id",
			"state",
			"type",
		},
	})
	if err != nil {
		utils.Fail("error reading from Intercom", "error", err)
	}

	fmt.Println("Reading conversations..")
	utils.DumpJSON(res, os.Stdout)

	if res.Rows > intercom.DefaultPageSize {
		utils.Fail(fmt.Sprintf("expected max %v rows", intercom.DefaultPageSize))
	}
}
