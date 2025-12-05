package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/deepmock"
	"github.com/amp-labs/connectors/internal/datautils"
	deepmocktest "github.com/amp-labs/connectors/test/deepmock"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Set up logging
	utils.SetupLogging()

	// Create context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Get configured connector
	conn := deepmocktest.GetDeepMockConnector(ctx)

	// Create sample data for testing
	setupSampleData(ctx, conn)

	slog.Info("=== DeepMock Read Operations Examples ===")
	fmt.Println()

	// Run test functions
	testReadContacts(ctx, conn)
	testReadWithPagination(ctx, conn)
	testReadWithTimeFilter(ctx, conn)
	testReadMultipleObjects(ctx, conn)

	slog.Info("All read operations completed successfully")
}

// setupSampleData creates sample records for testing read operations
func setupSampleData(ctx context.Context, conn *deepmock.Connector) {
	slog.Info("Setting up sample data for read operations")

	// Create 25 contacts for pagination testing
	for i := 0; i < 25; i++ {
		contactData := map[string]any{
			"email":     fmt.Sprintf("contact%d@example.com", i),
			"firstName": fmt.Sprintf("First%d", i),
			"lastName":  fmt.Sprintf("Last%d", i),
			"status":    "active",
			"tags":      []any{"test", "sample"},
		}
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "contacts",
			RecordData: contactData,
		})
		if err != nil {
			slog.Error("Failed to create sample contact", "index", i, "error", err)
		}
	}

	// Create some companies
	companies := []map[string]any{
		{
			"name":          "Tech Corp",
			"industry":      "technology",
			"employeeCount": 100,
			"website":       "https://techcorp.example.com",
		},
		{
			"name":          "Finance Inc",
			"industry":      "finance",
			"employeeCount": 250,
			"website":       "https://financeinc.example.com",
		},
		{
			"name":          "Health Systems",
			"industry":      "healthcare",
			"employeeCount": 500,
			"website":       "https://healthsystems.example.com",
		},
	}

	for _, company := range companies {
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "companies",
			RecordData: company,
		})
		if err != nil {
			slog.Error("Failed to create sample company", "error", err)
		}
	}

	// Create some deals with different timestamps
	baseTime := time.Now().Unix()
	deals := []map[string]any{
		{
			"title":        "Deal 1 - Old",
			"amount":       10000,
			"stage":        "prospecting",
			"lastModified": baseTime - 1000,
		},
		{
			"title":        "Deal 2 - Recent",
			"amount":       25000,
			"stage":        "qualification",
			"lastModified": baseTime - 100,
		},
		{
			"title":        "Deal 3 - Very Recent",
			"amount":       50000,
			"stage":        "proposal",
			"lastModified": baseTime - 10,
		},
	}

	for _, deal := range deals {
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "deals",
			RecordData: deal,
		})
		if err != nil {
			slog.Error("Failed to create sample deal", "error", err)
		}
	}

	slog.Info("Sample data setup completed")
	fmt.Println()
}

// testReadContacts demonstrates reading all contacts with field filtering
func testReadContacts(ctx context.Context, conn *deepmock.Connector) {
	slog.Info("Reading contacts with field filtering")

	result, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     datautils.NewStringSet("id", "email", "firstName", "lastName", "status"),
	})
	if err != nil {
		slog.Error("Failed to read contacts", "error", err)
		return
	}

	slog.Info("Contacts retrieved",
		"totalRecords", len(result.Data),
		"done", result.Done)

	// Print first 3 records as sample
	sampleSize := 3
	if len(result.Data) < sampleSize {
		sampleSize = len(result.Data)
	}

	printJSON("Sample Contacts (first 3)", result.Data[:sampleSize])
	fmt.Println()
}

// testReadWithPagination demonstrates pagination with PageSize and NextPage token
func testReadWithPagination(ctx context.Context, conn *deepmock.Connector) {
	slog.Info("Reading contacts with pagination (PageSize=10)")

	pageSize := 10
	currentPage := 1
	var nextPageToken common.NextPageToken
	totalRecords := 0

	for {
		result, err := conn.Read(ctx, common.ReadParams{
			ObjectName: "contacts",
			Fields:     datautils.NewStringSet("id", "email"),
			PageSize:   pageSize,
			NextPage:   nextPageToken,
		})
		if err != nil {
			slog.Error("Failed to read page", "page", currentPage, "error", err)
			break
		}

		recordCount := len(result.Data)
		totalRecords += recordCount

		slog.Info("Page retrieved",
			"page", currentPage,
			"recordsInPage", recordCount,
			"hasNextPage", !result.Done)

		if result.Done {
			slog.Info("Pagination complete", "totalPages", currentPage, "totalRecords", totalRecords)
			break
		}

		nextPageToken = result.NextPage
		currentPage++

		// Safety check to prevent infinite loops
		if currentPage > 100 {
			slog.Warn("Stopping pagination after 100 pages")
			break
		}
	}

	fmt.Println()
}

// testReadWithTimeFilter demonstrates incremental reads using Since parameter
func testReadWithTimeFilter(ctx context.Context, conn *deepmock.Connector) {
	slog.Info("Reading deals with time filtering")

	baseTime := time.Now().Unix()

	// Read deals modified in the last 500 seconds
	result, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "deals",
		Fields:     datautils.NewStringSet("id", "title", "amount", "stage", "lastModified"),
		Since:      time.Unix(baseTime-500, 0),
	})
	if err != nil {
		slog.Error("Failed to read deals with time filter", "error", err)
		return
	}

	slog.Info("Deals retrieved with time filter",
		"since", baseTime-500,
		"totalRecords", len(result.Data))

	printJSON("Recent Deals (last 500 seconds)", result.Data)
	fmt.Println()

	// Read deals modified in the last 50 seconds (should get fewer)
	result2, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "deals",
		Fields:     datautils.NewStringSet("id", "title", "amount"),
		Since:      time.Unix(baseTime-50, 0),
	})
	if err != nil {
		slog.Error("Failed to read deals with stricter time filter", "error", err)
		return
	}

	slog.Info("Deals retrieved with stricter time filter",
		"since", baseTime-50,
		"totalRecords", len(result2.Data))

	printJSON("Very Recent Deals (last 50 seconds)", result2.Data)
	fmt.Println()
}

// testReadMultipleObjects demonstrates reading from different object types
func testReadMultipleObjects(ctx context.Context, conn *deepmock.Connector) {
	slog.Info("Reading from multiple object types")

	// Read contacts
	contactsResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     datautils.NewStringSet("id", "email", "status"),
		PageSize:   5,
	})
	if err != nil {
		slog.Error("Failed to read contacts", "error", err)
	} else {
		slog.Info("Contacts retrieved", "count", len(contactsResult.Data))
		printJSON("Sample Contacts", contactsResult.Data)
	}

	fmt.Println()

	// Read companies
	companiesResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "companies",
		Fields:     datautils.NewStringSet("id", "name", "industry", "employeeCount"),
	})
	if err != nil {
		slog.Error("Failed to read companies", "error", err)
	} else {
		slog.Info("Companies retrieved", "count", len(companiesResult.Data))
		printJSON("All Companies", companiesResult.Data)
	}

	fmt.Println()

	// Read deals
	dealsResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "deals",
		Fields:     datautils.NewStringSet("id", "title", "amount", "stage"),
	})
	if err != nil {
		slog.Error("Failed to read deals", "error", err)
	} else {
		slog.Info("Deals retrieved", "count", len(dealsResult.Data))
		printJSON("All Deals", dealsResult.Data)
	}

	fmt.Println()

	// Summary
	slog.Info("Multi-object read summary",
		"contacts", len(contactsResult.Data),
		"companies", len(companiesResult.Data),
		"deals", len(dealsResult.Data))
}

// printJSON prints data as formatted JSON
func printJSON(label string, data any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal JSON", "error", err)
		return
	}
	fmt.Printf("%s:\n%s\n", label, string(jsonData))
}
