package testscenario

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testscenario/internal/core"
	"github.com/amp-labs/connectors/test/utils/testscenario/internal/webhook"
)

type ConnectorWebhookSubscriber interface {
	testconn.TestableSubscriptionCreator
	testconn.TestableWebhookMessageVerifier
	connectors.ReadConnector
	connectors.WriteConnector
	connectors.DeleteConnector
}

type SubscribeParamBuilder func(webhookURL string) *common.SubscribeParams

type VerificationParamsBuilder func(subscriptionResult *common.SubscriptionResult) (*common.VerificationParams, error)

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
	// WebhookProcessor handles incoming webhook events.
	// The webhook.Server runs in the background, passes events to the connector for verification,
	// and then sends them to this Processor.
	WebhookProcessor *WebhookProcessor
	// VerificationParamsBuilder creates the parameters required to verify
	// webhook requests after Subscribe has returned its result.
	VerificationParamsBuilder VerificationParamsBuilder
	// AutoRemoveSubscription, if true, removes at the end of script execution any subscriptions
	// that were created as part of this test run. If false, subscriptions are left in place
	// and must be cleaned up manually.
	AutoRemoveSubscription bool
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

	server := webhook.CreateServer(ctx, suite.WebhookProcessor)
	webhookURL, shutdown := server.Start(ctx)
	defer shutdown()

	fmt.Printf("Local webhook server started at: \"%s\"\n", webhookURL)
	publicURL, ok := webhook.GetPublicUrl(ctx)
	if !ok {
		return
	}

	fmt.Println("============== Invoking connector.Subscribe() ==================")
	params := *suite.SubscribeParamBuilder(publicURL)
	subscriptionResult, err := conn.Subscribe(ctx, params)
	if core.PrintError(err) {
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

	// Register a defer function to clean up the successful subscription at the end of the script.
	defer cleanupSubscription(conn, suite, subscriptionResult)()

	verificationParamBuilder := suite.VerificationParamsBuilder
	if verificationParamBuilder == nil {
		verificationParamBuilder = func(_ *common.SubscriptionResult) (*common.VerificationParams, error) {
			return nil, nil
		}
	}

	verificationParams, err := verificationParamBuilder(subscriptionResult)
	if err != nil {
		fmt.Printf("Unable to configure webhook verification: %v\n", err)
		return
	}

	server.SetupConnectorVerification(conn, verificationParams)

	fmt.Println("============== Invoking connector.Write/Delete() ==================")
	for _, trigger := range suite.Operations {
		switch trigger.Method {
		case ConnectorMethodCreate:
			fmt.Printf("Creating object %v:\n", trigger.ObjectName)
			createResult, err := createObject[any](ctx, conn, trigger.ObjectName, &trigger.Payload)
			if core.PrintError(err) {
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
			if core.PrintError(err) {
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
			if core.PrintError(err) {
				return
			}
			fmt.Println("... object deleted.")
		}
	}

	// Waiting for the events to arrive. Then report on them and exit.
	// This can be stopped prematurely via context cancellation.
	fmt.Printf("============== Waiting for %d webhook messages ==================\n", suite.ExpectedWebhookCalls)

	receivedNumEvents := 0
	suite.WebhookProcessor.Run(ctx, func(event webhook.Event) bool {
		receivedNumEvents++
		fmt.Printf("[%d/%d] Received webhook message:\n", receivedNumEvents, suite.ExpectedWebhookCalls)
		if event.Error == "" {
			utils.DumpJSON(event.Body, os.Stdout)
		} else {
			utils.DumpJSON(event.Error, os.Stdout)
		}

		// Done condition.
		return receivedNumEvents == suite.ExpectedWebhookCalls
	})

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
	if core.PrintError(err) {
		return "", false
	}

	search := procedure.SearchBy
	object, err := searchObjectRecord(res, search.Key, search.Value)
	if core.PrintError(err) {
		return "", false
	}

	objectID := object.getRecordIdentifierValue(procedure.RecordIdentifierKey)

	return objectID, true
}

func cleanupSubscription(
	conn ConnectorWebhookSubscriber, suite SubscribeReceiveEventsSuite,
	subscriptionResult *common.SubscriptionResult,
) func() {
	return func() {
		if !suite.AutoRemoveSubscription {
			fmt.Println(
				"REMINDER: subscription is still active and must be removed manually.\n" +
					"To automate cleanup in the future, enable the `AutoRemoveSubscription` option.",
			)
			return
		}

		remover, ok := conn.(testconn.TestableSubscriptionRemover)
		if !ok {
			fmt.Println(
				"REMINDER: subscription is still active and must be removed manually.\n" +
					"The connector does not yet implement `components.SubscriptionRemover`.",
			)
			return
		}

		fmt.Println("[CLEANUP] Removing subscription.")

		// New fresh context. It is not canceled, nor is it expired.
		ctx := context.Background()
		// Subscription should remove all events.
		subscriptionResult.ObjectEvents = nil

		err := remover.DeleteSubscription(ctx, *subscriptionResult)
		if !core.PrintError(err) {
			fmt.Println("[CLEANUP] Subscription removed.")
		}
	}
}
