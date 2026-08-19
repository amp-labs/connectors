package subscribe

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// mockSalesloftTaskPayload is a trimmed copy of the example webhook payload documented in
// providers/salesloft/subscribeEvent.go. Salesloft delivers one record per webhook as a bare
// JSON object; the event name travels in the x-salesloft-event request header.
const mockSalesloftTaskPayload = `{
  "created_at": "2025-11-20T07:21:04.176528-05:00",
  "current_state": "scheduled",
  "description": "noted updated ",
  "due_date": "2025-11-27",
  "id": 693873531,
  "person": {
    "_href": "https://api.salesloft.com/v2/people/436664215",
    "id": 436664215
  },
  "subject": "Follow-up with John Kelly",
  "task_type": "general",
  "updated_at": "2025-11-20T07:21:16.193200-05:00"
}`

// TestMockSalesloftConfigResolution verifies the mocksalesloft registry entry resolves through
// GetProviderConfig exactly like Salesloft's: verification bypassed, the WatchFieldsAuto quirk
// declared, and subscribe-via-API supported.
func TestMockSalesloftConfigResolution(t *testing.T) { //nolint:paralleltest // mutates the provider catalog
	providers.SetupMockSalesloftProvider()

	info, err := providers.ReadInfo(providers.MockSalesloft)
	if err != nil {
		t.Fatalf("ReadInfo: %v", err)
	}

	cfg, err := GetProviderConfig("", ResolveProviderInfoAlias(info), deps.Dependencies{})
	if err != nil {
		t.Fatalf("GetProviderConfig: %v", err)
	}

	if !cfg.Verification.ShouldBypass() {
		t.Error("expected mocksalesloft webhook verification to be bypassed")
	}

	if !cfg.Subscription.RequiresWatchFieldsAutoAll() {
		t.Error("expected mocksalesloft to mirror Salesloft's WatchFieldsAuto quirk")
	}

	if !cfg.Subscription.IsSupportedViaAPI() {
		t.Error("expected mocksalesloft to mirror Salesloft's subscribe-by-API support")
	}
}

// TestMockSalesloftEventParsing runs a Salesloft-shaped webhook payload through the same
// object-payload dispatch the server's receive path uses and asserts the real Salesloft event
// implementation classifies it: the mock provider name must parse events identically to
// Salesloft itself.
func TestMockSalesloftEventParsing(t *testing.T) {
	t.Parallel()

	var rawEvent map[string]any
	if err := json.Unmarshal([]byte(mockSalesloftTaskPayload), &rawEvent); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	events, err := GetObjectTypeSubscribeEventsList(providers.MockSalesloft, rawEvent)
	if err != nil {
		t.Fatalf("GetObjectTypeSubscribeEventsList: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected Salesloft's one-record-per-webhook shape, got %d events", len(events))
	}

	event := events[0]

	headers := http.Header{}
	headers.Set("x-salesloft-event", "task_updated")

	if err := event.PreLoadData(&common.SubscriptionEventPreLoadData{RequestHeaders: &headers}); err != nil {
		t.Fatalf("PreLoadData: %v", err)
	}

	eventType, err := event.EventType()
	if err != nil {
		t.Fatalf("EventType: %v", err)
	}

	if eventType != common.SubscriptionEventTypeUpdate {
		t.Errorf("expected update event, got %q", eventType)
	}

	objectName, err := event.ObjectName()
	if err != nil {
		t.Fatalf("ObjectName: %v", err)
	}

	if objectName != "tasks" {
		t.Errorf("expected object name %q, got %q", "tasks", objectName)
	}

	recordID, err := event.RecordId()
	if err != nil {
		t.Fatalf("RecordId: %v", err)
	}

	if recordID != "693873531" {
		t.Errorf("expected record id %q, got %q", "693873531", recordID)
	}

	rawEventName, err := event.RawEventName()
	if err != nil {
		t.Fatalf("RawEventName: %v", err)
	}

	if rawEventName != "task_updated" {
		t.Errorf("expected raw event name %q, got %q", "task_updated", rawEventName)
	}
}
