package mocksub_test

import (
	"context"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/mocksub"
	"github.com/amp-labs/connectors/providers"
)

// fakeEvent is a minimal common.SubscriptionEvent used to exercise GetObjectNameFromEvent.
type fakeEvent struct {
	objectName string
}

func (e fakeEvent) RawMap() (map[string]any, error) { return map[string]any{}, nil }

func (e fakeEvent) EventType() (common.SubscriptionEventType, error) {
	return common.SubscriptionEventTypeCreate, nil
}

func (e fakeEvent) RawEventName() (string, error) { return "fake_created", nil }

func (e fakeEvent) ObjectName() (string, error) { return e.objectName, nil }

func (e fakeEvent) Workspace() (string, error) { return "", nil }

func (e fakeEvent) RecordId() (string, error) { return "1", nil }

func (e fakeEvent) EventTimeStampNano() (int64, error) { return 0, nil }

func (e fakeEvent) PreLoadData(*common.SubscriptionEventPreLoadData) error { return nil }

func TestGetRecordsByIds(t *testing.T) {
	t.Parallel()

	store := mocksub.NewStore()
	store.Seed("people", "436664215", map[string]any{
		"id":            float64(436664215),
		"email_address": "john.kelly@example.com",
		"first_name":    "John",
		"last_name":     "Kelly",
	})
	store.Seed("people", "436664216", map[string]any{
		"id":            float64(436664216),
		"email_address": "ada.osei@example.com",
		"first_name":    "Ada",
		"last_name":     "Osei",
	})

	conn := mocksub.NewConnector(providers.MockSalesloft, mocksub.WithStore(store))

	rows, err := conn.GetRecordsByIds(
		context.Background(),
		"people",
		[]string{"436664215", "436664216", "999999999"}, // last id is not seeded
		[]string{"email_address"},
		nil,
	)
	if err != nil {
		t.Fatalf("GetRecordsByIds: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (missing id skipped), got %d", len(rows))
	}

	first := rows[0]
	if first.Id != "436664215" {
		t.Errorf("expected Id extracted from record, got %q", first.Id)
	}

	if got := first.Fields["email_address"]; got != "john.kelly@example.com" {
		t.Errorf("expected requested field in Fields, got %v", got)
	}

	if _, ok := first.Fields["first_name"]; ok {
		t.Error("expected unrequested field to be absent from Fields")
	}

	if got := first.Raw["first_name"]; got != "John" {
		t.Errorf("expected Raw to carry the full record, got first_name=%v", got)
	}
}

func TestGetRecordsByIdsEmptyStore(t *testing.T) {
	t.Parallel()

	conn := mocksub.NewConnector(providers.MockSalesloft, mocksub.WithStore(mocksub.NewStore()))

	rows, err := conn.GetRecordsByIds(context.Background(), "people", []string{"1"}, nil, nil)
	if err != nil {
		t.Fatalf("GetRecordsByIds: %v", err)
	}

	if len(rows) != 0 {
		t.Fatalf("expected no rows from empty store, got %d", len(rows))
	}
}

func TestVerifyWebhookMessage(t *testing.T) {
	t.Parallel()

	conn := mocksub.NewConnector(providers.MockSalesloft, mocksub.WithStore(mocksub.NewStore()))

	valid, err := conn.VerifyWebhookMessage(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("VerifyWebhookMessage: %v", err)
	}

	if !valid {
		t.Error("expected mock verification to always pass")
	}
}

func TestGetObjectNameFromEvent(t *testing.T) {
	t.Parallel()

	// Without a resolver, the connector falls back to the event's own ObjectName.
	conn := mocksub.NewConnector(providers.MockSalesloft, mocksub.WithStore(mocksub.NewStore()))

	name, err := conn.GetObjectNameFromEvent(context.Background(), fakeEvent{objectName: "people"})
	if err != nil {
		t.Fatalf("GetObjectNameFromEvent: %v", err)
	}

	if name != "people" {
		t.Errorf("expected fallback to event.ObjectName, got %q", name)
	}

	// With a resolver, the connector delegates to it.
	resolved := mocksub.NewConnector(
		providers.MockSalesloft,
		mocksub.WithStore(mocksub.NewStore()),
		mocksub.WithObjectNameFromEvent(func(context.Context, common.SubscriptionEvent) (string, error) {
			return "companies", nil
		}),
	)

	name, err = resolved.GetObjectNameFromEvent(context.Background(), fakeEvent{objectName: "people"})
	if err != nil {
		t.Fatalf("GetObjectNameFromEvent: %v", err)
	}

	if name != "companies" {
		t.Errorf("expected declared resolver to win, got %q", name)
	}
}

func TestStoreForSingleton(t *testing.T) {
	t.Parallel()

	first := mocksub.StoreFor("mocksub-test-provider")
	second := mocksub.StoreFor("mocksub-test-provider")

	if first != second {
		t.Fatal("expected StoreFor to return the same store per provider")
	}

	first.Seed("people", "1", map[string]any{"id": "1"})

	if _, ok := second.Get("people", "1"); !ok {
		t.Fatal("expected record seeded via one reference to be visible via the other")
	}

	first.Clear()

	if _, ok := second.Get("people", "1"); ok {
		t.Fatal("expected Clear to remove seeded records")
	}
}
