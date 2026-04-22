package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	housecallpro "github.com/amp-labs/connectors/providers/housecallPro"
	connTest "github.com/amp-labs/connectors/test/housecallPro"
	"github.com/amp-labs/connectors/test/utils"
)

const (
	serverAddr = ":8080"
	// envWebhookSecret is the signing secret from Housecall Pro webhook settings (not the API key).
	envWebhookSecret = "HOUSECALL_PRO_WEBHOOK_SECRET"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)

	// The webhook endpoint must be publicly reachable by Housecall Pro.
	// Since this server runs locally on port 8080, use a tunneling tool like ngrok
	// to expose the local server to the internet. Example:
	//   1. Start ngrok: ngrok http 8080
	//   2. Copy the ngrok HTTPS URL (for example: https://abc123.ngrok.io)
	//   3. Configure the URL in Housecall Pro webhook settings
	//   4. Trigger webhook events from Housecall Pro
	//
	// verify signature -> parse webhook -> log parsed events.
	secret, ok := loadWebhookSecret()
	if !ok {
		slog.Error("Webhook signing secret is required", "env", envWebhookSecret, "hint", "export HOUSECALL_PRO_WEBHOOK_SECRET=<secret from Housecall Pro webhook settings>")
		os.Exit(1)
	}

	verificationParams := &common.VerificationParams{
		Param: &housecallpro.HousecallProVerificationParams{
			Secret: secret,
		},
	}

	handler := CreateWebhookHandler(ctx, conn, verificationParams, func(body []byte) error {
		var collapsed housecallpro.CollapsedSubscriptionEvent
		if err := json.Unmarshal(body, &collapsed); err != nil {
			return err
		}

		events, err := collapsed.SubscriptionEventList()
		if err != nil {
			return err
		}

		for _, evt := range events {
			slog.Info("Webhook verified and parsed")
			rawName, err := evt.RawEventName()
			if err != nil {
				slog.Error("Failed parsing RawEventName", "error", err)
			}
			slog.Info("RawEventName", "rawName", rawName)

			objectName, err := evt.ObjectName()
			if err != nil {
				slog.Error("Failed parsing ObjectName", "error", err)
			}
			slog.Info("ObjectName", "objectName", objectName)
			recordID, err := evt.RecordId()
			if err != nil {
				slog.Error("Failed parsing RecordId", "error", err)
			}
			slog.Info("RecordId", "recordID", recordID)
			eventType, err := evt.EventType()
			if err != nil {
				slog.Error("Failed parsing EventType", "error", err)
			}
			slog.Info("EventType", "eventType", eventType)
			timestamp, err := evt.EventTimeStampNano()
			if err != nil {
				slog.Error("Failed parsing EventTimeStampNano", "error", err)
			}
			slog.Info("EventTimeStampNano", "timestamp", timestamp)
			workspace, err := evt.Workspace()
			if err != nil {
				slog.Error("Failed parsing Workspace", "error", err)
			}
			slog.Info("Workspace", "workspace", workspace)
			rawMap, err := evt.RawMap()
			if err != nil {
				slog.Error("Failed parsing RawMap", "error", err)
			}
			slog.Info("RawMap", "rawMap", rawMap)

		}

		return nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	server := &http.Server{
		Addr:              serverAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	slog.Info("Starting Housecall webhook server", "addr", serverAddr)
	slog.Info("Press Ctrl+C to stop")

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()

	<-ctx.Done()

	// Graceful shutdown of local webhook server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = server.Shutdown(shutdownCtx)
	slog.Info("Webhook server stopped")
}

func loadWebhookSecret() (string, bool) {
	secret := strings.TrimSpace(os.Getenv(envWebhookSecret))

	return secret, secret != ""
}
func CreateWebhookHandler(
	ctx context.Context,
	conn connectors.WebhookVerifierConnector,
	verificationParams *common.VerificationParams,
	onVerified func(body []byte) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Error reading request body", "error", err)
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}

		request := &common.WebhookRequest{
			Headers: r.Header,
			Body:    body,
			URL:     r.URL.String(),
			Method:  r.Method,
		}

		valid, err := conn.VerifyWebhookMessage(ctx, request, verificationParams)
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

		if onVerified != nil {
			if err := onVerified(body); err != nil {
				slog.Error("Error processing verified webhook", "error", err)
				http.Error(w, "Error processing webhook", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}
}
