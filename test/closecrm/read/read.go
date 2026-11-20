package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
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

	if err := readActivities(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readContacts(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readLeads(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readLeadsPaginated(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func readActivities(ctx context.Context, conn *closecrm.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "activity",
		Since:      time.Now().Add(-72 * time.Hour),
		Fields:     connectors.Fields("user_id", "user_name", "source", "id"),
	}

	result, err := conn.Read(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func readContacts(ctx context.Context, conn *closecrm.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "contact",
		Fields:     connectors.Fields("name", "title", "emails", "phones", "id"),
	}

	result, err := conn.Read(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func readLeads(ctx context.Context, conn *closecrm.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "lead",
		Since:      time.Now().Add(-72 * time.Hour),
		Fields:     connectors.Fields("display_name", "description", "name", "id"),
	}

	result, err := conn.Read(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

// readLeadsPaginated reads all leads in small pages, following NextPage until done.
// Since is not set, so the offset (_skip) pagination is used instead of the Search API cursor.
func readLeadsPaginated(ctx context.Context, conn *closecrm.Connector) error {
	const maxPages = 20

	config := connectors.ReadParams{
		ObjectName: "lead",
		Fields:     connectors.Fields("display_name", "id"),
		PageSize:   5,
	}

	seen := make(map[string]bool)

	for page := 1; page <= maxPages; page++ {
		result, err := conn.Read(ctx, config)
		if err != nil {
			return err
		}

		for _, row := range result.Data {
			id, _ := row.Fields["id"].(string)
			if seen[id] {
				return fmt.Errorf("page %d: lead %s was already returned by a previous page", page, id) // nolint:err113
			}

			seen[id] = true
		}

		fmt.Printf("page %d: rows=%d done=%v next=%s\n", page, result.Rows, result.Done, result.NextPage)

		if result.Done {
			fmt.Printf("read %d distinct leads in %d pages\n", len(seen), page)

			return nil
		}

		config.NextPage = result.NextPage
	}

	return fmt.Errorf("pagination did not finish after %d pages", maxPages) // nolint:err113
}
