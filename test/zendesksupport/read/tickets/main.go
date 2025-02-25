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
// * paginated		- cursor
// * incremental	- yes
var objectName = "tickets" // nolint: gochecknoglobals

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZendeskSupportConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("description"),
		// Since:      time.Now().Add(-1 * time.Hour * 24 * 180),
		// Pagination returned when since was not specified:
		// NextPage:   "https://d3v-ampersand.zendesk.com/api/v2/incremental/tickets/cursor.json?cursor=MTcwODAxNDI2MS4wfHw1fA%3D%3D&per_page=1", // nolint:lll
		// Pagination returned when since was set:
		// NextPage: "https://d3v-ampersand.zendesk.com/api/v2/incremental/tickets/cursor.json?cursor=MTczNTEwODQzNC4wfHwyM3w%3D&per_page=1", // nolint:lll
	})
	if err != nil {
		utils.Fail("error reading from Zendesk Support", "error", err)
	}

	fmt.Println("Reading tickets..")
	utils.DumpJSON(res, os.Stdout)
}
