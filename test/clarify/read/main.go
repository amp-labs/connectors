package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/clarify"
	connTest "github.com/amp-labs/connectors/test/clarify"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetClarifyConnector(ctx)

	// Pass object names to read only those, e.g. a custom object of your workspace: go run . c_project
	objectNames := os.Args[1:]
	if len(objectNames) == 0 {
		objectNames = clarify.SupportedReadObjects()
	}

	for _, objectName := range objectNames {
		res, err := conn.Read(ctx, common.ReadParams{
			ObjectName: objectName,
			Fields:     connectors.Fields("id", "_created_at", "_updated_at"),
			PageSize:   2,
		})
		if err != nil {
			utils.Fail("error reading from Clarify", "object", objectName, "error", err)
		}

		utils.DumpJSON(res, os.Stdout)
	}
}
