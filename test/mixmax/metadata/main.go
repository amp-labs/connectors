package main

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/test/mixmax"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := mixmax.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"snippets", "users/me", "integrations/sidebars"})
	if err != nil {
		return err
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)

	return nil
}
