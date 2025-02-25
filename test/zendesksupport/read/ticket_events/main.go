package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zendesksupport"
)

// READ:
// * paginated		- cursor, time based
// * incremental	- yes
var objectName = "ticket_events" // nolint: gochecknoglobals

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZendeskSupportConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id", "event_type"),
		// Since:      time.Now().Add(-1 * 24 * time.Hour * 62),
		// NextPage:   "https://d3v-ampersand.zendesk.com/api/v2/incremental/ticket_events.json?start_time=1735108434",
	})
	if err != nil {
		utils.Fail("error reading from Zendesk Support", "error", err)
	}

	fmt.Println("Reading ticket_events..")
	utils.DumpJSON(res, os.Stdout)
}
