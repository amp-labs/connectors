package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/hubspot"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		utils.Fail("test failed", "error", err)
	}
}

func run() error {
	ctx := context.Background()
	conn := connTest.GetHubspotConnector(ctx)

	waitAfterCreation := 20 * time.Second

	contactEmail := gofakeit.Email()
	contact, cleanup, err := testscenario.SetupRecord(ctx, conn, "contact",
		map[string]any{
			"email":     contactEmail,
			"firstname": gofakeit.FirstName(),
			"lastname":  gofakeit.LastName(),
		}, testscenario.RecordCreationRecipe{
			ReadFields:       datautils.NewSet("id", "email"),
			WaitBeforeSearch: waitAfterCreation,
			SearchBy: testscenario.Property{
				Key:   "email",
				Value: contactEmail,
				Since: time.Now().Add(-1 * time.Minute),
			},
			RecordIdentifierKey: "id",
		})
	if err != nil {
		return err
	}
	defer cleanup()

	taskContent1 := gofakeit.Quote()
	task1, cleanup, err := testscenario.SetupRecord(ctx, conn, "task",
		map[string]any{
			"hs_task_body": taskContent1,
			"hs_timestamp": "2024-11-12T15:48:22Z",
		}, testscenario.RecordCreationRecipe{
			ReadFields:       datautils.NewSet("id", "hs_task_body"),
			WaitBeforeSearch: waitAfterCreation,
			SearchBy: testscenario.Property{
				Key:   "hs_task_body",
				Value: taskContent1,
			},
			RecordIdentifierKey: "id",
		})
	if err != nil {
		return err
	}
	defer cleanup()

	taskContent2 := gofakeit.Quote()
	task2, cleanup, err := testscenario.SetupRecord(ctx, conn, "task",
		map[string]any{
			"hs_task_body": taskContent2,
			"hs_timestamp": "2024-11-12T15:48:22Z",
		}, testscenario.RecordCreationRecipe{
			ReadFields:       datautils.NewSet("id", "hs_task_body"),
			WaitBeforeSearch: waitAfterCreation,
			SearchBy: testscenario.Property{
				Key:   "hs_task_body",
				Value: taskContent2,
			},
			RecordIdentifierKey: "id",
		})
	if err != nil {
		return err
	}
	defer cleanup()

	fmt.Println("[TEST] Update contact and create associations")
	res, err := batchCreateItems(ctx, conn, contact.Id, task1.Id, task2.Id)
	if err != nil {
		return err
	}

	fmt.Println("================")
	utils.DumpJSON(res, os.Stdout)
	fmt.Println("================")

	return nil
}

func batchCreateItems(
	ctx context.Context, conn *hubspot.Connector,
	contactID, taskID1, taskID2 string,
) (*common.BatchWriteResult, error) {
	// https://developers.hubspot.com/docs/api-reference/latest/crm/associations/associate-records/guide#contact-to-object
	relContactToTask := 203
	relContactToContact := 449

	records := common.BatchItems{{
		Record: map[string]any{
			"id":        contactID, // required for update
			"lastname":  "Sponge_UPDATED",
			"firstname": "Bob_UPDATED",
		},
		Associations: []any{
			// relationship to some task #1
			newAssociation(taskID1, relContactToTask),
			// relationship to some task #2
			newAssociation(taskID2, relContactToTask),
			// relationship to unknown company (results in error)
			newAssociation("unknown_id_123456798", relContactToContact),
			// relationship to itself (not allowed)
			newAssociation(contactID, relContactToContact),
		},
	}}

	return conn.BatchWrite(ctx, &connectors.BatchWriteParam{
		ObjectName: "contacts",
		Type:       connectors.WriteTypeUpdate,
		Batch:      records,
	})
}

func newAssociation(toID string, associationTypeID int) map[string]any {
	return map[string]any{
		"to": map[string]any{
			"id": toID,
		},
		"types": []map[string]any{{
			"associationCategory": "HUBSPOT_DEFINED",
			"associationTypeId":   associationTypeID,
		}},
	}
}
