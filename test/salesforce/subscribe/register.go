package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
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

	conn := connTest.GetSalesforceConnector(ctx)

	uniqueString := strconv.Itoa(int(time.Now().UnixMilli()))

	arn := os.Getenv("AWS_NAMED_CRED_ARN")

	params := &common.SubscriptionRegistrationParams{
		Request: &salesforce.RegistrationParams{
			UniqueRef: "Amp" + uniqueString,
			Label:     "Amp" + uniqueString,
			AwsArn:    arn,
		},
	}

	result, err := conn.Register(ctx, params)
	if err != nil {
		slog.Error("Error registering", "error", err)
		return
	}

	fmt.Println("Registration result:", prettyPrint(result))

	if err := conn.RollbackRegister(ctx, result.Result.(*salesforce.ResultData)); err != nil {
		slog.Error("Error rolling back registration", "error", err)

		return
	}

	fmt.Println("Rollback successful")
}

func prettyPrint(v any) string {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}
