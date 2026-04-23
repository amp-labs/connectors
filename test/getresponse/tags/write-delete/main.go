package main

import (
	"context"
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

	name := tagName("TestTag_")

	type createPayload struct {
		Name string `json:"name"`
	}

	testscenario.ValidateCreateDelete(
		ctx, conn, "tags",
		createPayload{Name: name},
		testscenario.CRDTestSuite{
			ReadFields:       datautils.NewSet("tagId", "name"),
			WaitBeforeSearch: 2 * time.Second,
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: name,
			},
			RecordIdentifierKey: "tagid",
		},
	)
}

// tagName returns a name valid for GetResponse tags: only [A-Za-z0-9_] allowed.
func tagName(prefix string) string {
	return prefix + gofakeit.Numerify("########")
}
