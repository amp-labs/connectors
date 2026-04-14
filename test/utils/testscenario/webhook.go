package testscenario

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

// createWebhookHandler creates an HTTP handler function that verifies incoming webhook requests
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
func createWebhookHandler(
	ctx context.Context,
	conn components.WebhookMessageVerifier,
	router WebhookRouter,
	verificationParams *common.VerificationParams,
	messageChannel chan webhookMessageResult,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Select best handler.
		for _, route := range router.Routes {
			if route.Left(r) {
				// Invoke handler.
				route.Right(w, r)
				return
			}
		}

		// DEFAULT route handling.
		defaultWebhookHandling(w, r, conn, ctx, verificationParams, messageChannel)
	}
}

type webhookMessageResult struct {
	Body  []byte
	Error string
}

func defaultWebhookHandling(
	w http.ResponseWriter, r *http.Request,
	conn components.WebhookMessageVerifier, ctx context.Context, verificationParams *common.VerificationParams,
	messageChannel chan webhookMessageResult,
) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		messageChannel <- webhookMessageResult{
			Error: fmt.Sprintf("error reading request body %v", err),
		}

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
		http.Error(w, "Verification failed", http.StatusUnauthorized)
		messageChannel <- webhookMessageResult{
			Error: fmt.Sprintf("VerifyWebhookMessage failed %v", err),
		}

		return
	}

	if !valid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		messageChannel <- webhookMessageResult{
			Error: "according to VerifyWebhookMessage the message is invalid",
		}

		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
	messageChannel <- webhookMessageResult{
		Body: body,
	}
}
