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
	deleteFlag   = flag.Bool("delete", false, "Delete the subscription after creating it")
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()
	flag.Parse()

	connector := calendlytest.GetConnector(ctx)

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

	subscriptionParams := common.SubscribeParams{
		Request: subscriptionRequest,
	}

	result, err := connector.Subscribe(ctx, subscriptionParams)
	if err != nil {
		slog.Error("Failed to create subscription", "error", err)
		return
	}

	slog.Info("Subscription created successfully", "status", result.Status)

	if *deleteFlag && result.Result != nil {
		err = connector.DeleteSubscription(ctx, *result)
		if err != nil {
			slog.Warn("Failed to delete subscription", "error", err)
		} else {
			slog.Info("Subscription deleted successfully")
		}
	}
} 