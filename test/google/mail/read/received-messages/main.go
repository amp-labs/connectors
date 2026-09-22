package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetGoogleMailConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "AMPERSAND-messages",
		//Fields:     connectors.Fields("id"),
		Fields:   connectors.Fields("snippet"),
		PageSize: 1,
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading...")
	utils.PrintReadResultWithoutRaw(res, os.Stdout)
}
