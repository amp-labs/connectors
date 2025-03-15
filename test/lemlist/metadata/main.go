package main

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/test/lemlist"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := lemlist.GetLemlistConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"campaigns", "team", "schedules", "schema/people"})
	if err != nil {
		return err
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)

	return nil
}
