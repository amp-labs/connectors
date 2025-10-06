package main

import (
	"context"
	"fmt"
	"time"

	"github.com/amp-labs/connectors"
	connTest "github.com/amp-labs/connectors/test/atlassian"
)

func main() {
	ctx := context.Background()

	conn := connTest.GetAtlassianConnectConnector(ctx, map[string]any{
		"iss": "example",
	})

	// Use conn
	res, err := conn.Read(ctx, connectors.ReadParams{
		Fields: connectors.Fields("id", "project", "description"),
		Since:  time.Now().Add(-time.Hour * 24),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
