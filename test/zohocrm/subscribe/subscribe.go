package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zohocrm"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetZohoConnector(ctx)

	subscribeParams := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Leads": {},
		},
	}

	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing", "error", err, "subscribeResult", prettyPrint(subscribeResult))
	}

	fmt.Println("Subscribe results:", prettyPrint(subscribeResult))

	updateParams := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Contacts": {},
		},
	}

	updateResult, err := conn.UpdateSubscription(ctx, updateParams, subscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error updating subscription", "error", err, "subscribeResult", prettyPrint(subscribeResult))
	}

	fmt.Println("Update subscription results:", prettyPrint(updateResult))

}

func prettyPrint(v any) string {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}
