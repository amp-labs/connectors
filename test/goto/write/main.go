package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	connTest "github.com/amp-labs/connectors/test/goto"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetGoToConnector(ctx, providers.ModuleGoTo)

	startTime := time.Now().Add(7 * 24 * time.Hour).UTC().Format(time.RFC3339)
	endTime := time.Now().Add(7*24*time.Hour + time.Hour).UTC().Format(time.RFC3339)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "webinars",
		RecordData: map[string]any{
			"subject":     "Test Webinar from Ampersand",
			"description": "Created via integration test harness.",
			"times": []map[string]any{{
				"startTime": startTime,
				"endTime":   endTime,
			}},
			"timeZone": "UTC",
		},
	})
	if err != nil {
		utils.Fail("error writing to GoTo", "error", err)
	}

	fmt.Println("Created webinar..")
	utils.DumpJSON(res, os.Stdout)
}
