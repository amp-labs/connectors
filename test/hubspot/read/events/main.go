package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetHubspotConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "AMPERSAND-event-occurrences-e_visited_page",
		Fields:     connectors.Fields("hs_page_id", "hs_title"),
		//Since:      utils.Timestamp("2026-05-21T23:34:50.537Z"),
		//PageSize: 2,
		//NextPage: "https://api.hubapi.com/events/event-occurrences/2026-03?eventType=e_visited_page&limit=2&after=MTc3OTQwNjQ5MDUzN3wwLTF8NzgzMjgxMDQxNjE5fDQtOTYwMDB8LTEzNzQ3MTc4NDB8",
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading...")
	utils.DumpJSON(res, os.Stdout)
}
