package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/bigquery"
	testBQ "github.com/amp-labs/connectors/test/bigquery"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Parse flags.
	// Defaults match the connector constants in providers/bigquery/read.go
	tableName := flag.String("table", "", "Table name to read from (required)")
	pageSize := flag.Int("page-size", bigquery.DefaultPageSize, "Number of rows per page (default: 50000)")
	maxPages := flag.Int("max-pages", 10, "Maximum number of pages to read (0 = unlimited)")
	maxRows := flag.Int64("max-rows", 0, "Maximum number of rows to read (0 = unlimited)")
	fields := flag.String("fields", "", fmt.Sprintf("Comma-separated list of fields (max %d)", bigquery.MaxFields))
	warmup := flag.Int("warmup", 1, "Number of warmup pages before timing")
	useSQL := flag.Bool("sql", false, "Use SQL-based read (LIMIT/OFFSET) instead of Storage API")

	flag.Parse()

	if *tableName == "" {
		fmt.Println("BigQuery Read Benchmark")
		fmt.Println("=======================")
		fmt.Println()
		flag.Usage()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run ./test/bigquery/benchmark -table=my_table")
		fmt.Println("  go run ./test/bigquery/benchmark -table=my_table -page-size=50000 -max-pages=20")
		fmt.Println("  go run ./test/bigquery/benchmark -table=my_table -max-rows=1000000")
		fmt.Println("  go run ./test/bigquery/benchmark -table=my_table -fields=id,name,email")
		fmt.Println()
		testBQ.PrintUsage()
		os.Exit(1)
	}

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	if *useSQL {
		slog.Warn("--sql flag is deprecated; connector now uses Storage API only")
	}

	slog.Info("Starting benchmark",
		"table", *tableName,
		"pageSize", *pageSize,
		"maxPages", *maxPages,
		"maxRows", *maxRows,
		"fields", *fields,
		"warmup", *warmup,
	)

	conn := testBQ.GetBigQueryConnector(ctx)
	defer utils.Close(conn)

	// Build read params.
	// Default to 7 fields (MaxFields) to test the connector's limits.
	defaultFields := []string{
		"application_number",
		"country_code",
		"filing_date",
		"publication_date",
		"title_localized",
		"inventor",
		"assignee",
	}

	params := common.ReadParams{
		ObjectName: *tableName,
		PageSize:   *pageSize,
		Fields:     connectors.Fields(defaultFields...),
	}

	// Override with user-specified fields if provided.
	if *fields != "" {
		fieldList := splitFields(*fields)
		if len(fieldList) > bigquery.MaxFields {
			slog.Error("Too many fields", "count", len(fieldList), "max", bigquery.MaxFields)
			os.Exit(1)
		}
		params.Fields = connectors.Fields(fieldList...)
	}

	// Warmup phase.
	if *warmup > 0 {
		slog.Info("Running warmup", "pages", *warmup)
		warmupParams := params
		for i := 0; i < *warmup; i++ {
			result, err := conn.Read(ctx, warmupParams)
			if err != nil {
				slog.Error("Warmup failed", "error", err)
				os.Exit(1)
			}
			warmupParams.NextPage = result.NextPage
			if result.Done {
				break
			}
		}
		// Reset for actual benchmark.
		params.NextPage = ""
		runtime.GC() // Clean up warmup allocations.
	}

	// Benchmark phase.
	slog.Info("Starting timed benchmark")
	fmt.Println()
	fmt.Println("Page | Rows     | Page Time  | Total Rows | Total Time | Rows/sec")
	fmt.Println("-----|----------|------------|------------|------------|----------")

	var (
		totalRows   int64
		totalPages  int
		totalTime   time.Duration
		pageTimes   []time.Duration
		startMemory = getMemStats()
	)

	benchmarkStart := time.Now()

	for {
		pageStart := time.Now()

		result, err := conn.Read(ctx, params)
		if err != nil {
			slog.Error("Read failed", "page", totalPages+1, "error", err)
			break
		}

		pageTime := time.Since(pageStart)
		pageTimes = append(pageTimes, pageTime)
		totalPages++
		totalRows += result.Rows
		totalTime = time.Since(benchmarkStart)

		rowsPerSec := float64(totalRows) / totalTime.Seconds()

		fmt.Printf("%4d | %8d | %10s | %10d | %10s | %9.0f\n",
			totalPages,
			result.Rows,
			pageTime.Round(time.Millisecond),
			totalRows,
			totalTime.Round(time.Millisecond),
			rowsPerSec,
		)

		// Check stop conditions.
		if result.Done {
			slog.Info("Reached end of data")
			break
		}

		if *maxPages > 0 && totalPages >= *maxPages {
			slog.Info("Reached max pages limit", "maxPages", *maxPages)
			break
		}

		if *maxRows > 0 && totalRows >= *maxRows {
			slog.Info("Reached max rows limit", "maxRows", *maxRows)
			break
		}

		// Check for cancellation.
		select {
		case <-ctx.Done():
			slog.Info("Interrupted by user")
			goto summary
		default:
		}

		params.NextPage = result.NextPage
	}

summary:
	endMemory := getMemStats()

	// Calculate statistics.
	fmt.Println()
	fmt.Println("=== Benchmark Summary ===")
	fmt.Println()
	fmt.Printf("Table:           %s\n", *tableName)
	fmt.Printf("Page Size:       %d\n", *pageSize)
	fmt.Printf("Total Pages:     %d\n", totalPages)
	fmt.Printf("Total Rows:      %d\n", totalRows)
	fmt.Printf("Total Time:      %s\n", totalTime.Round(time.Millisecond))
	fmt.Println()

	if totalTime.Seconds() > 0 {
		rowsPerSec := float64(totalRows) / totalTime.Seconds()
		fmt.Printf("Throughput:      %.0f rows/sec\n", rowsPerSec)
		fmt.Printf("                 %.2f pages/sec\n", float64(totalPages)/totalTime.Seconds())
	}

	if len(pageTimes) > 0 {
		avgPageTime := totalTime / time.Duration(len(pageTimes))
		minPageTime, maxPageTime := minMax(pageTimes)
		fmt.Println()
		fmt.Printf("Avg Page Time:   %s\n", avgPageTime.Round(time.Millisecond))
		fmt.Printf("Min Page Time:   %s\n", minPageTime.Round(time.Millisecond))
		fmt.Printf("Max Page Time:   %s\n", maxPageTime.Round(time.Millisecond))
	}

	fmt.Println()
	fmt.Printf("Memory (start):  %.2f MB allocated\n", float64(startMemory.Alloc)/1024/1024)
	fmt.Printf("Memory (end):    %.2f MB allocated\n", float64(endMemory.Alloc)/1024/1024)
	fmt.Printf("Memory (total):  %.2f MB total allocated\n", float64(endMemory.TotalAlloc)/1024/1024)

	// Extrapolation for full table.
	if totalRows > 0 && totalTime.Seconds() > 0 {
		fullTableRows := int64(141853373) // User's table size
		estimatedTime := time.Duration(float64(fullTableRows) / float64(totalRows) * float64(totalTime))
		fmt.Println()
		fmt.Printf("=== Extrapolation for %d rows ===\n", fullTableRows)
		fmt.Printf("Estimated Time:  %s\n", estimatedTime.Round(time.Second))
		fmt.Printf("                 %.1f minutes\n", estimatedTime.Minutes())
		fmt.Printf("                 %.2f hours\n", estimatedTime.Hours())
	}
}

func splitFields(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for _, f := range splitByComma(s) {
		if f != "" {
			result = append(result, f)
		}
	}
	return result
}

func splitByComma(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func getMemStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

func minMax(times []time.Duration) (min, max time.Duration) {
	if len(times) == 0 {
		return 0, 0
	}
	min = times[0]
	max = times[0]
	for _, t := range times[1:] {
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
	}
	return min, max
}
