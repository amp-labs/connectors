package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/attio"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAttioConnector(ctx)

	// Step 1: Create multiple company records.
	slog.Info("Creating companies...")

	recordIDs := make([]string, 0, 3)

	for i := 0; i < 3; i++ {

		writeResult, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "companies",
			RecordId:   "",
			RecordData: map[string]any{
				"data": map[string]any{
					"values": map[string]any{
						"name":        "FireFox",
						"domains":     []string{gofakeit.DomainName()},
						"description": gofakeit.Sentence(10),
						"categories":  []string{"SAAS", "Web Services & Apps", "Internet"},
					},
				},
			},
		})
		if err != nil {
			utils.Fail("error writing companies to attio", "error", err, "iteration", i)
		}

		recordIDs = append(recordIDs, writeResult.RecordId)
		fmt.Printf("Created company %d with ID: %s\n", i+1, writeResult.RecordId)
	}

	// Step 2: Fetch the created records using GetRecordsByIds
	slog.Info("Fetching companies by IDs...")
	res, err := conn.GetRecordsByIds(ctx,
		"companies",
		recordIDs,
		[]string{"id", "name", "web_url", "description", "created_at"},
		nil)
	if err != nil {
		utils.Fail("error getting records by ids", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Successfully fetched companies", "count", len(res))
}
