package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/google"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
)

const webhookAddress = "https://play.svix.com/in/e_QwY3goanA4qKfh62do3rTkOPGOm/"

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetGoogleCalendarConnector(ctx)

	// Subscribe to both supported objects.
	subscribeResult, err := conn.Subscribe(ctx, common.SubscribeParams{
		Request: &google.CalendarWatchRequest{
			Address: webhookAddress,
			Token:   "my-secret-verification-token",
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"events": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete,
				},
			},
			"calendarList": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete,
				},
			},
		},
	})
	if err != nil {
		utils.Fail("error subscribing to Google Calendar", "error", err)
	}

	slog.Info("Subscribe result", "status", subscribeResult.Status)
	utils.DumpJSON(subscribeResult, os.Stdout)
}
