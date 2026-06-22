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

	// Subscribe to two supported objects.
	subscribeParams := common.SubscribeParams{
		Request: &google.CalendarWatchRequest{
			Address: webhookAddress,
			Token:   "my-secret-verification-token",
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"events": {
				Events: []common.SubscriptionEventType{},
			},
			"calendarList": {
				Events: []common.SubscriptionEventType{},
			},
		},
	}

	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		utils.Fail("error subscribing to Google Calendar", "error", err)
	}

	slog.Info("Subscribe result", "status", subscribeResult.Status)
	utils.DumpJSON(subscribeResult, os.Stdout)

	// RunScheduledMaintenance renews the channels (recreate-then-stop). Feed it the
	// latest result so it knows which previous channels to stop.
	maintenanceResult, err := conn.RunScheduledMaintenance(ctx, subscribeParams, subscribeResult)
	if err != nil {
		utils.Fail("error running scheduled maintenance for Google Calendar", "error", err)
	}

	slog.Info("RunScheduledMaintenance result", "status", maintenanceResult.Status)
	utils.DumpJSON(maintenanceResult, os.Stdout)

	// UpdateSubscription to a different desired object set: drop calendarList and add
	// settings, keeping events. This exercises both the add and the remove paths.
	// Pass the latest result as the previous state so the old channels get stopped.
	updateParams := common.SubscribeParams{
		Request: &google.CalendarWatchRequest{
			Address: webhookAddress,
			Token:   "my-secret-verification-token",
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"events": {
				Events: []common.SubscriptionEventType{},
			},
			"settings": {
				Events: []common.SubscriptionEventType{},
			},
		},
	}

	updateResult, err := conn.UpdateSubscription(ctx, updateParams, maintenanceResult)
	if err != nil {
		utils.Fail("error updating Google Calendar subscription", "error", err)
	}

	slog.Info("UpdateSubscription result", "status", updateResult.Status)
	utils.DumpJSON(updateResult, os.Stdout)
}
