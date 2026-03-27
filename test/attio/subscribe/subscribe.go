package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/attio"
	connTest "github.com/amp-labs/connectors/test/attio"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetAttioConnector(ctx)

	subscribeParams := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"people": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
				},
			},

			"notes": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
				},
			},
		},

		Request: &attio.SubscriptionRequest{
			WebhookEndpoint: "https://play.svix.com/in/e_tY2Mh8qhnoPeC1ZNVHZt26mroV7/",
		},
	}

	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing", "error", err)
		utils.DumpJSON(subscribeResult, os.Stdout)

		return
	}

	slog.Info("Subscription results:")
	utils.DumpJSON(subscribeResult, os.Stdout)

	err = conn.DeleteSubscription(ctx, *subscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error deleting subscription", "error", err)
		return
	}

	slog.Info("Subscription deleted successfully")

}
