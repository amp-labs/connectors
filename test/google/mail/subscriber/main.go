package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := google.GetGoogleMailConnector(ctx)

	subscribeParams := common.SubscribeParams{
		Request: map[string]any{
			"topicName":           "projects/ampersanddev/topics/gmail-event-received",
			"labelIds":            []string{"INBOX", "UNREAD", "Label_1"}, // system & custom labels.
			"labelFilterBehavior": "include",
		},
	}

	subscribeResult, err := conn.Mail.Subscribe(ctx, subscribeParams)
	if err != nil {
		slog.Error("subscription connector", "subscribe action", err)
		return
	}

	utils.DumpJSON(subscribeResult, os.Stdout)

	slog.Info("created a subscriber")

	subscribeResult, err = conn.Mail.RunScheduledMaintenance(ctx, subscribeParams, subscribeResult)
	if err != nil {
		slog.Error("subscription connector", "schedule maintenance action", err)
		return
	}

	utils.DumpJSON(subscribeResult, os.Stdout)

	subscribeResult2, err := conn.Mail.UpdateSubscription(ctx, subscribeParams, subscribeResult)
	if err != nil {
		slog.Error("subscription connector", "update subscription action", err)
		return
	}

	utils.DumpJSON(subscribeResult2, os.Stdout)

	if err := conn.Mail.DeleteSubscription(ctx, *subscribeResult); err != nil {
		slog.Error("subscription connector", "delete action", err)
		return
	}

	utils.DumpJSON(subscribeResult, os.Stdout)
}
