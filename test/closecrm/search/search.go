package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/providers/closecrm"
	testConn "github.com/amp-labs/connectors/test/closecrm"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := testConn.GetCloseConnector(ctx)

	slog.Info("Testing Close CRM Search function")

	if err := searchLeads(ctx, conn); err != nil {
		slog.Error("Failed to search leads", "error", err)
	}

	if err := searchContacts(ctx, conn); err != nil {
		slog.Error("Failed to search contacts", "error", err)
	}

	if err := searchOpportunities(ctx, conn); err != nil {
		slog.Error("Failed to search opportunities", "error", err)
	}
}

func searchLeads(ctx context.Context, conn *closecrm.Connector) error {
	slog.Info("Searching leads updated in the last 72 hours...")

	config := closecrm.SearchParams{
		ObjectName: "lead",
		Since:      time.Now().Add(-72 * time.Hour),
		Fields:     []string{"display_name", "description", "name", "id", "date_updated"},
	}

	result, err := conn.Search(ctx, config)
	if err != nil {
		return fmt.Errorf("search error: %w", err)
	}

	slog.Info("Search completed", "records", len(result.Data), "nextPage", result.NextPage)

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	fmt.Println("=== LEADS SEARCH RESULTS ===")
	fmt.Println(string(jsonStr))
	fmt.Println()

	// Test pagination if there's a next page
	if len(result.NextPage) > 0 {
		slog.Info("Testing pagination with next page token...")

		nextPageConfig := closecrm.SearchParams{
			ObjectName: "lead",
			Since:      time.Now().Add(-72 * time.Hour),
			Fields:     []string{"display_name", "description", "name", "id"},
			NextPage:   result.NextPage,
		}

		nextResult, err := conn.Search(ctx, nextPageConfig)
		if err != nil {
			return fmt.Errorf("next page error: %w", err)
		}

		slog.Info("Next page retrieved", "records", len(nextResult.Data))
	}

	return nil
}

func searchContacts(ctx context.Context, conn *closecrm.Connector) error {
	slog.Info("Searching contacts updated in the last 7 days...")

	config := closecrm.SearchParams{
		ObjectName: "contact",
		Since:      time.Now().Add(-7 * 24 * time.Hour),
		Fields:     []string{"name", "title", "emails", "phones", "id", "date_updated"},
	}

	result, err := conn.Search(ctx, config)
	if err != nil {
		return fmt.Errorf("search error: %w", err)
	}

	slog.Info("Search completed", "records", len(result.Data), "nextPage", result.NextPage)

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	fmt.Println("=== CONTACTS SEARCH RESULTS ===")
	fmt.Println(string(jsonStr))
	fmt.Println()

	return nil
}

func searchOpportunities(ctx context.Context, conn *closecrm.Connector) error {
	slog.Info("Searching opportunities updated in the last 30 days...")

	config := closecrm.SearchParams{
		ObjectName: "opportunity",
		Since:      time.Now().Add(-30 * 24 * time.Hour),
		Fields:     []string{"lead_name", "value", "status_label", "id", "date_updated"},
	}

	result, err := conn.Search(ctx, config)
	if err != nil {
		return fmt.Errorf("search error: %w", err)
	}

	slog.Info("Search completed", "records", len(result.Data), "nextPage", result.NextPage)

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	fmt.Println("=== OPPORTUNITIES SEARCH RESULTS ===")
	fmt.Println(string(jsonStr))
	fmt.Println()

	return nil
}
