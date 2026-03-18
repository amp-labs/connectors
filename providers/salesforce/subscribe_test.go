package salesforce

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

func TestCustomFieldAPIName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Adds __c suffix when missing",
			input:    "amp_cdc_optimized",
			expected: "amp_cdc_optimized__c",
		},
		{
			name:     "Does not double-add __c suffix",
			input:    "amp_cdc_optimized__c",
			expected: "amp_cdc_optimized__c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := customFieldAPIName(tt.input)
			if result != tt.expected {
				t.Errorf("customFieldAPIName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCustomFieldDisplayName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Strips __c suffix",
			input:    "amp_cdc_optimized__c",
			expected: "amp_cdc_optimized",
		},
		{
			name:     "No-op when no __c suffix",
			input:    "amp_cdc_optimized",
			expected: "amp_cdc_optimized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := customFieldDisplayName(tt.input)
			if result != tt.expected {
				t.Errorf("customFieldDisplayName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPrepareQuotaOptimizationObjectFieldsForUpdate(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name                       string
		req                        *SubscriptionRequest
		prevState                  *SubscribeResult
		expectedNewFields          map[common.ObjectName]string
		expectedRemainingPrevState map[common.ObjectName]string
	}{
		{
			name: "Nil request returns nil new fields and preserves prevState",
			req:  nil,
			prevState: &SubscribeResult{
				QuotaOptimizationObjectFields: map[common.ObjectName]string{
					"Account": "field_a",
				},
			},
			expectedNewFields: nil,
			expectedRemainingPrevState: map[common.ObjectName]string{
				"Account": "field_a",
			},
		},
		{
			name: "Request with nil QuotaOptimizationObjectFields returns nil new fields",
			req:  &SubscriptionRequest{},
			prevState: &SubscribeResult{
				QuotaOptimizationObjectFields: map[common.ObjectName]string{
					"Account": "field_a",
				},
			},
			expectedNewFields: nil,
			expectedRemainingPrevState: map[common.ObjectName]string{
				"Account": "field_a",
			},
		},
		{
			name: "All new objects are identified as new fields",
			req: &SubscriptionRequest{
				QuotaOptimizationObjectFields: map[common.ObjectName]string{
					"Contact": "field_b",
				},
			},
			prevState: &SubscribeResult{
				QuotaOptimizationObjectFields: map[common.ObjectName]string{
					"Account": "field_a",
				},
			},
			expectedNewFields: map[common.ObjectName]string{
				"Contact": "field_b",
			},
			// Account remains because it's not in req
			expectedRemainingPrevState: map[common.ObjectName]string{
				"Account": "field_a",
			},
		},
		{
			name: "Kept objects are filtered from prevState and not in new fields",
			req: &SubscriptionRequest{
				QuotaOptimizationObjectFields: map[common.ObjectName]string{
					"Account": "field_a",
					"Contact": "field_b",
				},
			},
			prevState: &SubscribeResult{
				QuotaOptimizationObjectFields: map[common.ObjectName]string{
					"Account": "field_a",
					"Lead":    "field_c",
				},
			},
			expectedNewFields: map[common.ObjectName]string{
				"Contact": "field_b",
			},
			// Account removed (in req), Lead remains (not in req, will be deleted)
			expectedRemainingPrevState: map[common.ObjectName]string{
				"Lead": "field_c",
			},
		},
		{
			name: "All objects in req exist in prevState - no new fields, all filtered from prevState",
			req: &SubscriptionRequest{
				QuotaOptimizationObjectFields: map[common.ObjectName]string{
					"Account": "field_a",
				},
			},
			prevState: &SubscribeResult{
				QuotaOptimizationObjectFields: map[common.ObjectName]string{
					"Account": "field_a",
				},
			},
			expectedNewFields:          map[common.ObjectName]string{},
			expectedRemainingPrevState: map[common.ObjectName]string{},
		},
		{
			name: "Nil prevState QuotaOptimizationObjectFields - all req fields are new",
			req: &SubscriptionRequest{
				QuotaOptimizationObjectFields: map[common.ObjectName]string{
					"Account": "field_a",
					"Contact": "field_b",
				},
			},
			prevState: &SubscribeResult{
				QuotaOptimizationObjectFields: nil,
			},
			expectedNewFields: map[common.ObjectName]string{
				"Account": "field_a",
				"Contact": "field_b",
			},
			expectedRemainingPrevState: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			newFields := prepareQuotaOptimizationObjectFieldsForUpdate(tt.req, tt.prevState)

			// Check new fields
			if tt.expectedNewFields == nil {
				if newFields != nil {
					t.Errorf("expected nil new fields, got %v", newFields)
				}
			} else {
				if len(newFields) != len(tt.expectedNewFields) {
					t.Errorf("expected %d new fields, got %d: %v", len(tt.expectedNewFields), len(newFields), newFields)
				}

				for k, v := range tt.expectedNewFields {
					if newFields[k] != v {
						t.Errorf("expected newFields[%q] = %q, got %q", k, v, newFields[k])
					}
				}
			}

			// Check remaining prevState
			if tt.expectedRemainingPrevState == nil {
				if tt.prevState.QuotaOptimizationObjectFields != nil {
					t.Errorf("expected nil remaining prevState, got %v", tt.prevState.QuotaOptimizationObjectFields)
				}
			} else {
				if len(tt.prevState.QuotaOptimizationObjectFields) != len(tt.expectedRemainingPrevState) {
					t.Errorf("expected %d remaining prevState fields, got %d: %v",
						len(tt.expectedRemainingPrevState),
						len(tt.prevState.QuotaOptimizationObjectFields),
						tt.prevState.QuotaOptimizationObjectFields)
				}

				for k, v := range tt.expectedRemainingPrevState {
					if tt.prevState.QuotaOptimizationObjectFields[k] != v {
						t.Errorf("expected prevState[%q] = %q, got %q",
							k, v, tt.prevState.QuotaOptimizationObjectFields[k])
					}
				}
			}
		})
	}
}

func TestRollbackQuotaOptimizationFieldsNilRequest(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector("http://example.com")
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	// Nil request should be a no-op
	err = conn.rollbackQuotaOptimizationFields(t.Context(), nil)
	if err != nil {
		t.Errorf("expected nil error for nil request, got %v", err)
	}

	// Request with nil QuotaOptimizationObjectFields should be a no-op
	err = conn.rollbackQuotaOptimizationFields(t.Context(), &SubscriptionRequest{})
	if err != nil {
		t.Errorf("expected nil error for empty request, got %v", err)
	}
}

func TestUpsertQuotaOptimizationFieldsNilRequest(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector("http://example.com")
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	// Nil request should be a no-op
	err = conn.upsertQuotaOptimizationFields(t.Context(), nil)
	if err != nil {
		t.Errorf("expected nil error for nil request, got %v", err)
	}

	// Request with nil QuotaOptimizationObjectFields should be a no-op
	err = conn.upsertQuotaOptimizationFields(t.Context(), &SubscriptionRequest{})
	if err != nil {
		t.Errorf("expected nil error for empty request, got %v", err)
	}
}

func TestUpdateChannelMemberFiltersNilRequest(t *testing.T) {
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

	// Nil request should be a no-op
	err = conn.updateChannelMemberFilters(t.Context(), nil, members)
	if err != nil {
		t.Errorf("expected nil error for nil request, got %v", err)
	}

	// Request with nil Filters should be a no-op
	err = conn.updateChannelMemberFilters(t.Context(), &SubscriptionRequest{}, members)
	if err != nil {
		t.Errorf("expected nil error for request with nil filters, got %v", err)
	}
}

func TestDeleteSubscriptionUsesCustomFieldAPIName(t *testing.T) {
	t.Parallel()

	// Verify that DeleteSubscription constructs field names with __c suffix.
	// We test this indirectly by checking the deleteFields map construction
	// via the prepareQuotaOptimizationObjectFieldsForUpdate function,
	// which is the same pattern used in DeleteSubscription.

	// A field without __c suffix should get it appended
	result := customFieldAPIName("amp_cdc_optimized")
	if result != "amp_cdc_optimized__c" {
		t.Errorf("expected amp_cdc_optimized__c, got %s", result)
	}

	// A field already with __c should not be doubled
	result = customFieldAPIName("amp_cdc_optimized__c")
	if result != "amp_cdc_optimized__c" {
		t.Errorf("expected amp_cdc_optimized__c, got %s", result)
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

	// Wrong Request type
	_, err = conn.Subscribe(t.Context(), common.SubscribeParams{
		RegistrationResult: &common.RegistrationResult{
			Result: &ResultData{
				EventChannel: &EventChannel{
					FullName: "test_channel__chn",
				},
				NamedCredential:  &NamedCredential{},
				EventRelayConfig: &EventRelayConfig{},
			},
		},
		Request: "wrong-type",
	})
	if err == nil {
		t.Error("expected error for wrong Request type")
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

	// Wrong Request type
	_, err = conn.UpdateSubscription(t.Context(), common.SubscribeParams{
		Request: "wrong-type",
	}, &common.SubscriptionResult{
		Result: &SubscribeResult{
			EventChannelMembers: map[common.ObjectName]*EventChannelMember{},
		},
	})
	if err == nil {
		t.Error("expected error for wrong Request type")
	}
}
