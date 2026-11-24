package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/qargo"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := qargo.NewConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{"company", "sales-invoice", "sales-credit-note", "sales-invoice/payment"})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	fmt.Println("Metadata")
	utils.DumpJSON(metadata, os.Stdout)
}
