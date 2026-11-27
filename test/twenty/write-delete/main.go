package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/twenty"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

// Creates a company, finds it by name in an incremental read, renames it, reads the new name back and
// soft-deletes it: go run ./test/twenty/write-delete
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetTwentyConnector(ctx)

	// Reads start a few minutes back, so a clock behind the instance's still finds the record.
	since := time.Now().Add(-5 * time.Minute) // nolint:mnd
	name := fmt.Sprintf("amp-wd-%s", gofakeit.LetterN(8))
	updatedName := name + "-updated"

	testscenario.ValidateCreateUpdateDelete(ctx, conn, "companies",
		map[string]any{
			"name": name,
		},
		map[string]any{
			"name": updatedName,
		},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "name"),
			SearchBy:            testscenario.Property{Key: "name", Value: name, Since: since},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"name": updatedName,
			},
		},
	)
}
