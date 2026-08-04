package webhook

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

const (
	EnvArgWebhookURL   = "WEBHOOK_URL"
	HandlerDefaultPort = "4550"
)

// Server runs an HTTP server in the background to receive webhook events.
type Server struct {
	Handler http.HandlerFunc
}

// CreateServer creates a Server that listens for incoming webhooks.
// It uses the connector to verify the message and then passes the event to the Processor.
func CreateServer(ctx context.Context,
	processor *Processor,
	conn testconn.TestableWebhookMessageVerifier,
	verificationParams *common.VerificationParams,
) Server {
	if processor.channel == nil {
		processor.channel = make(chan Event)
	}

	return Server{
		Handler: func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("failed to read body", "error", err)
				http.Error(w, "Error reading body", http.StatusBadRequest)
				return
			}

			if processor.Interceptor != nil {
				if done := processor.Interceptor(w, r, bodyBytes); done {
					return
				}
			}

			message := createWebhookEvent(ctx, w, r, conn, bodyBytes, verificationParams)

			// Each incoming event will produce a goroutine waiting for the event to be consumed or unless
			// the context is canceled indicating the server stopped.
			// This will release the HTTP request/response just in case.
			go processor.processEvent(ctx, message)
		},
	}
}

// Start starts server on the localhost that should be exposed via ngrok
// to receive webhook messages.
//
// Webhook handler will send WebhookEvent using go channel.
func (h Server) Start(ctx context.Context) (string, func()) {
	// Main server loop.
	var webhookCancel context.CancelFunc
	ctx, webhookCancel = context.WithCancel(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Handler)

	// Construct and start server.
	fmt.Printf("Starting webhook server on :%v\n", HandlerDefaultPort)
	server := &http.Server{
		Addr:    ":" + HandlerDefaultPort,
		Handler: mux,
	}

	fmt.Printf("Run ngrok with command: `ngrok http %v`\n", HandlerDefaultPort)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", "error", err)
		}
	}()

	shutdown := func() {
		fmt.Println("=> Shutting down webhook handler")
		webhookCancel()
		_ = server.Shutdown(context.Background())
		fmt.Println("=> completed")
	}

	return "http://localhost:" + HandlerDefaultPort, shutdown
}
