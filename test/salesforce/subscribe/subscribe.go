package main

import (
	"context"
	"encoding/json"
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
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	var (
		uniqueString string
		namedCredArn string
		mode         string
	)

	flag.StringVar(&uniqueString, "unique", "", "Unique string to append to the registration label")
	flag.StringVar(&namedCredArn, "arn", "", "AWS Named Credential ARN")
	flag.StringVar(&mode, "mode", "auto",
		"Which scenario to run: 'auto' (connector creates quota fields and deploys triggers), "+
			"'manual' (caller-managed quota fields and triggers), 'none' (no quota optimization), "+
			"or 'all' (run all three)")
	flag.Parse()

	if uniqueString == "" || namedCredArn == "" {
		log.Fatalf("missing flags: go run <your path>.go -unique <unique string> -arn <AWS Named Credential ARN> [-mode auto|manual|none|all]")
	}

	conn := connTest.GetSalesforceConnector(ctx)
	ctx = common.WithAuthToken(ctx, connTest.GetSalesforceAccessToken())

	switch mode {
	case "auto":
		runAutoFlow(ctx, conn, uniqueString, namedCredArn)
	case "manual":
		runManualFlow(ctx, conn, uniqueString, namedCredArn)
	case "none":
		runNoOptimizationFlow(ctx, conn, uniqueString, namedCredArn)
	case "all":
		runAutoFlow(ctx, conn, uniqueString, namedCredArn)
		runManualFlow(ctx, conn, uniqueString, namedCredArn)
		runNoOptimizationFlow(ctx, conn, uniqueString, namedCredArn)
	default:
		log.Fatalf("invalid -mode %q (expected auto|manual|none|all)", mode)
	}
}

// runAutoFlow exercises the connector-managed path: it creates quota fields,
// auto-deploys apex triggers, then walks Subscribe → UpdateSubscription twice → Delete.
func runAutoFlow(ctx context.Context, conn *salesforce.Connector, uniqueString, arn string) {
	params := common.SubscriptionRegistrationParams{
		Request: &salesforce.RegistrationParams{
			UniqueRef:             "Amp" + uniqueString,
			Label:                 "Amp" + uniqueString,
			AwsNamedCredentialArn: arn,
		},
	}

	result, err := conn.Register(ctx, params)
	if err != nil {
		logging.Logger(ctx).Error("Error registering", "error", err)
		return
	}

	fmt.Println("Registration result:", prettyPrint(result))

	subscribeParams := common.SubscribeParams{
		RegistrationResult: result,
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Account": {
				WatchFields: []string{"Name", "Phone"},
			},
		},
		Request: &salesforce.SubscriptionRequest{
			QuotaOptimizationObjectFields: map[common.ObjectName]string{
				"Account": "amp_cdc_optimized",
			},
		},
	}

	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing", "error", err, "subscribeResult", prettyPrint(subscribeResult))

		return
	}

	fmt.Println("Subscribe result:", prettyPrint(subscribeResult))

	// Update subscription: keep Account, add Contact, remove nothing.
	// This exercises:
	// - Kept objects with filter updates via PATCH (Account stays, gets auto-built filter)
	// - New objects being subscribed (Contact)
	// - Quota optimization fields: Account is kept (no delete+recreate), Contact is new
	updateParams := common.SubscribeParams{
		RegistrationResult: result,
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Account": {
				WatchFields: []string{"Name", "Phone"},
			},
			"Contact": {
				WatchFields: []string{"Email"},
			},
		},
		Request: &salesforce.SubscriptionRequest{
			QuotaOptimizationObjectFields: map[common.ObjectName]string{
				"Account": "amp_cdc_optimized",
				"Contact": "amp_cdc_optimized",
			},
		},
	}

	updateResult, err := conn.UpdateSubscription(ctx, updateParams, subscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error updating subscription", "error", err)

		return
	}

	fmt.Println("Update subscription result (keep Account + add Contact):", prettyPrint(updateResult))

	// Second update: remove Account, keep Contact. This exercises:
	// - Removed objects having their quota fields deleted
	// - Kept objects not being touched
	update2Params := common.SubscribeParams{
		RegistrationResult: result,
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Contact": {
				WatchFields: []string{"Email"},
			},
		},
		Request: &salesforce.SubscriptionRequest{
			QuotaOptimizationObjectFields: map[common.ObjectName]string{
				"Contact": "amp_cdc_optimized",
			},
		},
	}

	update2Result, err := conn.UpdateSubscription(ctx, update2Params, updateResult)
	if err != nil {
		logging.Logger(ctx).Error("Error updating subscription (remove Account)", "error", err)

		return
	}

	fmt.Println("Update subscription result (remove Account):", prettyPrint(update2Result))

	if update2Result != nil && update2Result.Status == common.SubscriptionStatusSuccess {
		if err := conn.DeleteSubscription(ctx, *update2Result); err != nil {
			logging.Logger(ctx).Error("Error unsubscribing", "error", err)

			return
		}
	}

	fmt.Println("Delete subscription successful")

	if err := conn.DeleteRegistration(ctx, *result); err != nil {
		logging.Logger(ctx).Error("Error rolling back registration", "error", err)

		return
	}

	fmt.Println("Delete registration successful")
}

// runManualFlow exercises the caller-managed path: UseExistingQuotaOptimizationFields=true
// and ManualApexTriggerDeployment=true. The connector skips both quota field creation and
// apex trigger deployment — the caller is expected to have created the checkbox field
// (amp_cdc_optimized) and the trigger out of band.
func runManualFlow(ctx context.Context, conn *salesforce.Connector, uniqueString, arn string) {
	manualResult, err := conn.Register(ctx, common.SubscriptionRegistrationParams{
		Request: &salesforce.RegistrationParams{
			UniqueRef:             "AmpManual" + uniqueString,
			Label:                 "AmpManual" + uniqueString,
			AwsNamedCredentialArn: arn,
		},
	})
	if err != nil {
		logging.Logger(ctx).Error("Error registering (manual flow)", "error", err)

		return
	}

	fmt.Println("Registration result (manual flow):", prettyPrint(manualResult))

	manualSubscribeParams := common.SubscribeParams{
		RegistrationResult: manualResult,
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Account": {
				WatchFields: []string{"Name", "Phone"},
			},
		},
		Request: &salesforce.SubscriptionRequest{
			QuotaOptimizationObjectFields: map[common.ObjectName]string{
				"Account": "amp_cdc_optimized",
			},
			UseExistingQuotaOptimizationFields: true,
			ManualApexTriggerDeployment:        true,
		},
	}

	manualSubscribeResult, err := conn.Subscribe(ctx, manualSubscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing (manual flow)",
			"error", err, "subscribeResult", prettyPrint(manualSubscribeResult))

		return
	}

	fmt.Println("Subscribe result (manual flow):", prettyPrint(manualSubscribeResult))

	// Update: keep Account, add Contact. Both flags stay true so the connector
	// continues to skip quota field creation and apex trigger deployment for the new object too.
	manualUpdateParams := common.SubscribeParams{
		RegistrationResult: manualResult,
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Account": {
				WatchFields: []string{"Name", "Phone"},
			},
			"Contact": {
				WatchFields: []string{"Email"},
			},
		},
		Request: &salesforce.SubscriptionRequest{
			QuotaOptimizationObjectFields: map[common.ObjectName]string{
				"Account": "amp_cdc_optimized",
				"Contact": "amp_cdc_optimized",
			},
			UseExistingQuotaOptimizationFields: true,
			ManualApexTriggerDeployment:        true,
		},
	}

	manualUpdateResult, err := conn.UpdateSubscription(ctx, manualUpdateParams, manualSubscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error updating subscription (manual flow)", "error", err)

		return
	}

	fmt.Println("Update subscription result (manual flow, keep Account + add Contact):", prettyPrint(manualUpdateResult))

	// Second update: remove Account, keep Contact. With manual mode, the caller
	// owns the field/trigger lifecycle, so the connector won't try to delete them.
	manualUpdate2Params := common.SubscribeParams{
		RegistrationResult: manualResult,
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Contact": {
				WatchFields: []string{"Email"},
			},
		},
		Request: &salesforce.SubscriptionRequest{
			QuotaOptimizationObjectFields: map[common.ObjectName]string{
				"Contact": "amp_cdc_optimized",
			},
			UseExistingQuotaOptimizationFields: true,
			ManualApexTriggerDeployment:        true,
		},
	}

	manualUpdate2Result, err := conn.UpdateSubscription(ctx, manualUpdate2Params, manualUpdateResult)
	if err != nil {
		logging.Logger(ctx).Error("Error updating subscription (manual flow, remove Account)", "error", err)

		return
	}

	fmt.Println("Update subscription result (manual flow, remove Account):", prettyPrint(manualUpdate2Result))

	if manualUpdate2Result != nil && manualUpdate2Result.Status == common.SubscriptionStatusSuccess {
		if err := conn.DeleteSubscription(ctx, *manualUpdate2Result); err != nil {
			logging.Logger(ctx).Error("Error unsubscribing (manual flow)", "error", err)

			return
		}
	}

	fmt.Println("Delete subscription successful (manual flow)")

	if err := conn.DeleteRegistration(ctx, *manualResult); err != nil {
		logging.Logger(ctx).Error("Error rolling back registration (manual flow)", "error", err)

		return
	}

	fmt.Println("Delete registration successful (manual flow)")
}

// runNoOptimizationFlow exercises Subscribe with no quota optimization at all:
// no QuotaOptimizationObjectFields, so the connector skips quota field creation
// and apex trigger deployment. CDC events fire on every change to a watched object.
func runNoOptimizationFlow(ctx context.Context, conn *salesforce.Connector, uniqueString, arn string) {
	noneResult, err := conn.Register(ctx, common.SubscriptionRegistrationParams{
		Request: &salesforce.RegistrationParams{
			UniqueRef:             "AmpNone" + uniqueString,
			Label:                 "AmpNone" + uniqueString,
			AwsNamedCredentialArn: arn,
		},
	})
	if err != nil {
		logging.Logger(ctx).Error("Error registering (no-optimization flow)", "error", err)

		return
	}

	fmt.Println("Registration result (no-optimization flow):", prettyPrint(noneResult))

	noneSubscribeParams := common.SubscribeParams{
		RegistrationResult: noneResult,
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Account": {
				WatchFields: []string{"Name", "Phone"},
			},
		},
	}

	noneSubscribeResult, err := conn.Subscribe(ctx, noneSubscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing (no-optimization flow)",
			"error", err, "subscribeResult", prettyPrint(noneSubscribeResult))

		return
	}

	fmt.Println("Subscribe result (no-optimization flow):", prettyPrint(noneSubscribeResult))

	if noneSubscribeResult != nil && noneSubscribeResult.Status == common.SubscriptionStatusSuccess {
		if err := conn.DeleteSubscription(ctx, *noneSubscribeResult); err != nil {
			logging.Logger(ctx).Error("Error unsubscribing (no-optimization flow)", "error", err)

			return
		}
	}

	fmt.Println("Delete subscription successful (no-optimization flow)")

	if err := conn.DeleteRegistration(ctx, *noneResult); err != nil {
		logging.Logger(ctx).Error("Error rolling back registration (no-optimization flow)", "error", err)

		return
	}

	fmt.Println("Delete registration successful (no-optimization flow)")
}

func prettyPrint(v any) string {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}
