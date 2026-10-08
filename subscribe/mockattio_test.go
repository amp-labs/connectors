package subscribe

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/mocksub"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// mockAttioWebhookPayload is an object-shaped Attio webhook body ({webhook_id, events}). The
// envelope mirrors the example documented in providers/attio/subscriptionEvent.go; the record.*
// id shape ({workspace_id, object_id, record_id}) follows the same file's webhook-reference
// notes (https://docs.attio.com/rest-api/webhook-reference, record-events/recordcreated).
const mockAttioWebhookPayload = `{
  "webhook_id": "04731154-70d3-42bb-8320-760304c9bbfd",
  "events": [
    {
      "event_type": "record.created",
      "id": {
        "workspace_id": "e293215c-210a-4d4a-9913-e2b33da318ab",
        "object_id": "ee1e6aa1-ec69-4ef4-a101-3a9abb12e281",
        "record_id": "9bcad14b-55a5-478d-963b-a4ec598265c6"
      },
      "actor": {
        "type": "workspace-member",
        "id": "f0519378-80b8-4d7c-8874-c6acc1850442"
      }
    },
    {
      "event_type": "note.updated",
      "id": {
        "workspace_id": "e293215c-210a-4d4a-9913-e2b33da318ab",
        "note_id": "f83d5cab-571b-47a8-8018-57146f848d19"
      },
      "actor": {
        "type": "workspace-member",
        "id": "f0519378-80b8-4d7c-8874-c6acc1850442"
      }
    }
  ]
}`

// TestMockAttioConfigResolution verifies the mockattio registry entry resolves like Attio's:
// the WatchFieldsAuto quirk and subscribe-by-API support are mirrored, with verification
// bypassed (the real config's stored-secret HMAC has no mock counterpart).
func TestMockAttioConfigResolution(t *testing.T) { //nolint:paralleltest // mutates the provider catalog
	providers.SetupMockAttioProvider()

	info, err := providers.ReadInfo(providers.MockAttio)
	if err != nil {
		t.Fatalf("ReadInfo: %v", err)
	}

	cfg, err := GetProviderConfig("", ResolveProviderInfoAlias(info), deps.Dependencies{})
	if err != nil {
		t.Fatalf("GetProviderConfig: %v", err)
	}

	if !cfg.Verification.ShouldBypass() {
		t.Error("expected mockattio webhook verification to be bypassed")
	}

	if !cfg.Subscription.RequiresWatchFieldsAutoAll() {
		t.Error("expected mockattio to mirror Attio's WatchFieldsAuto quirk")
	}

	if !cfg.Subscription.IsSupportedViaAPI() {
		t.Error("expected mockattio to mirror Attio's subscribe-by-API support")
	}
}

// TestMockAttioEventParsingAndObjectNameResolution runs an Attio-shaped webhook body through
// the same object-payload dispatch the server's receive path uses (Attio's real
// CollapsedSubscriptionEvent fan-out), then resolves object names through the mock connector:
// record.* events answer from the store's seeded id.object_id index (standing in for Attio's
// API-backed resolver), and core-object events fall back to the event's own object name.
func TestMockAttioEventParsingAndObjectNameResolution(t *testing.T) { //nolint:funlen
	t.Parallel()

	var rawEvent map[string]any
	if err := json.Unmarshal([]byte(mockAttioWebhookPayload), &rawEvent); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	events, err := GetObjectTypeSubscribeEventsList(providers.MockAttio, rawEvent)
	if err != nil {
		t.Fatalf("GetObjectTypeSubscribeEventsList: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected the events array to fan out into 2 events, got %d", len(events))
	}

	recordCreated, noteUpdated := events[0], events[1]

	if evtType, err := recordCreated.EventType(); err != nil || evtType != common.SubscriptionEventTypeCreate {
		t.Errorf("expected create event, got %q (err %v)", evtType, err)
	}

	if recordID, err := recordCreated.RecordId(); err != nil || recordID != "9bcad14b-55a5-478d-963b-a4ec598265c6" {
		t.Errorf("unexpected record id %q (err %v)", recordID, err)
	}

	if workspace, err := recordCreated.Workspace(); err != nil || workspace != "e293215c-210a-4d4a-9913-e2b33da318ab" {
		t.Errorf("unexpected workspace %q (err %v)", workspace, err)
	}

	store := mocksub.NewStore()
	store.SeedObjectName("ee1e6aa1-ec69-4ef4-a101-3a9abb12e281", "people")

	conn := mocksub.NewConnector(
		providers.MockAttio,
		mocksub.WithStore(store),
		mocksub.WithObjectNameFromEvent(mocksub.ObjectIDIndexResolver(store)),
	)

	name, err := conn.GetObjectNameFromEvent(context.Background(), recordCreated)
	if err != nil {
		t.Fatalf("GetObjectNameFromEvent (record.*): %v", err)
	}

	if name != "people" {
		t.Errorf("expected seeded object-name index to resolve %q, got %q", "people", name)
	}

	name, err = conn.GetObjectNameFromEvent(context.Background(), noteUpdated)
	if err != nil {
		t.Fatalf("GetObjectNameFromEvent (core object): %v", err)
	}

	if name != "note" {
		t.Errorf("expected core-object fallback to event object name %q, got %q", "note", name)
	}

	// An unseeded object_id must surface as not-found rather than silently misrouting.
	store.Clear()

	if _, err := conn.GetObjectNameFromEvent(context.Background(), recordCreated); err == nil {
		t.Error("expected unseeded object id to resolve with an error")
	}
}
