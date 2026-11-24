package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/qargo"
	"github.com/amp-labs/connectors/test/utils"
)

// Reports a payment status for one sales invoice of a test tenant:
//
//	go run ./test/qargo/write <invoice id> PAID|UNPAID
//
// UNPAID clears the paid amount and date Qargo recorded for the invoice.
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	if len(os.Args) != 3 { // nolint:mnd
		utils.Fail("usage: write <sales invoice id> PAID|UNPAID")
	}

	conn := qargo.NewConnector(ctx)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "sales-invoice/payment",
		RecordId:   os.Args[1],
		RecordData: map[string]any{"status": os.Args[2]},
	})
	if err != nil {
		utils.Fail("error writing payment status", "error", err)
	}

	fmt.Println("Payment status")
	utils.DumpJSON(res, os.Stdout)
}
