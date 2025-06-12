package main

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/test/helpscoutmailbox"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := helpscoutmailbox.GetHelpScoutConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"conversations", "customers", "mailboxes"})
	if err != nil {
		return err
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)

	return nil
}
