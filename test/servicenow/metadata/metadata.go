package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/test/servicenow"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	conn := servicenow.GetServiceNowConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"now/table/incident", "now/contact", "now/consumer"})
	if err != nil {
		return err
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)

	return nil
}
