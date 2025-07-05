package main

import (
	"context"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	connTest "github.com/amp-labs/connectors/test/atlassian"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := connTest.GetAtlassianConnectConnector(ctx, map[string]any{
		"iss": "example",
	})

	// Use conn
	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "issue",
		Fields:     connectors.Fields("id", "project", "description"),
		Since:      time.Now().Add(-time.Hour * 24),
	})
	if err != nil {
		utils.Fail("error reading", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
}
