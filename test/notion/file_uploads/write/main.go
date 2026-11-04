package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/notion"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetNotionConnector(ctx)

	filename := fmt.Sprintf("ampersand-test-%s.txt", gofakeit.UUID())

	// Notion has create for file_uploads only; no update or delete API — ValidateCreate, not CRUD.
	testscenario.ValidateCreate(ctx, conn, "file_uploads",
		map[string]any{
			"mode":         "single_part",
			"filename":     filename,
			"content_type": "text/plain",
		},
		testscenario.CRDTestSuite{
			ReadFields:          datautils.NewSet("id", "filename", "status"),
			RecordIdentifierKey: "id",
			SearchBy: testscenario.Property{
				Key:   "filename",
				Value: filename,
			},
		},
	)
}
