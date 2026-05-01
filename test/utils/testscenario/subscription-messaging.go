package testscenario

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils"
)

type ConnectorWebhookSubscriber interface {
	components.SubscriptionCreator
	components.WebhookMessageVerifier
	connectors.ReadConnector
	connectors.WriteConnector
	connectors.DeleteConnector
}

type SubscribeParamBuilder func(webhookURL string) *common.SubscribeParams

type SubscribeReceiveEventsSuite struct {
	SubscribeParamBuilder SubscribeParamBuilder
	// ExpectedWebhookCalls is the number of events to wait before quiting the script.
	//
	// Number of Operations may be greater than the number of expected events.
	// Some operations can be utilized for cleaning records.
	// Ex: You are testing only update. The operations would still need create, delete, not just update.
	//
	// Webhook may be called more if there are already some subscriptions that would intervene.
	// There is no way to protect against these side effects.
	ExpectedWebhookCalls int
	Operations           []ConnectorOperation
	// WebhookRouter allows a list of alternative request handling by the webhook handler.
	// When conditions are met for the Route that handler is executed.
	// If there are no custom routes or none match then default webhook handling will take place.
	// This includes printing events until reaching the number of ExpectedWebhookCalls or
	// script cancellation.
	WebhookRouter      WebhookRouter
	VerificationParams *common.VerificationParams
}

type ConnectorMethod string

const (
	ConnectorMethodCreate ConnectorMethod = "create"
	ConnectorMethodUpdate ConnectorMethod = "update"
	ConnectorMethodDelete ConnectorMethod = "delete"
)

type ConnectorOperation struct {
	// ObjectName object to create, update or to remove.
	ObjectName string
	// Method invokes Write() or Delete().
	Method ConnectorMethod
	// Payload relevant for ConnectorMethodCreate and ConnectorMethodUpdate.
	Payload any
	// SearchProcedure relevant for ConnectorMethodUpdate and ConnectorMethodDelete.
	SearchProcedure SearchProcedure
}

type SearchProcedure struct {
	ReadFields          datautils.StringSet
	RecordIdentifierKey string
	WaitBeforeSearch    time.Duration
	SearchBy            Property
}

// ValidateSubscribeReceiveEvents is a comprehensive test scenario utilizing subscription connector operations.
//
// Flow:
// 1. Starts local server
// 2. Asks user for public URL (ngrok)
// 3. Creates subscription
// 4. Optionally triggers events (Write)
// 5. Waits for webhook(s)
// 6. Exits cleanly
func ValidateSubscribeReceiveEvents(
	ctx context.Context,
	conn ConnectorWebhookSubscriber,
	suite SubscribeReceiveEventsSuite,
) {
	fmt.Println("> TEST Subscribe/Write/Recieve")

	fmt.Println("============== Starting Webhook Handler ==================")
	messageChannel := make(chan webhookMessageResult)
	webhookURL, shutdown := startWebhookHandler(ctx, conn,
		suite.WebhookRouter, suite.VerificationParams, messageChannel,
	)
	defer shutdown()

	fmt.Printf("Local webhook server started at: \"%s\"\n", webhookURL)
	publicURL, ok := getPublicWebhookURL(ctx)
	if !ok {
		return
	}

	fmt.Println("============== Invoking connector.Subscribe() ==================")
	params := *suite.SubscribeParamBuilder(publicURL)
	subscriptionResult, err := conn.Subscribe(ctx, params)
	if printError(err) {
		return
	}

	switch subscriptionResult.Status {
	case common.SubscriptionStatusPending:
		fmt.Printf("Connector returned status (%v). Script is not designed to handle this.\n",
			common.SubscriptionStatusPending)
		return
	case common.SubscriptionStatusFailed:
		utils.DumpJSON(subscriptionResult.Result, os.Stdout)
		fmt.Println("Subscription failed ❌")
		return
	case common.SubscriptionStatusSuccess:
		utils.DumpJSON(subscriptionResult.Result, os.Stdout)
		utils.DumpJSON(subscriptionResult.ObjectEvents, os.Stdout)
		// continue script execution.
	case common.SubscriptionStatusFailedToRollback:
		fmt.Println("Subscription encountered failures and then failed to rollback ❌")
		utils.DumpJSON(subscriptionResult.Result, os.Stdout)
		utils.DumpJSON(subscriptionResult.ObjectEvents, os.Stdout)
		return
	default:
		fmt.Printf("Unknown subscription status (%v)\n", subscriptionResult.Status)
		return
	}

	fmt.Println("============== Invoking connector.Write/Delete() ==================")
	for _, trigger := range suite.Operations {
		switch trigger.Method {
		case ConnectorMethodCreate:
			fmt.Printf("Creating object %v:\n", trigger.ObjectName)
			createResult, err := createObject[any](ctx, conn, trigger.ObjectName, &trigger.Payload)
			if printError(err) {
				return
			}
			utils.DumpJSON(createResult, os.Stdout)
		case ConnectorMethodUpdate:
			objectID, ok := searchForRecord(ctx, conn, trigger.ObjectName, trigger.SearchProcedure)
			if !ok {
				return
			}

			fmt.Printf("Updating object %v:\n", trigger.ObjectName)
			updateResult, err := updateObject[any](ctx, conn, trigger.ObjectName, objectID, &trigger.Payload)
			if printError(err) {
				return
			}
			utils.DumpJSON(updateResult, os.Stdout)
		case ConnectorMethodDelete:
			objectID, ok := searchForRecord(ctx, conn, trigger.ObjectName, trigger.SearchProcedure)
			if !ok {
				return
			}

			fmt.Printf("Deleting object %v:\n", trigger.ObjectName)
			err = removeObject(ctx, conn, trigger.ObjectName, objectID)
			if printError(err) {
				return
			}
			fmt.Println("... object deleted.")
		}
	}

	// Waiting for the events to arrive. Then report on them and exit.
	// This can be stopped prematurely via context cancellation.
	receivedNumEvents := 0
	fmt.Printf("============== Waiting for %d webhook messages ==================\n", suite.ExpectedWebhookCalls)

	for receivedNumEvents < suite.ExpectedWebhookCalls {
		select {
		case message := <-messageChannel:
			receivedNumEvents++
			fmt.Printf("[%d/%d] Received webhook message:\n", receivedNumEvents, suite.ExpectedWebhookCalls)
			if message.Error == "" {
				utils.DumpJSON(message.Body, os.Stdout)
			} else {
				utils.DumpJSON(message.Error, os.Stdout)
			}

		case <-ctx.Done():
			fmt.Println("Context cancelled, stopping...")
			return
		}
	}

	fmt.Println("============== Done ==================")
}

func searchForRecord(
	ctx context.Context, conn ConnectorWebhookSubscriber, objectName string, procedure SearchProcedure,
) (string, bool) {
	if procedure.WaitBeforeSearch != 0 {
		fmt.Println("... waiting")
		time.Sleep(procedure.WaitBeforeSearch)
	}

	fmt.Printf("Search object %v by %v\n", objectName, procedure.SearchBy.String())
	res, err := readObjects(ctx, conn, objectName, procedure.ReadFields, procedure.SearchBy.Since)
	if printError(err) {
		return "", false
	}

	search := procedure.SearchBy
	object, err := searchObjectRecord(res, search.Key, search.Value)
	if printError(err) {
		return "", false
	}

	objectID := object.getRecordIdentifierValue(procedure.RecordIdentifierKey)

	return objectID, true
}
