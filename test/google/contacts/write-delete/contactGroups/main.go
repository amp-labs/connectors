package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type contactGroupPayload struct {
	ContactGroup contactGroupObject `json:"contactGroup"`
}

type contactGroupObject struct {
	Name string `json:"name"`
	Etag string `json:"etag,omitempty"` // required for update
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGoogleContactsConnector(ctx)

	name := gofakeit.Name()
	updatedName := gofakeit.Name()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"contactGroups",
		contactGroupPayload{
			ContactGroup: contactGroupObject{
				Name: name,
			},
		},
		contactGroupPayload{
			ContactGroup: contactGroupObject{
				Name: updatedName,
				// Etag: attached from create in the post processor,
			},
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "name", "etag"),
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: name,
			},
			RecordIdentifierKey: "id",
			PreprocessUpdatePayload: func(createResult *common.WriteResult, updatePayload any) {
				// Updating record requires accompanying etag which is part of create response.
				payload := updatePayload.(*contactGroupPayload)
				payload.ContactGroup.Etag = createResult.Data["etag"].(string)
			},
			UpdatedFields: map[string]string{
				"name": updatedName,
			},
		},
	)
}
