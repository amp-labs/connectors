package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/keap"
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
		ObjectName: "emails",
		Fields:     connectors.Fields("id", "subject", "sent_from_address"),
	})
	if err != nil {
		utils.Fail("error reading from microsoft CRM", "error", err)
	}

	fmt.Println("Reading emails..")
	utils.DumpJSON(res, os.Stdout)

	if res.Rows > keap.DefaultPageSize {
		utils.Fail(fmt.Sprintf("expected max %v rows", keap.DefaultPageSize))
	}
}
