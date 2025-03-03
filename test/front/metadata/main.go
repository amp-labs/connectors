package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/test/front"
)

func main() {
	ctx := context.Background()

	conn := front.GetFrontConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"accounts", "teams", "company_rules"})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
