package webhook

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

// Event represents a verified webhook message or an error during verification.
type Event struct {
	Body  []byte
	Error string
}

// createWebhookEvent uses the connector to verify the incoming HTTP request and returns an Event.
func createWebhookEvent(ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	conn testconn.TestableWebhookMessageVerifier,
	body []byte,
	verificationParams *common.VerificationParams,
) Event {
	request := &common.WebhookRequest{
		Headers: r.Header,
		Body:    body,
		URL:     r.URL.String(),
		Method:  r.Method,
	}

	valid, err := conn.VerifyWebhookMessage(ctx, request, verificationParams)
	if err != nil {
		http.Error(w, "Verification failed", http.StatusUnauthorized)

		return Event{
			Error: fmt.Sprintf("VerifyWebhookMessage failed %v", err),
		}
	}

	if !valid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)

		return Event{
			Error: "according to VerifyWebhookMessage the message is invalid",
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))

	return Event{
		Body: body,
	}
}
