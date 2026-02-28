package stripe

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestDeleteSubscription(t *testing.T) {
	t.Parallel()

	tests := []testroutines.TestCase[common.SubscriptionResult, error]{
		{
			Name:         "Nil result",
			Input:        common.SubscriptionResult{Result: nil},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Empty subscriptions",
			Input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					Subscriptions: map[common.ObjectName]WebhookResponse{},
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Delete subscription always deletes endpoint",
			Input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					Subscriptions: map[common.ObjectName]WebhookResponse{
						"account": {
							ID:            "we_123:account",
							EnabledEvents: []string{"account.application.authorized", "account.updated"},
						},
					},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodDELETE(),
							mockcond.Path("/v1/webhook_endpoints/we_123"),
						},
						Then: mockserver.Response(http.StatusOK),
					},
				},
			}.Server(),
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			t.Cleanup(func() {
				tt.Close()
			})

			conn, err := constructTestConnector(tt.Server.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			err = conn.DeleteSubscription(t.Context(), tt.Input)
			tt.Validate(t, err, nil)
		})
	}
}

func TestValidateSubscriptionResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       common.SubscriptionResult
		expectedErr error
		description string
	}{
		{
			name:        "Nil result",
			input:       common.SubscriptionResult{Result: nil},
			expectedErr: errMissingParams,
			description: "Test validation with nil result field",
		},
		{
			name: "Invalid result type",
			input: common.SubscriptionResult{
				Result: "invalid type",
			},
			expectedErr: errInvalidRequestType,
			description: "Test validation with invalid result type",
		},
		{
			name: "Empty subscriptions",
			input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					Subscriptions: map[common.ObjectName]WebhookResponse{},
				},
			},
			expectedErr: errMissingParams,
			description: "Test validation with empty subscriptions map",
		},
		{
			name: "Valid result",
			input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					Subscriptions: map[common.ObjectName]WebhookResponse{
						"account": {
							ID:            "we_123:account",
							EnabledEvents: []string{"account.updated"},
						},
					},
				},
			},
			expectedErr: nil,
			description: "Test validation with valid subscription result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateSubscriptionResult(tt.input)
			if tt.expectedErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
