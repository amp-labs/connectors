package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/outreach"
	connTest "github.com/amp-labs/connectors/test/outreach"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/google/uuid"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetOutreachConnector(ctx)

	uniqueRef := uuid.New().String()

	subscribeParams := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			// "prospects": {
			// 	Events: []common.SubscriptionEventType{
			// 		common.SubscriptionEventTypeCreate,
			// 		common.SubscriptionEventTypeUpdate,
			// 		common.SubscriptionEventTypeDelete,
			// 	},
			// },
			"account": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete,
				},
			},
			// "mailings": {
			// 	Events: []common.SubscriptionEventType{
			// 		common.SubscriptionEventTypeCreate,
			// 		common.SubscriptionEventTypeUpdate,
			// 	},
			// },
		},
		Request: &outreach.SubscriptionRequest{
			UniqueRef:       "amp_" + uniqueRef,
			WebhookEndPoint: "https://play.svix.com/in/e_BVbta2ttNmjqeA1md230npV13f5/",
			Secret:          "test-secret-key",
		},
	}

	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing", "error", err, "subscribeResult", prettyPrint(subscribeResult))

		return
	}

	fmt.Println("Subscribe results:", prettyPrint(subscribeResult))

	fmt.Println("================================================")

	err = conn.DeleteSubscription(ctx, *subscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error deleting subscription", "error", err)

		return
	}

	fmt.Println("================================================")

	fmt.Println("Delete subscription successful")
}

func prettyPrint(v any) string {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}
