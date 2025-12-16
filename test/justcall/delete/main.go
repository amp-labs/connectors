package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	testJustCall "github.com/amp-labs/connectors/test/justcall"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type tagPayload struct {
	Name      string `json:"name"`
	ColorCode string `json:"color_code"`
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testJustCall.GetJustCallConnector(ctx)

	// JustCall has a 15 character limit for tag names
	tagName := fmt.Sprintf("DelTag%d", gofakeit.IntRange(1000, 9999))

	testscenario.ValidateCreateDelete(ctx, conn,
		"tags",
		tagPayload{
			Name:      tagName,
			ColorCode: "#FF0000",
		},
		testscenario.CRDTestSuite{
			ReadFields: datautils.NewSet("id", "name"),
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: tagName,
			},
			RecordIdentifierKey: "id",
		},
	)
}
