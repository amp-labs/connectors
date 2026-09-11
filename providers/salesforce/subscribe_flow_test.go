package salesforce

import (
	"context"
	"errors"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/metadata"
)

func TestFlowRecordTriggerType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		events          common.SubscriptionEventTypes
		expectedTrigger metadata.RecordTriggerType
		expectErr       bool
	}{
		{
			name: "Create and update",
			events: common.SubscriptionEventTypes{
				common.SubscriptionEventTypeCreate,
				common.SubscriptionEventTypeUpdate,
			},
			expectedTrigger: metadata.RecordTriggerTypeCreateAndUpdate,
		},
		{
			name:            "Create only",
			events:          common.SubscriptionEventTypes{common.SubscriptionEventTypeCreate},
			expectedTrigger: metadata.RecordTriggerTypeCreate,
		},
		{
			name:            "Update only",
			events:          common.SubscriptionEventTypes{common.SubscriptionEventTypeUpdate},
			expectedTrigger: metadata.RecordTriggerTypeUpdate,
		},
		{
			name: "Delete with create is an error",
			events: common.SubscriptionEventTypes{
				common.SubscriptionEventTypeCreate,
				common.SubscriptionEventTypeDelete,
			},
			expectErr: true,
		},
		{
			name:      "Delete only is an error",
			events:    common.SubscriptionEventTypes{common.SubscriptionEventTypeDelete},
			expectErr: true,
		},
		{
			name:      "No events is an error",
			events:    common.SubscriptionEventTypes{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			trigger, err := flowRecordTriggerType(common.ObjectEvents{Events: tt.events})
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !errors.Is(err, errFlowNoSupportedEvents) {
					t.Fatalf("expected errFlowNoSupportedEvents, got: %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if trigger != tt.expectedTrigger {
				t.Errorf("trigger = %q, want %q", trigger, tt.expectedTrigger)
			}
		})
	}
}

func TestSubscribeWithFlowRequiresFlowConfig(t *testing.T) {
	t.Parallel()

	conn := &Connector{}

	tests := []struct {
		name    string
		req     *SubscriptionRequest
		wantErr error
	}{
		{
			name:    "Nil flow config",
			req:     &SubscriptionRequest{UseFlow: true},
			wantErr: errFlowEndpointURLRequired,
		},
		{
			name: "Missing endpoint URL",
			req: &SubscriptionRequest{
				UseFlow: true,
				Flow:    &FlowConfig{IntegrationUsername: "user@example.com"},
			},
			wantErr: errFlowEndpointURLRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Subscribe must reject before touching Salesforce (or requiring a
			// RegistrationResult, which the flow path doesn't use).
			_, err := conn.Subscribe(context.Background(), common.SubscribeParams{
				Request: tt.req,
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

// TestFlowObjectsToRemove covers the only decision reconcile makes on its own:
// which of the previous subscription's objects are no longer covered and must be
// torn down. Everything else is an upsert, which needs no diffing.
func TestFlowObjectsToRemove(t *testing.T) {
	t.Parallel()

	recorded := func(names ...string) *SubscribeResult {
		flows := make(map[common.ObjectName]*FlowSubscription, len(names))
		for _, name := range names {
			flows[common.ObjectName(name)] = &FlowSubscription{
				ObjectName:      common.ObjectName(name),
				Flow:            &FlowMetadata{Name: "AmpSubscribe_" + name},
				OutboundMessage: &OutboundMessageMetadata{Name: "amp_" + name},
			}
		}

		return &SubscribeResult{UseFlow: true, Flows: flows}
	}

	tests := []struct {
		name string
		prev *SubscribeResult
		next *SubscribeResult
		want []common.ObjectName
	}{{
		name: "Unchanged object set removes nothing",
		prev: recorded("Account", "Contact"),
		next: recorded("Account", "Contact"),
		want: []common.ObjectName{},
	}, {
		name: "Dropped objects are removed",
		prev: recorded("Account", "Contact", "Lead"),
		next: recorded("Account"),
		want: []common.ObjectName{"Contact", "Lead"},
	}, {
		name: "Added objects remove nothing",
		prev: recorded("Account"),
		next: recorded("Account", "Contact"),
		want: []common.ObjectName{},
	}, {
		name: "No previous state removes nothing",
		prev: nil,
		next: recorded("Account"),
		want: nil,
	}, {
		name: "Emptied subscription removes everything",
		prev: recorded("Account", "Contact"),
		next: &SubscribeResult{UseFlow: true},
		want: []common.ObjectName{"Account", "Contact"},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := flowObjectsToRemove(tt.prev, tt.next)
			if len(got) != len(tt.want) {
				t.Fatalf("flowObjectsToRemove() = %v, want %v", got, tt.want)
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("flowObjectsToRemove()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// A previous entry with no recorded artifacts has nothing to delete; reconcile
// must skip it rather than dereference a nil Flow during teardown.
func TestFlowObjectsToRemoveSkipsUnrecordedArtifacts(t *testing.T) {
	t.Parallel()

	prev := &SubscribeResult{
		UseFlow: true,
		Flows: map[common.ObjectName]*FlowSubscription{
			"Account": nil,
			"Contact": {ObjectName: "Contact"},
			"Lead": {
				ObjectName:      "Lead",
				Flow:            &FlowMetadata{Name: "AmpSubscribe_Lead"},
				OutboundMessage: &OutboundMessageMetadata{Name: "amp_Lead"},
			},
		},
	}

	got := flowObjectsToRemove(prev, &SubscribeResult{UseFlow: true})
	if len(got) != 1 || got[0] != "Lead" {
		t.Errorf("flowObjectsToRemove() = %v, want [Lead]", got)
	}
}
