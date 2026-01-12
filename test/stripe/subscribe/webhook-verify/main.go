package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/stripe"
	connTest "github.com/amp-labs/connectors/test/stripe"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetStripeConnector(ctx)

	// The webhookURL must be a publicly accessible URL that Stripe can reach.
	// Since this server runs locally on port 8080, use a tunneling tool like ngrok
	// to expose the local server to the internet. Example:
	//   1. Start ngrok: ngrok http 8080
	//   2. Copy the ngrok HTTPS URL (e.g., https://abc123.ngrok.io)
	//   3. Set this URL in the credentials JSON file as 'webhookURL'
	//   4. Ensure the URL ends with the path you want (e.g., https://abc123.ngrok.io/)
	//   5. trigger a webhook event
	webhookURLField := credscanning.Field{
		Name:      "webhookURL",
		PathJSON:  "webhookURL",
		SuffixENV: "WEBHOOK_URL",
	}
	filePath := credscanning.LoadPath(providers.Stripe)
	reader := utils.MustCreateProvCredJSON(filePath, false, webhookURLField)
	webhookURL := reader.Get(webhookURLField)

	if webhookURL == "" {
		slog.Error("Webhook URL is required. Add 'webhookURL' field to stripe credentials JSON file")
		os.Exit(1)
	}

	subscribeParams := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"account": {
				Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
				PassThroughEvents: []string{"account.application.authorized", "account.application.deauthorized"},
			},
			"charge": {
				PassThroughEvents: []string{"charge.succeeded"},
			},
			"payment_intent": {
				PassThroughEvents: []string{"payment_intent.succeeded"},
			},
		},
		Request: &stripe.SubscriptionRequest{
			WebhookEndPoint: webhookURL,
		},
	}

	slog.Info("Creating subscription...")
	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing", "error", err)
		return
	}

	if subscribeResult == nil {
		slog.Error("Subscription result is nil")
		return
	}

	stripeResult, ok := subscribeResult.Result.(*stripe.SubscriptionResult)
	if !ok || stripeResult == nil {
		slog.Error("Invalid subscription result type")
		return
	}

	if len(stripeResult.Subscriptions) == 0 {
		slog.Error("No subscriptions created")
		return
	}
	var webhookSecret string
	for _, endpoint := range stripeResult.Subscriptions {
		if endpoint.Secret != "" {
			webhookSecret = endpoint.Secret
			break
		}
	}
	if webhookSecret == "" {
		slog.Error("No webhook secret found in subscription result")
		return
	}

	slog.Info("Webhook server ready")

	verificationParams := &common.VerificationParams{
		Param: &stripe.VerificationParams{
			Secret: webhookSecret,
		},
	}

	handler := testscenario.CreateWebhookHandler(ctx, conn, verificationParams, func(body []byte) error {
		slog.Info("Webhook verified and processed", "bodySize", len(body))
		return nil
	})

	http.HandleFunc("/", handler)

	slog.Info("Starting webhook server on :8080/")
	slog.Info("Press Ctrl+C to stop")

	server := &http.Server{Addr: ":8080"}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()

	<-ctx.Done()
	_ = server.Shutdown(context.Background())

	slog.Info("Cleaning up subscription...")
	if err := conn.DeleteSubscription(context.Background(), *subscribeResult); err != nil {
		slog.Error("Error deleting subscription", "error", err)
	} else {
		slog.Info("Subscription deleted")
	}
}
