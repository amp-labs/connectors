package subscribe

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/reply"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// replyDelivery loads a real (sanitized) Reply webhook delivery captured for the connector tests.
func replyDelivery(t *testing.T, name string) map[string]any {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "providers", "reply", "test", "subscribe", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}

	var delivery map[string]any
	if err := json.Unmarshal(raw, &delivery); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}

	return delivery
}

// TestReplyConfig pins the declared shape: subscribes via API, verification bypassed (Reply sends no
// signature), and none of Registration, Maintenance or PostProcess.
func TestReplyConfig(t *testing.T) {
	t.Parallel()

	cfg, err := GetProviderConfig("", makeProviderInfo(providers.Reply, ""), deps.Dependencies{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Subscription.IsSupportedViaAPI() {
		t.Error("Subscription.IsSupportedViaAPI() = false, want true")
	}

	if cfg.Subscription.SubscribeManually() {
		t.Error("Subscription.SubscribeManually() = true, want false")
	}

	if cfg.Subscription.RequiresWatchFieldsAutoAll() {
		t.Error("Subscription.RequiresWatchFieldsAutoAll() = true, want false")
	}

	// Reply has no shared resource to register and documents no subscription expiry.
	if cfg.Registration.IsRequired(&reply.Connector{}) {
		t.Error("Registration.IsRequired() = true, want false")
	}

	if cfg.Maintenance.ShouldPerform() {
		t.Error("Maintenance.ShouldPerform() = true, want false")
	}

	if _, ok := cfg.Maintenance.Interval(); ok {
		t.Error("Maintenance.Interval() declared, want none")
	}

	if cfg.PostProcess.ShouldPerform() {
		t.Error("PostProcess.ShouldPerform() = true, want false")
	}

	// Reply sends no signature, so verification is skipped.
	if !cfg.Verification.ShouldBypass() {
		t.Error("Verification.ShouldBypass() = false, want true")
	}

	verifier, err := cfg.Verification.Connector(context.Background())
	if err != nil {
		t.Fatalf("Verification.Connector() error = %v", err)
	}

	if _, ok := verifier.(*reply.Connector); !ok {
		t.Errorf("Verification.Connector() = %T, want *reply.Connector", verifier)
	}
}

// TestReplyBuildRequest verifies the subscribe-time payload carries the caller-constructed webhook URL.
func TestReplyBuildRequest(t *testing.T) {
	t.Parallel()

	cfg, err := GetProviderConfig("", makeProviderInfo(providers.Reply, ""), deps.Dependencies{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := cfg.Subscription.BuildRequest(
		context.Background(), &openapi.Installation{Id: "inst-1"}, nil, nil, "https://example.com/hooks/reply",
	)
	if err != nil {
		t.Fatalf("BuildRequest() error = %v", err)
	}

	request, ok := got.(*reply.SubscriptionRequest)
	if !ok {
		t.Fatalf("BuildRequest() = %T, want *reply.SubscriptionRequest", got)
	}

	if request.WebhookURL != "https://example.com/hooks/reply" {
		t.Errorf("WebhookURL = %q, want the caller-built URL", request.WebhookURL)
	}
}

// TestReplyCastEvents verifies a raw delivery is cast into a typed Reply SubscriptionEvent.
func TestReplyCastEvents(t *testing.T) {
	t.Parallel()

	cfg, err := GetProviderConfig("", makeProviderInfo(providers.Reply, ""), deps.Dependencies{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events, err := cfg.Verification.CastEvents([]map[string]any{replyDelivery(t, "email_replied.json")})
	if err != nil {
		t.Fatalf("CastEvents() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("CastEvents() len = %d, want 1", len(events))
	}

	name, err := events[0].RawEventName()
	if err != nil || name != "email_replied" {
		t.Errorf("RawEventName() = %q, %v; want email_replied", name, err)
	}
}

// TestReplyGetObjectTypeSubscribeEventsList drives captured deliveries through the provider dispatch
// in GetObjectTypeSubscribeEventsList, so a missing Reply case is caught here rather than at enable time.
func TestReplyGetObjectTypeSubscribeEventsList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		fixture    string
		wantEvent  string
		wantObject string
	}{
		{"email_replied.json", "email_replied", "contacts"},
		{"contact_opted_out.json", "contact_opted_out", "contacts"},
		{"email_account_error.json", "email_account_error", "email-accounts"},
		{"linkedin_account_alerts.json", "linkedin_account_alerts", "linkedin-accounts"},
		{"autopilot_stopped.json", "autopilot_stopped", "sequences"},
	}

	for _, testCase := range tests {
		t.Run(testCase.fixture, func(t *testing.T) {
			t.Parallel()

			events, err := GetObjectTypeSubscribeEventsList(providers.Reply, replyDelivery(t, testCase.fixture))
			if err != nil {
				t.Fatalf("GetObjectTypeSubscribeEventsList() error = %v", err)
			}

			if len(events) != 1 {
				t.Fatalf("len(events) = %d, want 1", len(events))
			}

			assertReplyEvent(t, events[0], testCase.wantEvent, testCase.wantObject)
		})
	}
}

func assertReplyEvent(t *testing.T, event common.SubscriptionEvent, wantEvent, wantObject string) {
	t.Helper()

	name, err := event.RawEventName()
	if err != nil || name != wantEvent {
		t.Errorf("RawEventName() = %q, %v; want %q", name, err, wantEvent)
	}

	object, err := event.ObjectName()
	if err != nil || object != wantObject {
		t.Errorf("ObjectName() = %q, %v; want %q", object, err, wantObject)
	}

	recordID, err := event.RecordId()
	if err != nil || recordID == "" {
		t.Errorf("RecordId() = %q, %v; want a non-empty id", recordID, err)
	}
}
