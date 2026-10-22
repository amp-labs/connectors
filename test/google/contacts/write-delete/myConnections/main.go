package main

import (
	"context"
	"errors"
	"fmt"
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

func updateValidation(expectedGivenName string, expectedFamilyName string) func(record map[string]any) error {
	return func(record map[string]any) error {
		object, err := jsonquery.Convertor.NodeFromMap(record)
		if err != nil {
			return err
		}

		array, err := jsonquery.New(object).ArrayRequired("names")
		if err != nil {
			return errors.New("missing names property in response")
		}

		if len(array) == 0 {
			return errors.New("names is an empty array")
		}

		nameObject := array[0]

		actualGivenName, err := jsonquery.New(nameObject).StringRequired("givenName")
		if err != nil {
			return errors.New("missing givenName property")
		}

		actualFamilyName, err := jsonquery.New(nameObject).StringRequired("familyName")
		if err != nil {
			return errors.New("missing familyName property")
		}

		if actualGivenName != expectedGivenName {
			return fmt.Errorf("mismatching givenName expected(%v), actual(%v)", expectedGivenName, actualGivenName)
		}

		if actualFamilyName != expectedFamilyName {
			return fmt.Errorf("mismatching familyName expected(%v), actual(%v)", expectedFamilyName, actualFamilyName)
		}

		return nil
	}
}
