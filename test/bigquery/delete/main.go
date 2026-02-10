package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	testBQ "github.com/amp-labs/connectors/test/bigquery"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	if len(os.Args) < 3 {
		fmt.Println("Usage: go run ./test/bigquery/delete <table_name> <record_id>")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  go run ./test/bigquery/delete my_table abc123")
		fmt.Println()
		testBQ.PrintUsage()
		os.Exit(1)
	}

	tableName := os.Args[1]
	recordId := os.Args[2]

	slog.Info("Testing Delete", "table", tableName, "recordId", recordId)

	conn := testBQ.GetBigQueryConnector(ctx)
	defer utils.Close(conn)

	result, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: tableName,
		RecordId:   recordId,
	})
	if err != nil {
		slog.Error("Delete failed", "error", err)
		os.Exit(1)
	}

	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal result", "error", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonStr))

	if result.Success {
		slog.Info("Delete completed successfully")
	} else {
		slog.Warn("Delete was not successful")
	}
}
