package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/connectwise"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type payload struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type patchPayload []patchOperation

type patchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetConnectWiseConnector(ctx)

	firstName := gofakeit.Name()
	updatedFirstName := gofakeit.Name()
	lastName := gofakeit.Name()
	updatedLastName := gofakeit.Name()

	fmt.Println(">>> Using PUT")

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"contacts",
		payload{
			FirstName: firstName,
			LastName:  lastName,
		},
		payload{
			FirstName: updatedFirstName,
			LastName:  updatedLastName,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "firstName", "lastName"),
			SearchBy: testscenario.Property{
				Key:   "firstname",
				Value: firstName,
				Since: time.Now().Add(-10 * time.Second),
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"firstname": updatedFirstName,
				"lastname":  updatedLastName,
			},
		},
	)

	fmt.Println()
	fmt.Println(">>> Using PATCH")

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"contacts",
		payload{
			FirstName: firstName,
			LastName:  lastName,
		},
		patchPayload{{
			// Not clear why this is needed. This was discovered by trial and error.
			Op:    "replace",
			Path:  "/customFields/1/value",
			Value: true,
		}, {
			Op:    "replace",
			Path:  "firstName",
			Value: updatedFirstName,
		}, {
			Op:    "replace",
			Path:  "/lastName",
			Value: updatedLastName,
		}},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "firstName", "lastName"),
			SearchBy: testscenario.Property{
				Key:   "firstname",
				Value: firstName,
				Since: time.Now().Add(-10 * time.Second),
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"firstname": updatedFirstName,
				"lastname":  updatedLastName,
			},
		},
	)
}
