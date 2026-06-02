// Live scaffold: runs Subscribe → UpdateSubscription → DeleteSubscription.
// AccuLynx allows one Enabled subscription per installation, so re-running
// fails until any leftover is deleted.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/acculynx"
	connTest "github.com/amp-labs/connectors/test/acculynx"
	"github.com/amp-labs/connectors/test/utils"
)

const (
	consumerURL = "https://play.svix.com/in/e_t8O2GEkKHw4RL4bgd5a4it8Di5F/"
	techContact = "me@ejazkarim.com"

	// skipDelete keeps the subscription alive after Update so real events can
	// be triggered against it. Requires manual cleanup via curl.
	skipDelete = false
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetAccuLynxConnector(ctx)

	subscribeParams := common.SubscribeParams{
		Request: &acculynx.SubscriptionRequest{
			ConsumerURL: consumerURL,
			TechContact: techContact,
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"jobs": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
				},
			},
			"contacts": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeUpdate,
				},
			},
		},
	}

	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing",
			"error", err, "subscribeResult", prettyPrint(subscribeResult))

		return
	}

	fmt.Println("Subscribe results:", prettyPrint(subscribeResult))
	fmt.Println("================================================")

	updateParams := common.SubscribeParams{
		Request: &acculynx.SubscriptionRequest{
			ConsumerURL: consumerURL,
			TechContact: techContact,
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"jobs": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
				},
				PassThroughEvents: []string{
					"job.milestone.current_changed",
				},
			},
			"contacts": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
				},
			},
		},
	}

	updateResult, err := conn.UpdateSubscription(ctx, updateParams, subscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error updating subscription",
			"error", err, "subscribeResult", prettyPrint(subscribeResult))

		return
	}

	fmt.Println("Update results:", prettyPrint(updateResult))
	fmt.Println("================================================")

	if skipDelete {
		printKeepAliveInstructions(updateResult)

		return
	}

	if err := conn.DeleteSubscription(ctx, *updateResult); err != nil {
		logging.Logger(ctx).Error("Error deleting subscription", "error", err)

		return
	}

	fmt.Println("Delete subscription successful")
}

func printKeepAliveInstructions(result *common.SubscriptionResult) {
	stored, ok := result.Result.(*acculynx.SubscriptionResult)
	if !ok {
		fmt.Println("subscription left alive but couldn't recover ID for cleanup instructions")

		return
	}

	fmt.Println("Subscription left alive at:")
	fmt.Println("  id:          ", stored.SubscriptionID)
	fmt.Println("  consumerUrl: ", stored.ConsumerURL)
	fmt.Println()
	fmt.Println("Trigger an event from the AccuLynx UI (create/edit a job or contact) and")
	fmt.Println("watch the consumerUrl for the delivered webhook.")
	fmt.Println()
	fmt.Println("When done, clean up manually:")
	fmt.Printf("  curl -X DELETE -H 'Authorization: Bearer $TOKEN' \\\n"+
		"    https://api.acculynx.com/webhooks/v2/subscriptions/%s\n", stored.SubscriptionID)
}

func prettyPrint(v any) string {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}
