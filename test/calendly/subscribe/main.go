package main

import (
	"context"
	"flag"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/calendly"
	calendlytest "github.com/amp-labs/connectors/test/calendly"
	"github.com/amp-labs/connectors/test/utils"
)

var (
	callbackURL  = flag.String("callback", "https://example.com/webhook", "Webhook callback URL")
	scope        = flag.String("scope", "organization", "Subscription scope (user or organization)")
	organization = flag.String("org", "", "Organization URI (required for organization scope)")
	user         = flag.String("user", "", "User URI (required for user scope)")
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	flag.Parse()

	// Create connector using the helper function
	connector := calendlytest.GetConnector(ctx)

	// Prepare subscription request
	subscriptionRequest := &calendly.CalendlySubscriptionRequest{
		URL:    *callbackURL,
		Events: []calendly.CalendlySubscriptionEvent{
			calendly.CalendlyEventInviteeCreated,
			calendly.CalendlyEventInviteeCanceled,
		},
		Scope:        *scope,
		Organization: *organization,
		User:         *user,
	}

	// Create subscription parameters
	subscriptionParams := common.SubscribeParams{
		Request: subscriptionRequest,
	}

	// Call Subscribe
	result, err := connector.Subscribe(ctx, subscriptionParams)
	if err != nil {
		slog.Error("Failed to create subscription", "error", err)
		return
	}

	slog.Info("Subscription created successfully", "status", result.Status)
} 