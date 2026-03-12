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
	updatedName := tagName("Updated_")

	type createPayload struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	type updatePayload struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	testscenario.ValidateCreateUpdateDelete(
		ctx, conn, "tags",
		createPayload{
			Name:  name,
			Color: "#3498db",
		},
		updatePayload{
			Name:  updatedName,
			Color: "#e74c3c",
		},
		testscenario.CRUDTestSuite{
			ReadFields:       datautils.NewSet("tagId", "name"),
			WaitBeforeSearch: 2 * time.Second,
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: name,
			},
			RecordIdentifierKey: "tagid",
			UpdatedFields: map[string]string{
				"name": updatedName,
			},
		},
	)
}

// tagName returns a name valid for GetResponse tags: only [A-Za-z0-9_] allowed.
func tagName(prefix string) string {
	return prefix + gofakeit.Numerify("########")
}
