package subscribe

import (
	"context"
	"errors"
	"testing"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce"
	"github.com/amp-labs/connectors/subscribe/deps"
)

func TestBuildCDCEventFlagFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		appName        string
		enabledObjects []common.ObjectName
		expectedResult map[common.ObjectName]string
		expectedErr    error
	}{
		{
			name:           "empty appName returns error",
			appName:        "",
			enabledObjects: []common.ObjectName{"obj1"},
			expectedErr:    ErrAppNameRequired,
		},
		{
			name:           "appName that sanitizes to empty returns invalid error",
			appName:        "!@#$%",
			enabledObjects: []common.ObjectName{"Account"},
			expectedErr:    ErrAppNameInvalid,
		},
		{
			name:           "empty enabledObjects returns empty map",
			appName:        "myapp",
			enabledObjects: []common.ObjectName{},
			expectedResult: map[common.ObjectName]string{},
		},
		{
			name:           "nil enabledObjects returns empty map",
			appName:        "myapp",
			enabledObjects: nil,
			expectedResult: map[common.ObjectName]string{},
		},
		{
			name:           "single object name produces field with sanitized appName prefix",
			appName:        "myapp",
			enabledObjects: []common.ObjectName{"Account"},
			expectedResult: map[common.ObjectName]string{
				"Account": "myapp_cdc_event_flag__c",
			},
		},
		{
			name:           "appName with special characters is sanitized in field value",
			appName:        "My App-Name!",
			enabledObjects: []common.ObjectName{"Account"},
			expectedResult: map[common.ObjectName]string{
				"Account": "my_app_name_cdc_event_flag__c",
			},
		},
		{
			name:           "multiple object names share the one app-derived field",
			appName:        "myapp",
			enabledObjects: []common.ObjectName{"Account", "Contact"},
			expectedResult: map[common.ObjectName]string{
				"Account": "myapp_cdc_event_flag__c",
				"Contact": "myapp_cdc_event_flag__c",
			},
		},
		{
			name:           "repeated object name collapses to one entry",
			appName:        "myapp",
			enabledObjects: []common.ObjectName{"Account", "Account"},
			expectedResult: map[common.ObjectName]string{
				"Account": "myapp_cdc_event_flag__c",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := buildCDCEventFlagFields(testCase.appName, testCase.enabledObjects)

			if testCase.expectedErr != nil {
				if !errors.Is(err, testCase.expectedErr) {
					t.Fatalf("expected error %v, got %v", testCase.expectedErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(testCase.expectedResult) {
				t.Fatalf("expected %d entries, got %d: %v", len(testCase.expectedResult), len(result), result)
			}

			for k, v := range testCase.expectedResult {
				got, ok := result[k]
				if !ok {
					t.Errorf("expected key %q not found in result", k)
				}

				if got != v {
					t.Errorf("for key %q: expected %q, got %q", k, v, got)
				}
			}
		})
	}
}

func TestSanitizeAppNameForSalesforce(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase stays unchanged",
			input:    "myapp",
			expected: "myapp",
		},
		{
			name:     "uppercase is lowercased",
			input:    "MyApp",
			expected: "myapp",
		},
		{
			name:     "hyphens replaced with underscores",
			input:    "my-app",
			expected: "my_app",
		},
		{
			name:     "spaces replaced with underscores",
			input:    "my app",
			expected: "my_app",
		},
		{
			name:     "special characters removed",
			input:    "my@app!name#",
			expected: "myappname",
		},
		{
			name:     "underscores preserved",
			input:    "my_app_name",
			expected: "my_app_name",
		},
		{
			name:     "numbers preserved",
			input:    "app123",
			expected: "app123",
		},
		{
			name:     "trailing underscores trimmed",
			input:    "myapp---",
			expected: "myapp",
		},
		{
			name:     "trailing spaces trimmed as underscores",
			input:    "myapp   ",
			expected: "myapp",
		},
		{
			name:     "mixed case with hyphens spaces and special chars",
			input:    "My App-Name!",
			expected: "my_app_name",
		},
		{
			name:     "empty string returns empty",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters returns empty",
			input:    "!@#$%",
			expected: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := SanitizeAppNameForSalesforce(testCase.input)
			if result != testCase.expected {
				t.Errorf("SanitizeAppNameForSalesforce(%q) = %q, want %q", testCase.input, result, testCase.expected)
			}
		})
	}
}

// TestGetSalesforceRequestNoCDCResolver verifies the degrade-to-no-payload contract: when no
// CDCOptimization resolver is supplied (or it returns nil), the Salesforce builder returns
// (nil, nil) — the same behavior as a project with no CDC opt-in.
func TestGetSalesforceRequestNoCDCResolver(t *testing.T) {
	t.Parallel()

	req, err := getSalesforceRequest(
		context.Background(), deps.Dependencies{}, &openapi.Installation{Id: "inst-1", ProjectId: "proj-1"},
		nil, nil, nil, "")
	if err != nil {
		t.Errorf("getSalesforceRequest() error = %v, want nil", err)
	}

	if req != nil {
		t.Errorf("getSalesforceRequest() = %v, want nil", req)
	}
}

// TestGetSalesforceRequestNilConfig verifies that a resolver returning nil config makes
// getSalesforceRequest return (nil, nil). With no SubscriptionRequest, UpdateSubscription
// leaves existing CDC quota-optimization state unchanged.
func TestGetSalesforceRequestNilConfig(t *testing.T) {
	t.Parallel()

	dependencies := deps.Dependencies{
		Project:         stubProjectResolver{appName: "My App"},
		CDCOptimization: stubCDCOptimizationResolver{cfg: nil},
	}

	req, err := getSalesforceRequest(
		context.Background(), dependencies, &openapi.Installation{Id: "inst-1", ProjectId: "proj-1"},
		nil, nil, nil, "")
	if err != nil {
		t.Errorf("getSalesforceRequest() error = %v, want nil", err)
	}

	if req != nil {
		t.Errorf("getSalesforceRequest() = %v, want nil", req)
	}
}

// TestGetSalesforceRequestExplicitDisable verifies that a non-nil config with no objects enabled
// returns a SubscriptionRequest whose QuotaOptimizationObjectFields map is empty. That empty map
// tells UpdateSubscription to tear down CDC quota optimization; returning nil instead would leave
// existing filters and triggers in place.
func TestGetSalesforceRequestExplicitDisable(t *testing.T) {
	t.Parallel()

	// An object the caller resolved to disabled reaches here the same way one that was never
	// configured does — absent from EnabledObjects — so both arrive as one of the cases below.
	tests := []struct {
		name           string
		enabledObjects []common.ObjectName
	}{
		{
			name:           "empty opt-in list",
			enabledObjects: []common.ObjectName{},
		},
		{
			name:           "nil opt-in list",
			enabledObjects: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			dependencies := deps.Dependencies{
				Project: stubProjectResolver{appName: "My App"},
				CDCOptimization: stubCDCOptimizationResolver{cfg: &deps.CDCOptimizationConfig{
					ManualCheckboxManagement:    true,
					ManualApexTriggerManagement: true,
					EnabledObjects:              testCase.enabledObjects,
				}},
			}

			got, err := getSalesforceRequest(
				context.Background(), dependencies, &openapi.Installation{Id: "inst-1", ProjectId: "proj-1"},
				nil, nil, nil, "")
			if err != nil {
				t.Fatalf("getSalesforceRequest() error = %v, want nil", err)
			}

			req, ok := got.(*salesforce.SubscriptionRequest)
			if !ok {
				t.Fatalf("getSalesforceRequest() = %T, want *salesforce.SubscriptionRequest", got)
			}

			// Non-nil map, not just empty: the connector distinguishes the two nowhere today, but
			// the payload's whole job here is to be an explicit, present instruction.
			if req.QuotaOptimizationObjectFields == nil {
				t.Error("QuotaOptimizationObjectFields = nil, want non-nil empty map")
			}

			if len(req.QuotaOptimizationObjectFields) != 0 {
				t.Errorf("QuotaOptimizationObjectFields = %v, want empty", req.QuotaOptimizationObjectFields)
			}

			// The manual-mode flags still ride along: they decide whether the teardown may
			// destructively remove caller-owned artifacts.
			if !req.ManualCheckboxManagement || !req.ManualApexTriggerManagement {
				t.Errorf("flags = (%v, %v), want (true, true)",
					req.ManualCheckboxManagement, req.ManualApexTriggerManagement)
			}
		})
	}
}

// TestGetSalesforceRequestEnabledObjects verifies that exactly the enabled objects appear in
// QuotaOptimizationObjectFields, each pointing at the one app-derived checkbox field. Objects the
// caller left out are absent, which is what lets UpdateSubscription tear their artifacts down.
func TestGetSalesforceRequestEnabledObjects(t *testing.T) {
	t.Parallel()

	dependencies := deps.Dependencies{
		Project: stubProjectResolver{appName: "My App"},
		CDCOptimization: stubCDCOptimizationResolver{cfg: &deps.CDCOptimizationConfig{
			EnabledObjects: []common.ObjectName{"Account", "Contact"},
		}},
	}

	got, err := getSalesforceRequest(
		context.Background(), dependencies, &openapi.Installation{Id: "inst-1", ProjectId: "proj-1"},
		nil, nil, nil, "")
	if err != nil {
		t.Fatalf("getSalesforceRequest() error = %v, want nil", err)
	}

	req, ok := got.(*salesforce.SubscriptionRequest)
	if !ok {
		t.Fatalf("getSalesforceRequest() = %T, want *salesforce.SubscriptionRequest", got)
	}

	want := map[common.ObjectName]string{
		"Account": "my_app_cdc_event_flag__c",
		"Contact": "my_app_cdc_event_flag__c",
	}
	if len(req.QuotaOptimizationObjectFields) != len(want) {
		t.Fatalf("QuotaOptimizationObjectFields = %v, want %v", req.QuotaOptimizationObjectFields, want)
	}

	for objName, fieldName := range want {
		if req.QuotaOptimizationObjectFields[objName] != fieldName {
			t.Errorf("QuotaOptimizationObjectFields[%s] = %q, want %q",
				objName, req.QuotaOptimizationObjectFields[objName], fieldName)
		}
	}
}

// TestGetSalesforceRequestResolverError verifies that a failed resolution propagates instead of
// degrading to an empty payload. An empty payload would read as an explicit disable and tear down
// a live optimization on what may be a transient failure.
func TestGetSalesforceRequestResolverError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("resolving subscribe objects")

	dependencies := deps.Dependencies{
		Project:         stubProjectResolver{appName: "My App"},
		CDCOptimization: stubCDCOptimizationResolver{err: wantErr},
	}

	got, err := getSalesforceRequest(
		context.Background(), dependencies, &openapi.Installation{Id: "inst-1", ProjectId: "proj-1"},
		nil, nil, nil, "")
	if !errors.Is(err, wantErr) {
		t.Errorf("getSalesforceRequest() error = %v, want %v", err, wantErr)
	}

	if got != nil {
		t.Errorf("getSalesforceRequest() = %v, want nil", got)
	}
}
