package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/salesforce"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/tools/debug"
)

// This script verifies that when UpdateSubscription keeps an object subscribed
// but removes its quota config (omitting the object from
// SubscriptionRequest.QuotaOptimizationObjectFields), the connector:
//
//  1. Clears the FilterExpression on the kept object's PlatformEventChannelMember
//     (sends an empty filter via PATCH so Salesforce drops the prior reference)
//  2. Clears the EnrichedFields on the kept object's PlatformEventChannelMember
//  3. Removes the kept object's entry from SubscribeResult.ApexTriggers
//     (because its trigger is destructively deleted by the orphan-handling branch)
//
// Without these, the object would silently lose UPDATE events: the ECM filter
// would still reference an indicator field that no trigger maintains, so the
// filter always evaluates to false for UPDATE changeType.
//
// Run with:
//
//	go run ./test/salesforce/subscribe/kept-quota-removed/main.go \
//	  -unique <unique string> -arn <AWS Named Credential ARN>
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	var uniqueString, namedCredArn string

	flag.StringVar(&uniqueString, "unique", "", "Unique string to append to the registration label")
	flag.StringVar(&namedCredArn, "arn", "", "AWS Named Credential ARN")
	flag.Parse()

	if uniqueString == "" || namedCredArn == "" {
		log.Fatalf("missing flags: -unique <unique string> -arn <AWS Named Credential ARN>")
	}

	conn := connTest.GetSalesforceConnector(ctx)
	ctx = common.WithAuthToken(ctx, connTest.GetSalesforceAccessToken())

	registration, err := conn.Register(ctx, common.SubscriptionRegistrationParams{
		Request: &salesforce.RegistrationParams{
			UniqueRef:             "Amp" + uniqueString,
			Label:                 "Amp" + uniqueString,
			AwsNamedCredentialArn: namedCredArn,
		},
	})
	if err != nil {
		logging.Logger(ctx).Error("registration failed", "error", err)

		return
	}

	fmt.Println("Registration:", debug.PrettyFormatStringJSON(registration))

	const quotaField = "amp_cdc_optimized"

	// Initial subscription: Account WITH quota config.
	subscribeParams := common.SubscribeParams{
		RegistrationResult: registration,
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Account": {WatchFields: []string{"Name", "Phone"}},
		},
		Request: &salesforce.SubscriptionRequest{
			QuotaOptimizationObjectFields: map[common.ObjectName]string{
				"Account": quotaField,
			},
		},
	}

	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("subscribe failed", "error", err)

		return
	}

	fmt.Println("Subscribe:", debug.PrettyFormatStringJSON(subscribeResult))

	// Sanity check the initial state has filter + trigger so we can show they
	// were cleared after the update.
	initialState, ok := subscribeResult.Result.(*salesforce.SubscribeResult)
	if !ok {
		logging.Logger(ctx).Error("subscribe result has unexpected type")

		return
	}

	initialMember, ok := initialState.EventChannelMembers["Account"]
	if !ok || initialMember == nil || initialMember.Metadata == nil {
		logging.Logger(ctx).Error("subscribe did not produce a usable Account channel member")

		return
	}

	if initialMember.Metadata.FilterExpression == "" {
		logging.Logger(ctx).Error("precondition violated: Account FilterExpression is empty after initial Subscribe")

		return
	}

	if _, hasTrigger := initialState.ApexTriggers["Account"]; !hasTrigger {
		logging.Logger(ctx).Error("precondition violated: Account apex trigger missing after initial Subscribe")

		return
	}

	fmt.Printf("Initial Account FilterExpression: %q\n", initialMember.Metadata.FilterExpression)
	fmt.Printf("Initial Account trigger present: %s\n", initialState.ApexTriggers["Account"].TriggerName)

	// UpdateSubscription: keep Account subscribed but REMOVE its quota config.
	// The Request still exists, but its QuotaOptimizationObjectFields no longer
	// contains an entry for Account.
	updateParams := common.SubscribeParams{
		RegistrationResult: registration,
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Account": {WatchFields: []string{"Name", "Phone"}},
		},
		Request: &salesforce.SubscriptionRequest{
			QuotaOptimizationObjectFields: map[common.ObjectName]string{
				// Account intentionally omitted to remove its quota config.
			},
		},
	}

	updateResult, err := conn.UpdateSubscription(ctx, updateParams, subscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("update subscription failed", "error", err)

		return
	}

	fmt.Println("Update:", debug.PrettyFormatStringJSON(updateResult))

	state, ok := updateResult.Result.(*salesforce.SubscribeResult)
	if !ok {
		logging.Logger(ctx).Error("update result has unexpected type", "result", updateResult.Result)

		return
	}

	// Assertion 1: Account's ECM still exists but its FilterExpression is empty.
	updatedMember, ok := state.EventChannelMembers["Account"]
	if !ok || updatedMember == nil {
		logging.Logger(ctx).Error("FAIL: Account channel member is missing after update — the kept object should still be subscribed")
		cleanup(ctx, conn, updateResult, registration)

		return
	}

	if updatedMember.Metadata == nil {
		logging.Logger(ctx).Error("FAIL: Account channel member has nil Metadata after update")
		cleanup(ctx, conn, updateResult, registration)

		return
	}

	if updatedMember.Metadata.FilterExpression != "" {
		logging.Logger(ctx).Error("FAIL: Account FilterExpression was not cleared",
			"filter", updatedMember.Metadata.FilterExpression,
		)
		cleanup(ctx, conn, updateResult, registration)

		return
	}

	fmt.Println("PASS (1/3): Account FilterExpression is empty after quota config removed")

	// Assertion 2: Account's EnrichedFields are also cleared.
	if len(updatedMember.Metadata.EnrichedFields) != 0 {
		logging.Logger(ctx).Error("FAIL: Account EnrichedFields were not cleared",
			"count", len(updatedMember.Metadata.EnrichedFields),
		)
		cleanup(ctx, conn, updateResult, registration)

		return
	}

	fmt.Println("PASS (2/3): Account EnrichedFields are empty after quota config removed")

	// Assertion 3: Account's apex trigger has been removed from the result.
	if _, hasTrigger := state.ApexTriggers["Account"]; hasTrigger {
		logging.Logger(ctx).Error("FAIL: Account apex trigger still present in result — orphan-handling branch should have destructively deleted it")
		cleanup(ctx, conn, updateResult, registration)

		return
	}

	fmt.Println("PASS (3/3): Account apex trigger removed from result (destructively deleted as orphan)")

	fmt.Println("All assertions passed. Removing quota config from a kept object correctly clears its filter, enriched fields, and trigger.")

	cleanup(ctx, conn, updateResult, registration)
}

func cleanup(
	ctx context.Context,
	conn *salesforce.Connector,
	subscriptionResult *common.SubscriptionResult,
	registration *common.RegistrationResult,
) {
	if subscriptionResult != nil && subscriptionResult.Status == common.SubscriptionStatusSuccess {
		if err := conn.DeleteSubscription(ctx, *subscriptionResult); err != nil {
			logging.Logger(ctx).Error("cleanup DeleteSubscription failed", "error", err)
		} else {
			fmt.Println("Cleanup DeleteSubscription successful")
		}
	}

	if err := conn.DeleteRegistration(ctx, *registration); err != nil {
		logging.Logger(ctx).Error("cleanup DeleteRegistration failed", "error", err)
	} else {
		fmt.Println("Cleanup DeleteRegistration successful")
	}
}
