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

// Phase identifier is used to create tickets.
const phaseID = 3

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetConnectWiseConnector(ctx)

	ticketSummary := gofakeit.Name()

	testscenario.ValidateSubscribeReceiveEvents(ctx, conn,
		testscenario.SubscribeReceiveEventsSuite{
			SubscribeParamBuilder: func(webhookURL string) *common.SubscribeParams {
				return &common.SubscribeParams{
					Request: &connectwise.SubscribeRequest{
						WebhookURL: webhookURL,
					},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"project/tickets": {
							Events:            []common.SubscriptionEventType{},
							WatchFields:       nil,
							WatchFieldsAll:    false,
							PassThroughEvents: nil,
						},
					},
				}
			},
			ExpectedWebhookCalls: 2,
			Operations: []testscenario.ConnectorOperation{
				/* Create Ticket */ {
					ObjectName: "project/tickets",
					Method:     testscenario.ConnectorMethodCreate,
					Payload:    ticket{Summary: ticketSummary, Phase: phase{ID: phaseID}},
					SearchProcedure: testscenario.SearchProcedure{
						ReadFields:          datautils.NewSet("id", "summary"),
						RecordIdentifierKey: "id",
						SearchBy: testscenario.Property{
							Key:   "summary",
							Value: ticketSummary,
							Since: time.Now().Add(-10 * time.Second),
						},
					},
				},
				/* Remove Ticket */ {
					ObjectName: "project/tickets",
					Method:     testscenario.ConnectorMethodDelete,
					SearchProcedure: testscenario.SearchProcedure{
						ReadFields:          datautils.NewSet("id", "summary"),
						RecordIdentifierKey: "id",
						SearchBy: testscenario.Property{
							Key:   "summary",
							Value: ticketSummary,
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

type ticket struct {
	Summary string `json:"summary"`
	Phase   phase  `json:"phase"`
}

type phase struct {
	ID int `json:"id"`
}
