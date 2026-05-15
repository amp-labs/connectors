package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	cl "github.com/amp-labs/connectors/providers/calendly"
	connTest "github.com/amp-labs/connectors/test/calendly"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetCalendlyConnector(ctx)

	if _, err := conn.GetPostAuthInfo(ctx); err != nil {
		logging.Logger(ctx).Error("GetPostAuthInfo", "error", err)

		return
	}

	params := common.SubscribeParams{
		Request: &cl.SubscriptionRequest{
			CallbackURL: "https://play.svix.com/in/e_BVbta2ttNmjqeA1md230npV13f5/",
			SigningKey:  "test-secret-key",
			Scope:       "user",
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"event_types": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete,
				},
			},
		},
	}

	subscribeResult, err := conn.Subscribe(ctx, params)
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
