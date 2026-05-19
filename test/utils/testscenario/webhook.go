package testscenario

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

const (
	EnvArgWebhookURL          = "WEBHOOK_URL"
	WebhookHandlerDefaultPort = "4550"
)

// getPublicWebhookURL loads webhook URL from the environment variable or from the standard user input.
func getPublicWebhookURL(ctx context.Context) (url string, ok bool) {
	defer func() {
		if ok {
			fmt.Printf("Webhook URL: \"%v\"\n", url)
		}
	}()

	url, ok = os.LookupEnv(EnvArgWebhookURL)
	if !ok {
		fmt.Printf("Env variable is missing \"%v\"\n", EnvArgWebhookURL)
		url, ok = waitForWebhookURLInput(ctx)
	}

	return url, ok
}

func waitForWebhookURLInput(ctx context.Context) (string, bool) {
	fmt.Println("Please provide the public URL (e.g., from ngrok) that tunnels to this local server.")
	fmt.Print("Public Webhook URL (empty string to cancel): ")

	inputCh := make(chan string)
	errCh := make(chan error)

	// Routine waiting for standard input.
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			inputCh <- strings.TrimSpace(scanner.Text())
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Printf("\nContext cancelled while waiting for webhook URL input.\n")
		return "", false
	case err := <-errCh:
		printError(fmt.Errorf("failed to read public URL: %w", err))
		return "", false
	case publicURL := <-inputCh:
		if publicURL == "" {
			fmt.Println("Empty input for webhook URL: stopping script.")
			return "", false
		}

		if !isValidHTTPS(publicURL) {
			fmt.Printf("Invalid URL format: %v\n", publicURL)
			return "", false
		}

		// proceed normally
		return publicURL, true
	}
}

// startWebhookHandler starts server on the localhost that should be exposed via ngrok
// to receive webhook messages.
//
// Webhook handler will send webhookMessageResult using go channel.
func startWebhookHandler(
	ctx context.Context, conn ConnectorWebhookSubscriber,
	router WebhookRouter,
	verificationParams *common.VerificationParams,
	messageChannel chan webhookMessageResult,
) (string, func()) {
	// Main server loop.
	var webhookCancel context.CancelFunc
	ctx, webhookCancel = context.WithCancel(ctx)
	webhookHandler := createWebhookHandler(ctx, conn, router, verificationParams, messageChannel)

	mux := http.NewServeMux()
	mux.HandleFunc("/", webhookHandler)

	// Construct and start server.
	fmt.Printf("Starting webhook server on :%v\n", WebhookHandlerDefaultPort)
	server := &http.Server{
		Addr:    ":" + WebhookHandlerDefaultPort,
		Handler: mux,
	}

	fmt.Printf("Run ngrok with command: `ngrok http %v`\n", WebhookHandlerDefaultPort)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()

	shutdown := func() {
		fmt.Println("=> Shutting down webhook handler")
		webhookCancel()
		_ = server.Shutdown(context.Background())
		fmt.Println("=> completed")
	}

	return "http://localhost:" + WebhookHandlerDefaultPort, shutdown
}

// webhookMessageResult represents a message sent by defaultWebhookHandler.
// For a successful provider request to a webhook a Body will be populated.
// In case of any unexpected error in code or from the request itself an Error will be non-empty.
type webhookMessageResult struct {
	Body  []byte
	Error string
}

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
		// Select best custom webhook handler.
		for _, route := range router.Routes {
			if matches := route.Left(r); matches {
				// Invoke handler.
				route.Right(w, r)
				return
			}
		}

		// DEFAULT route handling.
		defaultWebhookHandler(w, r, conn, ctx, verificationParams, messageChannel)
	}
}

func defaultWebhookHandler(
	w http.ResponseWriter, r *http.Request,
	conn components.WebhookMessageVerifier, ctx context.Context, verificationParams *common.VerificationParams,
	messageChannel chan webhookMessageResult,
) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		sendMessage(ctx, messageChannel, webhookMessageResult{
			Error: fmt.Sprintf("error reading request body %v", err),
		})

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
		sendMessage(ctx, messageChannel, webhookMessageResult{
			Error: fmt.Sprintf("VerifyWebhookMessage failed %v", err),
		})

		return
	}

	if !valid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		sendMessage(ctx, messageChannel, webhookMessageResult{
			Error: "according to VerifyWebhookMessage the message is invalid",
		})

		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
	sendMessage(ctx, messageChannel, webhookMessageResult{
		Body: body,
	})
}

// sendMessage sends a webhookMessageResult to the channel if possible,
// or discards it if the context is canceled or done.
//
// This is useful when the provider may still be sending events while
// the server is shutting down (for example, because the expected number
// of webhook events has arrived, or because a developer canceled the test script).
// In that case, the send will not block the handler; instead, the message is silently dropped.
func sendMessage(ctx context.Context, channel chan webhookMessageResult, message webhookMessageResult) {
	select {
	case channel <- message:
	case <-ctx.Done():
		// We don't care if we couldn't send.
		// The server is shutting down.
	}
}

func isValidHTTPS(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}

	if u.Scheme != "https" {
		return false
	}

	if u.Host == "" {
		return false
	}

	return true
}
