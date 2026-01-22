package many

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
	"github.com/brianvoe/gofakeit/v6"
)

type record struct {
	ID         string         `json:"id"`
	Properties map[string]any `json:"properties"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetHubspotConnector(ctx)

	// Init records for creation.
	numRecords := 15
	records := make(common.BatchItems, numRecords)

	for i := 0; i < numRecords; i++ {
		records[i] = common.BatchItem{
			Record: map[string]any{
				"lastname":  gofakeit.Name() + " (TODO Delete)",
				"firstname": gofakeit.Name(),
			},
		}
	}

	// Batch create many records.
	res, err := conn.BatchWrite(ctx, &connectors.BatchWriteParam{
		ObjectName: "contacts",
		Type:       connectors.WriteTypeCreate,
		Batch:      records,
	})
	if err != nil {
		utils.Fail("error reading", "error", err)
	}

	fmt.Println("Creating more than 10 records")
	utils.DumpJSON(res, os.Stdout)
}
