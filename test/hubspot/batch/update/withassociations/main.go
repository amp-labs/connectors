package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetHubspotConnector(ctx)

	contactID := "211762804612"
	companyID1 := "18417469260"
	companyID2 := "29013756946"

	records := common.BatchItems{
		{
			Record: map[string]any{
				"id":        contactID, // required for update
				"lastname":  "Sponge_UPDATED",
				"firstname": "Bob_UPDATED",
			},
			Associations: []Association{
				newAssociation(
					// relationship to some company #1
					companyID1,
					// contact to company
					279,
				),
				newAssociation(
					// relationship to some company #2
					companyID2,
					// contact to company
					279,
				),
				newAssociation(
					// relationship to unknown company (results in error)
					"unknown_id_123456798",
					// contact to contact
					449,
				),
				newAssociation(
					// relationship to itself (not allowed)
					contactID,
					// contact to contact
					449,
				),
			},
		},
	}

	res, err := conn.BatchWrite(ctx, &connectors.BatchWriteParam{
		ObjectName: "contacts",
		Type:       connectors.WriteTypeUpdate,
		Batch:      records,
	})
	if err != nil {
		utils.Fail("error reading", "error", err)
	}

	fmt.Println("Update contact and create associations")
	utils.DumpJSON(res, os.Stdout)
}

type Association struct {
	To    AssociationTo     `json:"to"`
	Types []AssociationType `json:"types"`
}

type AssociationTo struct {
	ID string `json:"id"`
}

type AssociationType struct {
	AssociationCategory string `json:"associationCategory"`
	AssociationTypeId   int    `json:"associationTypeId"`
}

func newAssociation(toID string, associationTypeID int) Association {
	return Association{
		To: AssociationTo{
			ID: toID,
		},
		Types: []AssociationType{{
			AssociationCategory: "HUBSPOT_DEFINED",
			AssociationTypeId:   associationTypeID,
		}},
	}
}
