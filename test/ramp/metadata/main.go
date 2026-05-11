package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/ramp"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetRampConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{
		"transactions",
		"users",
		"cards",
		"departments",
		"vendors",
		"limits",
		"reimbursements",
	})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
