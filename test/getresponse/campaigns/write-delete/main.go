package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/getresponse"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetGetResponseConnector(ctx)

	name := fmt.Sprintf("Test Campaign %s", gofakeit.Word())
	updatedName := fmt.Sprintf("Updated Campaign %s", gofakeit.Word())

	type createPayload struct {
		Name         string `json:"name"`
		LanguageCode string `json:"languageCode"`
	}

	type updatePayload struct {
		Name string `json:"name"`
	}

	testscenario.ValidateCreateUpdateDelete(
		ctx, conn, "campaigns",
		createPayload{
			Name:         name,
			LanguageCode: "EN",
		},
		updatePayload{
			Name: updatedName,
		},
		testscenario.CRUDTestSuite{
			ReadFields:      datautils.NewSet("campaignId", "name"),
			WaitBeforeSearch: 3 * time.Second,
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: name,
			},
			RecordIdentifierKey: "campaignId",
			UpdatedFields: map[string]string{
				"name": updatedName,
			},
		},
	)
}
