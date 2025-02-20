package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/test/freshdesk"
)

func main() {
	ctx := context.Background()

	conn := freshdesk.GetFreshdeskConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"contacts", "tickets", "ticket", "products"})
	if err != nil {
		slog.Error(err.Error())
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
