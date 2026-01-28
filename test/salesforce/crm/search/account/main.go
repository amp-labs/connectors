package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

const TimeoutSeconds = 30

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)

	ctx, done = context.WithTimeout(ctx, TimeoutSeconds*time.Second)
	defer done()

	res, err := conn.Search(ctx, &connectors.SearchParams{
		ObjectName: "Account",
		Fields:     connectors.Fields("Id", "Name", "BillingCity", "IsDeleted", "SystemModstamp"),
		Filter: connectors.SearchFilter{
			FieldFilters: []common.FieldFilter{
				{
					FieldName: "Name",
					Operator:  common.FilterOperatorEQ,
					Value:     "Paris",
				},
			},
		},
		Limit: 3,
	})
	if err != nil {
		utils.Fail("error reading", "error", err)
	}

	fmt.Println("Reading..")
	utils.DumpJSON(res, os.Stdout)
}
