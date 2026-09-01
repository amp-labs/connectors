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
		expectedEvents  []common.SubscriptionEventType
		expectErr       bool
	}{
		{
			name: "Create and update",
			events: common.SubscriptionEventTypes{
				common.SubscriptionEventTypeCreate,
				common.SubscriptionEventTypeUpdate,
			},
			expectedTrigger: metadata.RecordTriggerTypeCreateAndUpdate,
			expectedEvents: []common.SubscriptionEventType{
				common.SubscriptionEventTypeCreate,
				common.SubscriptionEventTypeUpdate,
			},
		},
		{
			name:            "Create only",
			events:          common.SubscriptionEventTypes{common.SubscriptionEventTypeCreate},
			expectedTrigger: metadata.RecordTriggerTypeCreate,
			expectedEvents:  []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
		},
		{
			name:            "Update only",
			events:          common.SubscriptionEventTypes{common.SubscriptionEventTypeUpdate},
			expectedTrigger: metadata.RecordTriggerTypeUpdate,
			expectedEvents:  []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
		},
		{
			name: "Delete is dropped alongside supported events",
			events: common.SubscriptionEventTypes{
				common.SubscriptionEventTypeCreate,
				common.SubscriptionEventTypeDelete,
			},
			expectedTrigger: metadata.RecordTriggerTypeCreate,
			expectedEvents:  []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
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

			trigger, events, err := flowRecordTriggerType(
				t.Context(), "Account", common.ObjectEvents{Events: tt.events},
			)
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

			if !common.SubscriptionEventTypes(events).Equals(tt.expectedEvents) {
				t.Errorf("events = %v, want %v", events, tt.expectedEvents)
			}
		})
	}
}

func TestSubscribeWithFlowRequiresFlowConfig(t *testing.T) {
	t.Parallel()

	conn := &Connector{}

	tests := []struct {
		name string
		req  *SubscriptionRequest
	}{
		{
			name: "Nil flow config",
			req:  &SubscriptionRequest{UseFlow: true},
		},
		{
			name: "Missing endpoint URL",
			req: &SubscriptionRequest{
				UseFlow: true,
				Flow:    &FlowSubscriptionConfig{IntegrationUsername: "user@example.com"},
			},
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
			if !errors.Is(err, errFlowConfigMissing) {
				t.Fatalf("expected errFlowConfigMissing, got: %v", err)
			}
		})
	}
}

func TestFlowCoveredEvents(t *testing.T) {
	t.Parallel()

	sfRes := &SubscribeResult{
		Flows: map[common.ObjectName]*FlowSubscription{
			"account": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
				},
			},
			"contact": {
				Events: []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
			},
			"lead": nil,
		},
	}

	got := flowCoveredEvents(sfRes)

	expected := common.SubscriptionEventTypes{
		common.SubscriptionEventTypeCreate,
		common.SubscriptionEventTypeUpdate,
	}
	if !common.SubscriptionEventTypes(got).Equals(expected) {
		t.Errorf("flowCoveredEvents = %v, want %v", got, expected)
	}
}

func TestFlowParamsFromResult(t *testing.T) {
	t.Parallel()

	flowSub := &FlowSubscription{
		ObjectName: "Account",
		Flow: &FlowMetadata{
			Name:              "AmpSubscribe_Account",
			RecordTriggerType: string(metadata.RecordTriggerTypeUpdate),
			WatchFields:       []string{"Email"},
		},
		OutboundMessage: &OutboundMessageMetadata{
			Name:                "amp_Account",
			EndpointURL:         "https://example.com/webhook",
			IntegrationUsername: "user@example.com",
		},
	}

	params := flowParamsFromResult(flowSub)

	// The reconstructed params must be valid so the teardown path can
	// regenerate the deactivation flow XML without the original request.
	if err := metadata.ValidateFlowParams(params); err != nil {
		t.Fatalf("reconstructed params invalid: %v", err)
	}

	if params.RecordTriggerType != metadata.RecordTriggerTypeUpdate {
		t.Errorf("RecordTriggerType = %q, want %q", params.RecordTriggerType, metadata.RecordTriggerTypeUpdate)
	}
}

func TestResolveFlowConfigDoesNotMutateRequest(t *testing.T) {
	t.Parallel()

	conn := &Connector{}
	original := &FlowSubscriptionConfig{
		EndpointURL:         "https://example.com/webhook",
		IntegrationUsername: "user@example.com",
	}

	resolved, err := conn.resolveFlowConfig(t.Context(), original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved == original {
		t.Error("resolveFlowConfig must return a copy, not the caller's config")
	}

	if resolved.IntegrationUsername != "user@example.com" {
		t.Errorf("IntegrationUsername = %q, want the configured value", resolved.IntegrationUsername)
	}
}
