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
// * incremental	- n/a
var objectName = "triggers" // nolint: gochecknoglobals

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZendeskSupportConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("category_id", "description"),
		// NextPage:   "https://d3v-ampersand.zendesk.com/api/v2/triggers?page%5Bafter%5D=eyJvIjoicG9zaXRpb24scG9zaXRpb24sdGl0bGUsaWQiLCJ2IjoiYVFFQUFBQUFBQUFBYVFFQUFBQUFBQUFBY3gwQUFBQk9iM1JwWm5rZ1lYTnphV2R1WldVZ2IyWWdZWE56YVdkdWJXVnVkR21UOTlOQStoY0FBQT09In0%3D&page%5Bsize%5D=1", //nolint:lll
	})
	if err != nil {
		utils.Fail("error reading from Zendesk Support", "error", err)
	}

	fmt.Println("Reading triggers..")
	utils.DumpJSON(res, os.Stdout)
}
