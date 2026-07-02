package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/jobber"
	connTest "github.com/amp-labs/connectors/test/jobber"
	"github.com/amp-labs/connectors/test/utils"
)

const (
	webhookURL = "https://play.svix.com/in/kM103sb6r97kasJTxS0HQW8EmRK/"

	// skipDelete keeps the webhook endpoints alive after Update so real events
	// can be observed at webhookURL. The run then prints the endpoint IDs and
	// the cleanup mutation; the next run fails until they are deleted.
	skipDelete = false

	// deliveryWait gives Jobber time to dispatch webhook deliveries before the
	// endpoints are modified or deleted. Jobber dispatches asynchronously
	// (observed 4-20s) and DROPS queued deliveries once the endpoint is gone,
	// so cleaning up too fast means nothing arrives at webhookURL.
	deliveryWait = 30 * time.Second
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetJobberConnector(ctx)

	subscribeParams := common.SubscribeParams{
		Request: &jobber.SubscriptionRequest{
			WebhookURL: webhookURL,
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"clients": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
				},
			},
			"quotes": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
				},
				PassThroughEvents: []string{
					"QUOTE_SENT",
				},
			},
		},
	}

	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing",
			"error", err, "subscribeResult", subscribeResult)

		return
	}

	fmt.Println("Subscribe results:")
	utils.DumpJSON(subscribeResult, os.Stdout)

	// Trigger real deliveries (visible at webhookURL): CLIENT_CREATE + CLIENT_UPDATE.
	clientID, err := triggerClientEvents(ctx, conn)
	if err != nil {
		logging.Logger(ctx).Error("Error triggering client events", "error", err)

		return
	}

	// Webhook payloads are ID-only; exercise the fetch-back used during
	// subscription processing.
	rows, err := conn.GetRecordsByIds(ctx, "clients", []string{clientID},
		[]string{"firstName", "lastName"}, nil)
	if err != nil {
		logging.Logger(ctx).Error("Error hydrating record", "error", err)

		return
	}

	fmt.Println("Hydrated record:")
	utils.DumpJSON(rows, os.Stdout)

	fmt.Printf("Waiting %s for Jobber to dispatch deliveries to %s ...\n", deliveryWait, webhookURL)
	time.Sleep(deliveryWait)

	updateParams := common.SubscribeParams{
		Request: &jobber.SubscriptionRequest{
			WebhookURL: webhookURL,
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"clients": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
				},
			},
			"jobs": {
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
			"error", err, "subscribeResult", subscribeResult)

		return
	}

	fmt.Println("Update results:")
	utils.DumpJSON(updateResult, os.Stdout)

	if skipDelete {
		return
	}

	if err := conn.DeleteSubscription(ctx, *updateResult); err != nil {
		logging.Logger(ctx).Error("Error deleting subscription", "error", err)

		return
	}

	fmt.Println("Delete subscription successful")
}

// triggerClientEvents creates and edits a client so the live subscription
// produces CLIENT_CREATE and CLIENT_UPDATE deliveries.
func triggerClientEvents(ctx context.Context, conn *jobber.Connector) (string, error) {
	writeResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "clients",
		RecordData: map[string]any{
			"firstName": "Subscriber",
			"lastName":  "LiveTest",
		},
	})
	if err != nil {
		return "", fmt.Errorf("create client: %w", err)
	}

	clientID := writeResult.RecordId

	_, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "clients",
		RecordId:   clientID,
		RecordData: map[string]any{"lastName": "LiveTestEdited"},
	})
	if err != nil {
		return "", fmt.Errorf("edit client: %w", err)
	}

	fmt.Println("Triggered CLIENT_CREATE and CLIENT_UPDATE for client:", clientID)

	return clientID, nil
}
