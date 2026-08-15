package webhook

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

const (
	EnvArgWebhookURL   = "WEBHOOK_URL"
	HandlerDefaultPort = "4550"
)

// Server runs an HTTP server in the background to receive webhook events.
type Server struct {
	handler   http.HandlerFunc
	processor *Processor

	verificationMu    sync.RWMutex
	verificationState VerificationState
}

type VerificationState struct {
	Enabled   bool
	Params    *common.VerificationParams
	Connector testconn.TestableWebhookMessageVerifier
}

// CreateServer creates a Server that listens for incoming webhooks.
// It uses the connector to verify the message and then passes the event to the Processor.
func CreateServer(ctx context.Context, processor *Processor) *Server {
	if processor.channel == nil {
		processor.channel = make(chan Event)
	}

	server := Server{
		processor:      processor,
		verificationMu: sync.RWMutex{},
		verificationState: VerificationState{
			Enabled:   false,
			Params:    nil,
			Connector: nil,
		},
	}

	server.handler = func(w http.ResponseWriter, r *http.Request) {
		server.handle(ctx, w, r)
	}

	return &server
}

// Start starts server on the localhost that should be exposed via ngrok
// to receive webhook messages.
//
// Webhook handler will send WebhookEvent using go channel.
func (s *Server) Start(ctx context.Context) (string, func()) {
	// Main server loop.
	var webhookCancel context.CancelFunc
	ctx, webhookCancel = context.WithCancel(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handler)

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

func (s *Server) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read body", "error", err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	// STAGE (1) -- custom interceptor.
	if s.processor.Interceptor != nil {
		if done := s.processor.Interceptor(w, r, bodyBytes); done {
			return
		}
	}

	// STAGE (2) -- connector message verification.
	verification := s.verification()
	if !verification.Enabled {
		// The interceptor did not consume the request, but connector
		// verification is not active yet.
		fmt.Println("webhook verification is not enabled, invoke server.SetupConnectorVerification first")
		http.Error(w, "webhook verification is not enabled", http.StatusInternalServerError)
		return
	}

	message := createWebhookEvent(ctx, w, r, verification.Connector, bodyBytes, verification.Params)

	// Each incoming event will produce a goroutine waiting for the event to be consumed or unless
	// the context is canceled indicating the server stopped.
	// This will release the HTTP request/response just in case.
	go s.processor.processEvent(ctx, message)
}

func (s *Server) SetupConnectorVerification(
	conn testconn.TestableWebhookMessageVerifier,
	params *common.VerificationParams,
) {
	s.verificationMu.Lock()
	defer s.verificationMu.Unlock()

	s.verificationState = VerificationState{
		Enabled:   true,
		Params:    params,
		Connector: conn,
	}
}

func (s *Server) verification() VerificationState {
	s.verificationMu.RLock()
	defer s.verificationMu.RUnlock()

	return s.verificationState
}
