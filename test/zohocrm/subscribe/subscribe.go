package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/zohocrm"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zohocrm"
	"github.com/google/uuid"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetZohoConnector(ctx)
	dur := time.Minute * 2

	uniqueRef := uuid.New().String()

	subscribeParams := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Leads": common.ObjectEvents{
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete,
				},
				WatchFields: []string{
					"phone",
				},
			},
			"Contacts": common.ObjectEvents{
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete,
				},
				WatchFields: []string{
					"phone",
					"last_name",
				},
			},
		},
		Request: &zohocrm.SubscriptionRequest{
			UniqueRef:       "amp_" + uniqueRef,
			WebhookEndPoint: "https://play.svix.com/in/e_BVbta2ttNmjqeA1md230npV13f5/",
			Duration:        &dur,
		},
	}

	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing", "error", err, "subscribeResult", prettyPrint(subscribeResult))

		return
	}

	fmt.Println("Subscribe results:", prettyPrint(subscribeResult))

	updateParams := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Leads": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete,
				},
				WatchFields: []string{
					"phone",
					"company",
					"last_name",
					"first_name",
				},
			},
			"Accounts": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete,
				},
				WatchFields: []string{
					"industry",
					"phone",
				},
			},
		},
		Request: &zohocrm.SubscriptionRequest{
			UniqueRef:       "amp_" + uniqueRef,
			WebhookEndPoint: "https://play.svix.com/in/e_BVbta2ttNmjqeA1md230npV13f5/",
			Duration:        &dur,
		},
	}

	updateResult, err := conn.UpdateSubscription(ctx, updateParams, subscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error updating subscription", "error", err, "subscribeResult", prettyPrint(subscribeResult))

		return
	}

	err = conn.DeleteSubscription(ctx, *updateResult)
	if err != nil {
		logging.Logger(ctx).Error("Error deleting subscription", "error", err)

		return
	}

	fmt.Println("Delete subscription successful")
}

func prettyPrint(v any) string {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}
