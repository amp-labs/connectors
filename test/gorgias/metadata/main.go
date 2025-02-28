package main

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/test/gorgias"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := gorgias.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"account", "customers", "events"})
	if err != nil {
		return err
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)

	return nil
}
