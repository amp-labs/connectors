package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/activecampaign"
	"github.com/amp-labs/connectors/test/utils"
)

var objectNames = []string{"contacts", "deals", "accounts", "lists", "tags", "campaigns"} // nolint: gochecknoglobals

const (
	// Pages are deliberately tiny so a handful of records is enough to cross a page boundary.
	probePageSize = 2
	// Guards against a cursor that never advances, which would otherwise hammer the live API.
	maxPages = 5
	// Incremental reads look back far enough to return records from a test account.
	incrementalLookback = -1 // years
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetActiveCampaignConnector(ctx)

	// Every supported object answers a first page.
	for _, objectName := range objectNames {
		res, err := conn.Read(ctx, common.ReadParams{
			ObjectName: objectName,
			Fields:     connectors.Fields("id"),
		})
		if err != nil {
			utils.Fail("error reading from ActiveCampaign", "error", err)
		}

		fmt.Printf("Reading %s...\n", objectName)
		utils.DumpJSON(res, os.Stdout)
	}

	// Contacts walk forward with id_greater.
	walkPages(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "email"),
		PageSize:   probePageSize,
	})

	// Every other object pages by offset.
	walkPages(ctx, conn, common.ReadParams{
		ObjectName: "deals",
		Fields:     connectors.Fields("id", "title"),
		PageSize:   probePageSize,
	})

	// Incremental reads combine filters[updated_after] with the id_greater cursor.
	walkPages(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "email"),
		PageSize:   probePageSize,
		Since:      time.Now().AddDate(incrementalLookback, 0, 0),
	})
}

// walkPages advances pagination by hand rather than using testscenario.ReadThroughPages, so that a
// cursor failing to advance is reported instead of looping against the live API indefinitely.
func walkPages(ctx context.Context, conn connectors.ReadConnector, params common.ReadParams) {
	slog.Info("Paginating", "object", params.ObjectName, "pageSize", params.PageSize, "since", params.Since)

	seenPages := make(map[string]bool)

	for page := 1; page <= maxPages; page++ {
		result, err := conn.Read(ctx, params)
		if err != nil {
			utils.Fail("error reading from ActiveCampaign", "error", err)
		}

		slog.Info("Page read",
			"object", params.ObjectName, "page", page, "rows", result.Rows, "done", result.Done)
		utils.PrintReadResultWithoutRaw(result, os.Stdout)

		if result.Done {
			slog.Info("Pagination finished", "object", params.ObjectName, "pages", page)

			return
		}

		next := result.NextPage.String()
		if seenPages[next] {
			utils.Fail("pagination is not advancing, the same page was requested twice",
				"object", params.ObjectName, "nextPage", next)
		}

		seenPages[next] = true
		params.NextPage = result.NextPage
	}

	slog.Warn("Stopped at the page cap without reaching the end of the collection",
		"object", params.ObjectName, "maxPages", maxPages)
}
