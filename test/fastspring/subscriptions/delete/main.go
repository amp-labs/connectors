package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/fastspring"
	connTest "github.com/amp-labs/connectors/test/fastspring"
	"github.com/amp-labs/connectors/test/utils"
)

// Cancels a subscription (DELETE /subscriptions/{id}). This is destructive.
// Set FASTSPRING_SUBSCRIPTION_ID to a disposable test subscription id before running.
const envSubscriptionID = "FASTSPRING_SUBSCRIPTION_ID"

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	subID := os.Getenv(envSubscriptionID)
	if subID == "" {
		slog.Info("Skipping: set FASTSPRING_SUBSCRIPTION_ID to a test subscription to cancel",
			"env", envSubscriptionID)
		return
	}

	conn := connTest.GetFastSpringConnector(ctx)

	slog.Warn("=== Cancelling subscription (DELETE) ===", "subscription", subID)

	if err := cancelSubscription(ctx, conn, subID); err != nil {
		slog.Error("Failed to cancel subscription", "error", err)
		os.Exit(1)
	}

	slog.Info("Subscription delete completed")
}

func cancelSubscription(ctx context.Context, conn *fastspring.Connector, subscriptionID string) error {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "subscriptions",
		RecordId:   subscriptionID,
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	if !res.Success {
		return fmt.Errorf("delete reported Success=false")
	}

	return nil
}
