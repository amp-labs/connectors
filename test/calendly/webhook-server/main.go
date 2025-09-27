package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/amp-labs/connectors/providers/calendly"
	calendlytest "github.com/amp-labs/connectors/test/calendly"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	utils.SetupLogging()
	ctx := context.Background()
	connector := calendlytest.GetConnector(ctx)

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		handleCalendlyWebhook(ctx, w, r, connector)
	})

	slog.Info("Starting webhook server on http://localhost:8080/webhook")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("Server failed", "error", err)
	}
}

func handleCalendlyWebhook(ctx context.Context, w http.ResponseWriter, r *http.Request, connector *calendly.Connector) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		http.Error(w, "Bad request", 400)
		return
	}

	var event calendly.CalendlyWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		slog.Error("Failed to parse webhook payload", "error", err)
		http.Error(w, "Bad request", 400)
		return
	}

	eventType, _ := event["event"].(string)
	var payloadURI string
	if payload, ok := event["payload"].(map[string]any); ok {
		if uri, ok := payload["uri"].(string); ok {
			payloadURI = uri
		}
	}

	slog.Info("Webhook event", "type", eventType, "uri", payloadURI)

	w.WriteHeader(200)
	fmt.Fprintf(w, "Webhook received successfully")
} 