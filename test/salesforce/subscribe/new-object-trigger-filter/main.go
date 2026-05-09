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

// This script verifies that when UpdateSubscription adds a new object with
// quota config, that object gets:
//
//  1. A non-empty FilterExpression on its PlatformEventChannelMember
//  2. A non-empty EnrichedFields list on its PlatformEventChannelMember
//  3. An entry in SubscribeResult.ApexTriggers
//
// Without all three, quota optimization for the newly-added object is broken
// (the bug we called out as 🔴 #1 during the audit). This test would have
// caught that bug at the time and now locks in the fix.
//
// Run with:
//
//	go run ./test/salesforce/subscribe/new-object-trigger-filter/main.go \
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

	// Initial subscription: Account only.
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

	// UpdateSubscription: keep Account, add Contact (the newly-added object).
	updateParams := common.SubscribeParams{
		RegistrationResult: registration,
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Account": {WatchFields: []string{"Name", "Phone"}},
			"Contact": {WatchFields: []string{"Email"}},
		},
		Request: &salesforce.SubscriptionRequest{
			QuotaOptimizationObjectFields: map[common.ObjectName]string{
				"Account": quotaField,
				"Contact": quotaField,
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

	// Assertion 1: Contact's PlatformEventChannelMember exists and has a
	// non-empty FilterExpression.
	contactMember, ok := state.EventChannelMembers["Contact"]
	if !ok || contactMember == nil {
		logging.Logger(ctx).Error("FAIL: update result is missing newly-added Contact channel member")
		cleanup(ctx, conn, updateResult, registration)

		return
	}

	if contactMember.Metadata == nil {
		logging.Logger(ctx).Error("FAIL: Contact channel member has nil Metadata")
		cleanup(ctx, conn, updateResult, registration)

		return
	}

	if contactMember.Metadata.FilterExpression == "" {
		logging.Logger(ctx).Error("FAIL: Contact channel member has empty FilterExpression — quota optimization not configured for the newly-added object")
		cleanup(ctx, conn, updateResult, registration)

		return
	}

	fmt.Printf("PASS (1/3): Contact FilterExpression is non-empty: %q\n", contactMember.Metadata.FilterExpression)

	// Assertion 2: Contact's PlatformEventChannelMember has at least one
	// EnrichedField.
	if len(contactMember.Metadata.EnrichedFields) == 0 {
		logging.Logger(ctx).Error("FAIL: Contact channel member has empty EnrichedFields — quota optimization filter cannot evaluate without the indicator field enriched")
		cleanup(ctx, conn, updateResult, registration)

		return
	}

	fmt.Printf("PASS (2/3): Contact EnrichedFields has %d entries\n", len(contactMember.Metadata.EnrichedFields))

	// Assertion 3: Contact's apex trigger was deployed.
	contactTrigger, ok := state.ApexTriggers["Contact"]
	if !ok || contactTrigger == nil {
		logging.Logger(ctx).Error("FAIL: update result is missing newly-added Contact apex trigger — the indicator field would never be set, so all UPDATE events would be filtered out")
		cleanup(ctx, conn, updateResult, registration)

		return
	}

	if contactTrigger.TriggerName == "" {
		logging.Logger(ctx).Error("FAIL: Contact apex trigger has empty TriggerName")
		cleanup(ctx, conn, updateResult, registration)

		return
	}

	fmt.Printf("PASS (3/3): Contact apex trigger was deployed: TriggerName=%s\n", contactTrigger.TriggerName)

	fmt.Println("All assertions passed. The newly-added Contact has a complete quota optimization setup (filter, enriched fields, and trigger).")

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
