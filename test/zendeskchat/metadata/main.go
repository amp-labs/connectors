package main

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/zendeskchat"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := zendeskchat.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"chats", "agents", "roles", "skills", "incremental/agent_events", "triggers"})
	if err != nil {
		return err
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)

	return nil
}
