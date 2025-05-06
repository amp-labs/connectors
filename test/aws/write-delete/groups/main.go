package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
	connTest "github.com/amp-labs/connectors/test/aws"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

type CreatePayload struct {
	DisplayName string `json:"DisplayName"`
	Description string `json:"Description"`
}

type UpdatePayload struct {
	Operations []AttributeOperation `json:"Operations"`
}

type AttributeOperation struct {
	AttributePath  string `json:"AttributePath"`
	AttributeValue any    `json:"AttributeValue"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAWSConnector(ctx, providers.ModuleAWSIdentityCenter)

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"Groups",
		CreatePayload{
			DisplayName: "Sales",
			Description: "Team that knows how to sell",
		},
		UpdatePayload{
			Operations: []AttributeOperation{
				{
					AttributePath:  "description",
					AttributeValue: "Best team ever",
				},
			},
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("GroupId", "DisplayName", "Description"),
			SearchBy: testscenario.Property{
				Key:   "displayname", // returned fields are in lowercase
				Value: "Sales",
			},
			RecordIdentifierKey: "groupid", // returned fields are in lowercase
			UpdatedFields: map[string]string{
				"description": "Best team ever",
			},
		},
	)
}
