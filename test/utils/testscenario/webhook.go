package testscenario

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// CreateWebhookHandler creates an HTTP handler function that verifies incoming webhook requests
// using the connector's VerifyWebhookMessage method. After successful verification, the handler
// invokes the provided callback function with the verified request body for further processing.
//
// The handler reads the request body, constructs a WebhookRequest from the HTTP request,
// verifies the webhook signature using the connector's verification method, and on success
// calls the callback with the raw body. If verification fails, the handler returns appropriate
// HTTP error responses.
//
// Parameters:
//   - ctx: context for the handler operations
//   - conn: connector implementing WebhookVerifierConnector interface
//   - verificationParams: provider-specific verification parameters (e.g., secrets)
//   - onVerified: callback function invoked with the verified request body on successful verification
//
// Returns an http.HandlerFunc that can be registered with an HTTP server.
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
