package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/stripe"
	connTest "github.com/amp-labs/connectors/test/stripe"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetStripeConnector(ctx)

	webhookURL := ""

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

	stripeResult := subscribeResult.Result.(*stripe.SubscriptionResult)
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
	fmt.Println(webhookSecret)
	if webhookSecret == "" {
		slog.Error("No webhook secret found in subscription result")
		return
	}

	slog.Info("Webhook server ready")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Error reading request body", "error", err)
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}

		request := &common.WebhookRequest{
			Headers: r.Header,
			Body:    body,
		}

		params := &common.VerificationParams{
			Param: &stripe.StripeVerificationParams{
				Secret: webhookSecret,
			},
		}

		valid, err := conn.VerifyWebhookMessage(ctx, request, params)
		if err != nil {
			slog.Error("Verification failed", "error", err)
			http.Error(w, "Verification failed", http.StatusUnauthorized)
			return
		}

		if !valid {
			slog.Error("Invalid signature")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		var event stripe.SubscriptionEvent
		if err := json.Unmarshal(body, &event); err != nil {
			slog.Error("Error unmarshaling event", "error", err)
			http.Error(w, "Invalid event format", http.StatusBadRequest)
			return
		}

		eventType, err := event.EventType()
		if err != nil {
			slog.Error("Error getting event type", "error", err)
			http.Error(w, "Invalid event type", http.StatusBadRequest)
			return
		}

		objectName, err := event.ObjectName()
		if err != nil {
			slog.Error("Error getting object name", "error", err)
			http.Error(w, "Invalid object name", http.StatusBadRequest)
			return
		}

		recordID, err := event.RecordId()
		if err != nil {
			slog.Error("Error getting record id", "error", err)
			http.Error(w, "Invalid record id", http.StatusBadRequest)
			return
		}

		slog.Info("Webhook verified",
			"type", eventType,
			"object", objectName,
			"id", recordID,
		)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

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

	if subscribeResult != nil {
		slog.Info("Cleaning up subscription...")
		if err := conn.DeleteSubscription(context.Background(), *subscribeResult); err != nil {
			slog.Error("Error deleting subscription", "error", err)
		} else {
			slog.Info("Subscription deleted")
		}
	}
}
