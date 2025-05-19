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
	UserName    string   `json:"UserName"`
	DisplayName string   `json:"DisplayName"`
	Name        userName `json:"Name"`
}

type userName struct {
	FamilyName string `json:"FamilyName"`
	GivenName  string `json:"GivenName"`
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
		"Users",
		CreatePayload{
			UserName:    "johnDoe",
			DisplayName: "Johnathan Doe",
			Name: userName{
				FamilyName: "Doe",
				GivenName:  "John",
			},
		},
		UpdatePayload{
			Operations: []AttributeOperation{
				{
					AttributePath:  "userName",
					AttributeValue: "Johnny",
				},
			},
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("UserId", "UserName", "DisplayName", "Name"),
			SearchBy: testscenario.Property{
				Key:   "username", // returned fields are in lowercase
				Value: "johnDoe",
			},
			RecordIdentifierKey: "userid", // returned fields are in lowercase
			UpdatedFields: map[string]string{
				"username": "Johnny",
			},
		},
	)
}
