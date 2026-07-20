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
	description := "Created by Ampersand write-delete integration test"

	testscenario.ValidateCreateUpdateDelete(ctx, conn, "asm_groups",
		map[string]any{
			"name":        name,
			"description": description,
		},
		map[string]any{
			"name":        updatedName,
			"description": description + " (updated)",
		},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "name", "description"),
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"name": updatedName,
			},
		},
	)
}
