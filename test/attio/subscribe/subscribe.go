// Live scaffold: runs Subscribe → UpdateSubscription → DeleteSubscription against
// a real Attio workspace with real credentials.
//
// It subscribes to a standard object (people, which uses record.* events with an
// id.object_id filter) and a core object (notes, which uses object-specific
// events), so a single run exercises both subscribe patterns. Set skipDelete=true
// to keep the webhook alive and print its id + signing secret so you can trigger
// real events and verify signatures.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/attio"
	connTest "github.com/amp-labs/connectors/test/attio"
	"github.com/amp-labs/connectors/test/utils"
)

const (
	// webhookEndpoint is where Attio delivers events. Replace with your own tunnel/inbox
	// (e.g. a Svix Play inbox or an ngrok URL) to observe deliveries.
	webhookEndpoint = "https://play.svix.com/in/e_tY2Mh8qhnoPeC1ZNVHZt26mroV7/"

	// skipDelete keeps the subscription alive after Update so real events can be triggered
	// against it. Requires manual cleanup via the DELETE curl printed at the end.
	skipDelete = false
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetAttioConnector(ctx)

	subscribeParams := common.SubscribeParams{
		Request: &attio.SubscriptionRequest{WebhookEndpoint: webhookEndpoint},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			// Standard object: subscribed via record.* events + an id.object_id filter.
			"people": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
				},
			},
			// Core object: subscribed via object-specific events (note.created, ...).
			"notes": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
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

	// Update: broaden people to include delete, and add companies.
	updateParams := common.SubscribeParams{
		Request: &attio.SubscriptionRequest{WebhookEndpoint: webhookEndpoint},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"people": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete,
				},
			},
			"companies": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
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

// printKeepAliveInstructions prints the webhook id + signing secret so real events can be triggered
// and signatures verified, plus the manual cleanup command.
func printKeepAliveInstructions(result *common.SubscriptionResult) {
	stored, ok := result.Result.(*attio.SubscriptionResult)
	if !ok {
		fmt.Println("subscription left alive but couldn't recover the webhook id for cleanup instructions")

		return
	}

	fmt.Println("Subscription left alive:")
	fmt.Println("  webhook_id: ", stored.Data.Id.WebhookId)
	fmt.Println("  secret:     ", stored.Data.Secret)
	fmt.Println("  target_url: ", stored.Data.TargetURL)
	fmt.Println()
	fmt.Println("Trigger an event in the Attio UI (create/edit a person, note, or company) and watch")
	fmt.Println("the target_url inbox for the delivered webhook. Verify the Attio-Signature header as")
	fmt.Println("HMAC-SHA256(rawBody) using the secret above.")
	fmt.Println()
	fmt.Println("When done, clean up manually:")
	fmt.Printf("  curl -X DELETE -H 'Authorization: Bearer $TOKEN' \\\n"+
		"    https://api.attio.com/v2/webhooks/%s\n", stored.Data.Id.WebhookId)
}

func prettyPrint(v any) string {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}
