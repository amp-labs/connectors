package main

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/test/clari"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := clari.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"export/jobs", "audit/events", "/admin/limits"})
	if err != nil {
		return err
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)

	return nil
}
