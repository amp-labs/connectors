package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/test/front"
)

func main() {

	ctx := context.Background()

	conn := front.GetFrontConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"accounts", "teams", "company_rules", "meme"})
	if err != nil {
		slog.Error(err.Error())
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
