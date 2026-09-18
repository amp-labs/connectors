package salesforce

import (
	"context"
	"errors"
	"strings"
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

// TestUpdateSubscriptionFlowSkipsRegistrationChecks pins that a flow update is not rejected for
// having no RegistrationResult. Flow subscriptions never register, so the CDC validation that
// UpdateSubscription runs up front must sit behind the flow branch, as it does in Subscribe.
func TestUpdateSubscriptionFlowSkipsRegistrationChecks(t *testing.T) {
	t.Parallel()

	conn := &Connector{}

	previous := &common.SubscriptionResult{
		Result: &SubscribeResult{UseFlow: true},
	}

	params := common.SubscribeParams{
		Request: &SubscriptionRequest{UseFlow: true, Flow: &FlowConfig{}},
		// RegistrationResult deliberately nil: a flow subscription has none.
	}

	_, err := conn.UpdateSubscription(t.Context(), params, previous)

	// The call cannot succeed without a live org, but it must fail past the registration checks
	// rather than on them.
	if err != nil && strings.Contains(err.Error(), "missing RegistrationResult") {
		t.Fatalf("flow update rejected for a missing registration: %v", err)
	}
}

// TestUpdateSubscriptionCDCStillRequiresRegistration pins the other side: a CDC update with no
// registration is still rejected up front, before any Salesforce-side mutation.
func TestUpdateSubscriptionCDCStillRequiresRegistration(t *testing.T) {
	t.Parallel()

	conn := &Connector{}

	previous := &common.SubscriptionResult{
		Result: &SubscribeResult{},
	}

	params := common.SubscribeParams{
		Request: &SubscriptionRequest{},
	}

	_, err := conn.UpdateSubscription(t.Context(), params, previous)
	if err == nil || !strings.Contains(err.Error(), "missing RegistrationResult") {
		t.Fatalf("CDC update error = %v, want it rejected for a missing registration", err)
	}
}
