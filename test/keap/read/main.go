package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/keap"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetKeapConnector(ctx)
	defer utils.Close(conn)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id"),
		// Since:      time.Now().Add(-30 * time.Minute),
		// NextPage: "https://api.infusionsoft.com/crm/rest/v1/contacts/?limit=1&offset=50&since=2024-12-17T21:39:36.099Z&order=id",
	})
	if err != nil {
		utils.Fail("error reading from Keap", "error", err)
	}

	fmt.Println("Reading emails..")
	utils.DumpJSON(res, os.Stdout)
}
