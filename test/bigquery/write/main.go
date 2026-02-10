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
		fmt.Println("Usage: go run ./test/bigquery/write <table_name> <json_file> [record_id]")
		fmt.Println()
		fmt.Println("For INSERT (no record_id):")
		fmt.Println("  go run ./test/bigquery/write my_table record.json")
		fmt.Println()
		fmt.Println("For UPDATE (with record_id):")
		fmt.Println("  go run ./test/bigquery/write my_table record.json abc123")
		fmt.Println()
		fmt.Println("JSON file format (example):")
		fmt.Println(`  {`)
		fmt.Println(`    "id": "test-123",`)
		fmt.Println(`    "name": "Test Record",`)
		fmt.Println(`    "description": "A test record"`)
		fmt.Println(`  }`)
		fmt.Println()
		testBQ.PrintUsage()
		os.Exit(1)
	}

	tableName := os.Args[1]
	jsonFile := os.Args[2]

	var recordId string
	if len(os.Args) > 3 {
		recordId = os.Args[3]
	}

	// Read JSON file.
	jsonData, err := os.ReadFile(jsonFile)
	if err != nil {
		slog.Error("Failed to read JSON file", "file", jsonFile, "error", err)
		os.Exit(1)
	}

	// Parse JSON into map.
	var record map[string]any
	if err := json.Unmarshal(jsonData, &record); err != nil {
		slog.Error("Failed to parse JSON", "error", err)
		os.Exit(1)
	}

	operation := "INSERT"
	if recordId != "" {
		operation = "UPDATE"
	}

	slog.Info("Testing Write", "operation", operation, "table", tableName, "recordId", recordId)

	conn := testBQ.GetBigQueryConnector(ctx)
	defer utils.Close(conn)

	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: tableName,
		RecordId:   recordId,
		RecordData: record,
	})
	if err != nil {
		slog.Error("Write failed", "error", err)
		os.Exit(1)
	}

	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal result", "error", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonStr))

	if result.Success {
		slog.Info("Write completed successfully", "operation", operation)
	} else {
		slog.Warn("Write completed with errors", "errors", result.Errors)
	}
}
