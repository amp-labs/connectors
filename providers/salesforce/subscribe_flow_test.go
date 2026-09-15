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
