package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type myConnectionsPayload struct {
	Names []connectionName `json:"names"`
	Etag  string           `json:"etag,omitempty"` // required for update
}

type connectionName struct {
	GivenName  string `json:"givenName,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGoogleContactsConnector(ctx)

	givenName := gofakeit.Name()
	updatedGivenName := gofakeit.Name()
	newFamilyName := gofakeit.Name()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"myConnections",
		myConnectionsPayload{
			Names: []connectionName{{
				GivenName: givenName,
			}},
		},
		myConnectionsPayload{
			Names: []connectionName{{
				GivenName:  updatedGivenName,
				FamilyName: newFamilyName,
			}},
			// Etag: attached from create in the post processor,
		},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "names", "etag"),
			RecordIdentifierKey: "id",
			PreprocessUpdatePayload: func(createResult *common.WriteResult, updatePayload any) {
				// Updating record requires accompanying etag which is part of create response.
				payload := updatePayload.(*myConnectionsPayload)
				payload.Etag = createResult.Data["etag"].(string)
			},
			ValidateUpdatedFields: updateValidation(updatedGivenName, newFamilyName),
		},
	)
}

func updateValidation(expectedGivenName string, expectedFamilyName string) func(record map[string]any) {
	return func(record map[string]any) {
		object, err := jsonquery.Convertor.NodeFromMap(record)
		if err != nil {
			utils.Fail("invalid test", "error", err)
		}

		array, err := jsonquery.New(object).ArrayRequired("names")
		if err != nil {
			utils.Fail("missing names property in response")
		}

		if len(array) == 0 {
			utils.Fail("names is an empty array")
		}

		nameObject := array[0]

		actualGivenName, err := jsonquery.New(nameObject).StringRequired("givenName")
		if err != nil {
			utils.Fail("missing givenName property")
		}

		actualFamilyName, err := jsonquery.New(nameObject).StringRequired("familyName")
		if err != nil {
			utils.Fail("missing familyName property")
		}

		if actualGivenName != expectedGivenName {
			utils.Fail("mismatching givenName", "expected", expectedGivenName, "actual", actualGivenName)
		}

		if actualFamilyName != expectedFamilyName {
			utils.Fail("mismatching familyName", "expected", expectedFamilyName, "actual", actualFamilyName)
		}
	}
}
