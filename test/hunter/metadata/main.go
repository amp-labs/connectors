package main

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/test/hunter"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := hunter.GetHunterConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"leads", "leads_lists"})
	if err != nil {
		return err
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)

	return nil
}
