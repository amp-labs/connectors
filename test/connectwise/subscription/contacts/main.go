package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/connectwise"
	connTest "github.com/amp-labs/connectors/test/connectwise"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetConnectWiseConnector(ctx)

	firstName := gofakeit.Name()
	updatedFirstName := gofakeit.Name()

	testscenario.ValidateSubscribeReceiveEvents(ctx, conn,
		testscenario.SubscribeReceiveEventsSuite{
			SubscribeParamBuilder: func(webhookURL string) *common.SubscribeParams {
				return &common.SubscribeParams{
					Request: &connectwise.SubscriptionRequest{
						WebhookURL: webhookURL,
					},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"contacts": {
							Events:            []common.SubscriptionEventType{},
							WatchFields:       nil,
							WatchFieldsAll:    false,
							PassThroughEvents: nil,
						},
					},
				}
			},
			ExpectedWebhookCalls: 3,
			Operations: []testscenario.ConnectorOperation{
				/* Create Contact */ {
					ObjectName: "contacts",
					Method:     testscenario.ConnectorMethodCreate,
					Payload:    contact{FirstName: firstName},
				},
				/* Update Contact */ {
					ObjectName: "contacts",
					Method:     testscenario.ConnectorMethodUpdate,
					Payload:    contact{FirstName: updatedFirstName},
					SearchProcedure: testscenario.SearchProcedure{
						ReadFields:          datautils.NewSet("id", "firstName"),
						RecordIdentifierKey: "id",
						SearchBy: testscenario.Property{
							Key:   "firstname",
							Value: firstName,
							Since: time.Now().Add(-10 * time.Second),
						},
					},
				},
				/* Remove Contact */ {
					ObjectName: "contacts",
					Method:     testscenario.ConnectorMethodDelete,
					SearchProcedure: testscenario.SearchProcedure{
						ReadFields:          datautils.NewSet("id", "firstName"),
						RecordIdentifierKey: "id",
						SearchBy: testscenario.Property{
							Key:   "firstname",
							Value: updatedFirstName,
							Since: time.Now().Add(-10 * time.Second),
						},
					},
				},
			},
			WebhookRouter:          testscenario.WebhookRouter{},
			VerificationParams:     nil,
			AutoRemoveSubscription: true,
		},
	)
}

type contact struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
