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

	var uniqueString string
	var namedCredArn string

	flag.StringVar(&uniqueString, "unique", "", "Unique string to append to the registration label")
	flag.StringVar(&namedCredArn, "arn", "", "AWS Named Credential ARN")
	flag.Parse()

	if uniqueString == "" || namedCredArn == "" {
		log.Fatalf("missing flags: go run <your path>.go -unique <unique string> -arn <AWS Named Credential ARN>")
	}

	conn := connTest.GetSalesforceConnector(ctx)

	arn := namedCredArn

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
			"Account": {},
		},
	}

	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing", "error", err, "subscribeResult", prettyPrint(subscribeResult))
	}

	fmt.Println("Subscribe result:", prettyPrint(subscribeResult))

	updateParams := common.SubscribeParams{
		RegistrationResult: result,
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"Contact": {},
		},
	}

	updateResult, err := conn.UpdateSubscription(ctx, updateParams, subscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error updating subscription", "error", err)
	}

	fmt.Println("Update subscription result:", prettyPrint(updateResult))

	if updateResult != nil && updateResult.Status == common.SubscriptionStatusSuccess {
		if err := conn.DeleteSubscription(ctx, *updateResult); err != nil {
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

func prettyPrint(v any) string {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}
