package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/sendgrid"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSendGridConnector(ctx)

	name := fmt.Sprintf("amp-wd-%s", gofakeit.LetterN(8))
	updatedName := name + "-updated"

	testscenario.ValidateCreateUpdateDelete(ctx, conn, "lists",
		map[string]any{
			"name": name,
		},
		map[string]any{
			"name": updatedName,
		},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "name"),
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"name": updatedName,
			},
		},
	)
}
