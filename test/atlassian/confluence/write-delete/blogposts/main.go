package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/atlassian"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type payload struct {
	SpaceId string `json:"spaceId"`
	Status  string `json:"status"`
	Title   string `json:"title"`
}

type updatePayload struct {
	payload
	Version updateVersion `json:"version"`
}

type updateVersion struct {
	Number  int    `json:"number"`
	Message string `json:"message"`
}

const (
	// Blogpost belongs to the space.
	spaceID = "196612"
	// Default status of blogposts which are returned by Read operation.
	// Other types of blogposts
	statusPublished = "current"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()
	utils.SetupLogging()

	conn := connTest.GetConfluenceConnector(ctx)

	oldTitle := gofakeit.Name()
	newTitle := gofakeit.Name()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"blogposts",
		payload{
			SpaceId: spaceID,
			Status:  statusPublished,
			Title:   oldTitle,
		},
		updatePayload{
			payload: payload{
				SpaceId: spaceID,
				Status:  statusPublished,
				Title:   newTitle,
			},
			Version: updateVersion{
				Number:  2,
				Message: "amending title",
			},
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "title"),
			SearchBy: testscenario.Property{
				Key:   "title",
				Value: oldTitle,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"title": newTitle,
			},
		},
	)
}
