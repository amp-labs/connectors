package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	testBQ "github.com/amp-labs/connectors/test/bigquery"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./test/bigquery/metadata <table_name> [table_name2 ...]")
		fmt.Println()
		testBQ.PrintUsage()
		os.Exit(1)
	}

	tableNames := os.Args[1:]
	slog.Info("Testing ListObjectMetadata", "tables", tableNames)

	conn := testBQ.GetBigQueryConnector(ctx)
	defer utils.Close(conn)

	result, err := conn.ListObjectMetadata(ctx, tableNames)
	if err != nil {
		slog.Error("ListObjectMetadata failed", "error", err)
		os.Exit(1)
	}

	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal result", "error", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonStr))
}
