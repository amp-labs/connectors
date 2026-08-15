package stripe

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestDeleteSubscription(t *testing.T) {
	t.Parallel()

	tests := []testconn.TestCaseDeleteSubscription{
		{
			Name:         "Nil result",
			Input:        common.SubscriptionResult{Result: nil},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Empty subscriptions",
			Input: common.SubscriptionResult{
				Result: &SubscriptionResult{},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Delete subscription always deletes endpoint",
			Input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					WebhookId: "we_123",
					Subscriptions: map[common.ObjectName][]string{
						"accounts": {"account.application.authorized", "account.updated"},
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

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSubscriptionRemover, error) {
				return constructTestConnector(tt.Server)
			})
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
				Result: &SubscriptionResult{},
			},
			expectedErr: errMissingParams,
			description: "Test validation with empty subscriptions map",
		},
		{
			name: "Valid result",
			input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					WebhookId: "we_123",
					Subscriptions: map[common.ObjectName][]string{
						"account": {"account.updated"},
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
