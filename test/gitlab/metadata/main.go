package main

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/test/gitlab"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := gitlab.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{
		"projects", "events", // expected to succeed with first schema provider
		"issues", "templates/gitignores", // expected to need second schema provider
		"issue", // expected to fail
	})
	if err != nil {
		return err
	}

	if len(m.Result) > 0 {
		fmt.Println("Successful Results:")

		for obj, metadata := range m.Result {
			fmt.Printf("  • %s (%s)\n", obj, metadata.DisplayName)
			fmt.Printf("%+v\n\n", metadata)
		}
	}

	if len(m.Errors) > 0 {
		fmt.Println("Errors:")

		for obj, err := range m.Errors {
			fmt.Printf("  • %s: %v\n", obj, err)
		}
	}

	return nil
}
