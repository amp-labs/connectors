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

// This script verifies that UpdateSubscription updates kept-object channel
// members in place via Tooling API PATCH (UpdateEventChannelMember) rather
// than the previous delete-and-recreate pattern.
//
// The proof is identity preservation: a PATCH leaves the
// PlatformEventChannelMember's Salesforce-assigned Id untouched, while a
// delete-and-recreate would mint a new Id. The script subscribes once,
// captures the kept member's Id, runs an UpdateSubscription that exercises
// the kept-object path, and asserts the Id is unchanged.
//
// Run with:
//
//	go run ./test/salesforce/subscribe/patch-kept-member/main.go \
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

	initialState, ok := subscribeResult.Result.(*salesforce.SubscribeResult)
	if !ok {
		logging.Logger(ctx).Error("subscribe result has unexpected type", "result", subscribeResult.Result)

		return
	}

	initialMember, ok := initialState.EventChannelMembers["Account"]
	if !ok || initialMember == nil {
		logging.Logger(ctx).Error("subscribe did not produce an Account channel member")

		return
	}

	initialID := initialMember.Id
	if initialID == "" {
		logging.Logger(ctx).Error("Account channel member has empty Id after subscribe")

		return
	}

	fmt.Printf("Initial Account PlatformEventChannelMember Id: %s\n", initialID)

	// Add Contact as a new object so UpdateSubscription has work to do; Account
	// stays as a kept object whose channel member should be PATCHed in place.
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

	updatedState, ok := updateResult.Result.(*salesforce.SubscribeResult)
	if !ok {
		logging.Logger(ctx).Error("update result has unexpected type", "result", updateResult.Result)

		return
	}

	updatedMember, ok := updatedState.EventChannelMembers["Account"]
	if !ok || updatedMember == nil {
		logging.Logger(ctx).Error("update result is missing the kept Account channel member")

		return
	}

	updatedID := updatedMember.Id

	fmt.Printf("Updated Account PlatformEventChannelMember Id: %s\n", updatedID)

	if updatedID != initialID {
		logging.Logger(ctx).Error("kept channel member Id changed across UpdateSubscription — PATCH did not happen",
			"before", initialID,
			"after", updatedID,
		)

		cleanup(ctx, conn, updateResult, registration)

		return
	}

	fmt.Println("PASS: kept Account channel member Id preserved across UpdateSubscription (PATCH used).")

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
