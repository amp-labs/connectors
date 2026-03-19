package salesforce

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

func TestUpdateChannelMemberFiltersNoEvents(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector("http://example.com")
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	members := map[common.ObjectName]*EventChannelMember{
		"Account": {
			Id:       "member-1",
			FullName: "test",
			Metadata: &EventChannelMemberMetadata{
				EventChannel: "test-channel",
			},
		},
	}

	// Empty SubscriptionEvents should be a no-op (no matching objects to update)
	err = conn.updateChannelMemberFilters(t.Context(), common.SubscribeParams{}, members)
	if err != nil {
		t.Errorf("expected nil error for empty params, got %v", err)
	}
}

func TestDeleteSubscriptionValidation(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector("http://example.com")
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	ctx := common.WithAuthToken(t.Context(), "TEST_TOKEN")

	// Missing Result
	err = conn.DeleteSubscription(ctx, common.SubscriptionResult{})
	if err == nil {
		t.Error("expected error for nil Result")
	}

	// Wrong Result type
	err = conn.DeleteSubscription(ctx, common.SubscriptionResult{
		Result: "wrong-type",
	})
	if err == nil {
		t.Error("expected error for wrong Result type")
	}
}

func TestSubscribeValidation(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector("http://example.com")
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	// Missing RegistrationResult
	_, err = conn.Subscribe(t.Context(), common.SubscribeParams{})
	if err == nil {
		t.Error("expected error for nil RegistrationResult")
	}

	// Missing RegistrationResult.Result
	_, err = conn.Subscribe(t.Context(), common.SubscribeParams{
		RegistrationResult: &common.RegistrationResult{},
	})
	if err == nil {
		t.Error("expected error for nil RegistrationResult.Result")
	}
}

func TestUpdateSubscriptionValidation(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector("http://example.com")
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	// Missing previousResult.Result
	_, err = conn.UpdateSubscription(t.Context(), common.SubscribeParams{}, &common.SubscriptionResult{})
	if err == nil {
		t.Error("expected error for nil previousResult.Result")
	}

	// Wrong previousResult.Result type
	_, err = conn.UpdateSubscription(t.Context(), common.SubscribeParams{}, &common.SubscriptionResult{
		Result: "wrong-type",
	})
	if err == nil {
		t.Error("expected error for wrong previousResult.Result type")
	}

	// Wrong Request type — no longer applicable since SubscriptionRequest was removed
}

func TestBuildWatchFieldsFilterExpression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		watchFields []string
		expected    string
	}{
		{
			name:        "Empty watch fields returns empty string",
			watchFields: nil,
			expected:    "",
		},
		{
			name:        "Single watch field",
			watchFields: []string{"Phone"},
			expected:    "(ChangeEventHeader.changeType != 'UPDATE') OR (ChangeEventHeader.changeType = 'UPDATE' AND ('Phone' IN ChangeEventHeader.changedFields))",
		},
		{
			name:        "Multiple watch fields",
			watchFields: []string{"Phone", "Email"},
			expected:    "(ChangeEventHeader.changeType != 'UPDATE') OR (ChangeEventHeader.changeType = 'UPDATE' AND ('Phone' IN ChangeEventHeader.changedFields OR 'Email' IN ChangeEventHeader.changedFields))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := buildWatchFieldsFilterExpression(tt.watchFields)
			if result != tt.expected {
				t.Errorf("buildWatchFieldsFilterExpression(%v) =\n  %q\nwant\n  %q", tt.watchFields, result, tt.expected)
			}
		})
	}
}
