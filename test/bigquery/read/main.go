package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/bigquery"
	testBQ "github.com/amp-labs/connectors/test/bigquery"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	tableName := os.Args[1]

	// Parse flags and options.
	// Use connector defaults for consistency with production behavior.
	pageSize := bigquery.DefaultPageSize
	readAll := false
	nextPage := ""
	useSQL := false
	var fieldsList []string

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]

		switch {
		case arg == "--all":
			readAll = true
		case arg == "--sql":
			useSQL = true
		case strings.HasPrefix(arg, "--page-size="):
			if ps, err := strconv.Atoi(strings.TrimPrefix(arg, "--page-size=")); err == nil {
				pageSize = ps
			}
		case strings.HasPrefix(arg, "--next-page="):
			nextPage = strings.TrimPrefix(arg, "--next-page=")
		case strings.HasPrefix(arg, "--fields="):
			fieldsList = strings.Split(strings.TrimPrefix(arg, "--fields="), ",")
		default:
			// Try to parse as page size for backwards compatibility.
			if ps, err := strconv.Atoi(arg); err == nil {
				pageSize = ps
			} else {
				// Assume it's a field name.
				fieldsList = append(fieldsList, arg)
			}
		}
	}

	if useSQL {
		slog.Warn("--sql flag is deprecated; connector now uses Storage API only")
	}

	slog.Info("Testing Read",
		"table", tableName,
		"pageSize", pageSize,
		"readAll", readAll,
		"nextPage", nextPage,
		"fields", fieldsList,
	)

	conn := testBQ.GetBigQueryConnector(ctx)
	defer utils.Close(conn)

	params := common.ReadParams{
		ObjectName: tableName,
		PageSize:   pageSize,
		Since:      time.Now().Add(-30 * 24 * time.Hour),
	}

	if len(fieldsList) > 0 {
		if len(fieldsList) > bigquery.MaxFields {
			slog.Error("Too many fields requested",
				"count", len(fieldsList),
				"max", bigquery.MaxFields,
			)
			os.Exit(1)
		}
		params.Fields = connectors.Fields(fieldsList...)
	}

	if nextPage != "" {
		params.NextPage = common.NextPageToken(nextPage)
	}

	totalRows := int64(0)
	pageNum := 1

	for {
		slog.Info("Reading page", "page", pageNum)

		result, err := conn.Read(ctx, params)

		if err != nil {
			slog.Error("Read failed", "error", err)
			os.Exit(1)
		}

		totalRows += result.Rows

		jsonStr, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			slog.Error("Failed to marshal result", "error", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonStr))

		slog.Info("Page completed",
			"page", pageNum,
			"rows", result.Rows,
			"totalRows", totalRows,
			"done", result.Done,
			"nextPage", string(result.NextPage),
		)

		// Stop if done or not reading all pages.
		if result.Done || !readAll {
			if !result.Done && result.NextPage != "" {
				fmt.Printf("\n--- More data available. Use --next-page=%s to continue ---\n", result.NextPage)
			}
			break
		}

		// Set up next page.
		params.NextPage = result.NextPage
		pageNum++
	}

	slog.Info("Read completed", "totalRows", totalRows, "pages", pageNum)
}

func printUsage() {
	fmt.Println("Usage: go run ./test/bigquery/read <table_name> [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Printf("  --page-size=N      Number of rows per page (default: %d)\n", bigquery.DefaultPageSize)
	fmt.Println("  --all              Read all pages (default: read one page)")
	fmt.Println("  --next-page=TOKEN  Continue from a specific page token")
	fmt.Printf("  --fields=a,b,c     Comma-separated list of fields (max %d)\n", bigquery.MaxFields)
	fmt.Println("  --sql              Use SQL-based read (LIMIT/OFFSET) instead of Storage API")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run ./test/bigquery/read my_table")
	fmt.Println("  go run ./test/bigquery/read my_table --page-size=10000")
	fmt.Println("  go run ./test/bigquery/read my_table --page-size=50000 --all")
	fmt.Println("  go run ./test/bigquery/read my_table --fields=id,name,email")
	fmt.Println("  go run ./test/bigquery/read my_table --sql --fields=id,name")
	fmt.Println()
	testBQ.PrintUsage()
}
