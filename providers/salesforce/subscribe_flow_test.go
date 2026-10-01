package salesforce

import (
	"context"
	"errors"
	"slices"
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

// legacyFlowSubscription records an object in the legacy shape: one CreateAndUpdate pair.
func legacyFlowSubscription(name string) *FlowSubscription {
	return &FlowSubscription{
		ObjectName:      common.ObjectName(name),
		Flow:            &FlowMetadata{Name: "amp_" + name},
		OutboundMessage: &OutboundMessageMetadata{Name: "amp_" + name},
	}
}

// perEventFlowSubscription records an object in the per-event shape: one pair per event type.
func perEventFlowSubscription(name string, eventTypes ...common.SubscriptionEventType) *FlowSubscription {
	pairs := make(map[common.SubscriptionEventType]*FlowPair, len(eventTypes))

	for _, eventType := range eventTypes {
		pairName := "amp_" + name + "_" + string(eventType)
		pairs[eventType] = &FlowPair{
			Flow:            &FlowMetadata{Name: pairName},
			OutboundMessage: &OutboundMessageMetadata{Name: pairName},
		}
	}

	return &FlowSubscription{ObjectName: common.ObjectName(name), Pairs: pairs}
}

func flowResult(subs ...*FlowSubscription) *SubscribeResult {
	flows := make(map[common.ObjectName]*FlowSubscription, len(subs))
	for _, sub := range subs {
		flows[sub.ObjectName] = sub
	}

	return &SubscribeResult{UseFlow: true, Flows: flows}
}

// TestDroppedFlowPairs covers the only decision reconcile makes on its own: which previously
// deployed pairs the new state no longer has and must be torn down. Dropped pairs are reported
// by flow name.
func TestDroppedFlowPairs(t *testing.T) {
	t.Parallel()

	create, update := common.SubscriptionEventTypeCreate, common.SubscriptionEventTypeUpdate

	tests := []struct {
		name string
		prev *SubscribeResult
		next *SubscribeResult
		want []string
	}{{
		name: "Unchanged object set removes nothing",
		prev: flowResult(legacyFlowSubscription("Account"), legacyFlowSubscription("Contact")),
		next: flowResult(legacyFlowSubscription("Account"), legacyFlowSubscription("Contact")),
		want: nil,
	}, {
		name: "Dropped objects are removed",
		prev: flowResult(legacyFlowSubscription("Account"), legacyFlowSubscription("Contact"),
			legacyFlowSubscription("Lead")),
		next: flowResult(legacyFlowSubscription("Account")),
		want: []string{"amp_Contact", "amp_Lead"},
	}, {
		name: "Added objects remove nothing",
		prev: flowResult(legacyFlowSubscription("Account")),
		next: flowResult(legacyFlowSubscription("Account"), legacyFlowSubscription("Contact")),
		want: nil,
	}, {
		name: "No previous state removes nothing",
		prev: nil,
		next: flowResult(legacyFlowSubscription("Account")),
		want: nil,
	}, {
		name: "Emptied subscription removes everything",
		prev: flowResult(legacyFlowSubscription("Account"), legacyFlowSubscription("Contact")),
		next: &SubscribeResult{UseFlow: true},
		want: []string{"amp_Account", "amp_Contact"},
	}, {
		name: "Unchanged per-event pairs remove nothing",
		prev: flowResult(perEventFlowSubscription("Account", create, update)),
		next: flowResult(perEventFlowSubscription("Account", create, update)),
		want: nil,
	}, {
		name: "An event dropped from a kept object removes only its pair",
		prev: flowResult(perEventFlowSubscription("Account", create, update)),
		next: flowResult(perEventFlowSubscription("Account", create)),
		want: []string{"amp_Account_update"},
	}, {
		name: "A dropped per-event object removes all its pairs",
		prev: flowResult(perEventFlowSubscription("Account", create, update)),
		next: &SubscribeResult{UseFlow: true},
		want: []string{"amp_Account_create", "amp_Account_update"},
	}, {
		name: "Moving to per-event pairs removes the legacy pair",
		prev: flowResult(legacyFlowSubscription("Account")),
		next: flowResult(perEventFlowSubscription("Account", create, update)),
		want: []string{"amp_Account"},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got []string

			for objName, pairs := range droppedFlowPairs(tt.prev, tt.next) {
				for _, pair := range pairs {
					if !tt.prev.Flows[objName].hasFlow(pair.Flow.Name) {
						t.Errorf("pair %s reported under %s, which never had it", pair.Flow.Name, objName)
					}

					got = append(got, pair.Flow.Name)
				}
			}

			slices.Sort(got)

			if !slices.Equal(got, tt.want) {
				t.Fatalf("droppedFlowPairs() = %v, want %v", got, tt.want)
			}
		})
	}
}

// A previous entry with no recorded artifacts has nothing to delete; reconcile
// must skip it rather than dereference a nil Flow during teardown.
func TestDroppedFlowPairsSkipsUnrecordedArtifacts(t *testing.T) {
	t.Parallel()

	prev := &SubscribeResult{
		UseFlow: true,
		Flows: map[common.ObjectName]*FlowSubscription{
			"Account": nil,
			"Contact": {ObjectName: "Contact"},
			"Opportunity": {
				ObjectName: "Opportunity",
				Pairs: map[common.SubscriptionEventType]*FlowPair{
					common.SubscriptionEventTypeCreate: {Flow: &FlowMetadata{Name: "amp_Opportunity_create"}},
					common.SubscriptionEventTypeUpdate: nil,
				},
			},
			"Lead": legacyFlowSubscription("Lead"),
		},
	}

	got := droppedFlowPairs(prev, &SubscribeResult{UseFlow: true})
	if len(got) != 1 || len(got["Lead"]) != 1 {
		t.Errorf("droppedFlowPairs() = %v, want only Lead's pair", got)
	}
}

// hasFlow reports whether the record holds a pair with the given flow name.
func (f *FlowSubscription) hasFlow(name string) bool {
	for _, pair := range f.FlowPairs() {
		if pair.Flow.Name == name {
			return true
		}
	}

	return false
}

// TestFlowPairs covers the one accessor every reader goes through: the legacy pair, or the per-event
// pairs in create, update order, and nothing from a nil, empty or half-recorded record.
func TestFlowPairs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sub  *FlowSubscription
		want []string
	}{
		{"nil record", nil, nil},
		{"legacy pair", legacyFlowSubscription("Account"), []string{"amp_Account"}},
		{"per-event in create, update order", perEventFlowSubscription("Account",
			common.SubscriptionEventTypeUpdate, common.SubscriptionEventTypeCreate),
			[]string{"amp_Account_create", "amp_Account_update"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got []string
			for _, pair := range tt.sub.FlowPairs() {
				got = append(got, pair.Flow.Name)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("FlowPairs() = %v, want %v", got, tt.want)
			}
		})
	}
}
