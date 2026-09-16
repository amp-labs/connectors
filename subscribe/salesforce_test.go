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

// flowInstallation builds an installation opted into (or out of) record-triggered flow subscribe.
// The opt-in is installation-level config, which is where the builder now reads the mode from.
func flowInstallation(useFlows bool) *openapi.Installation {
	inst := &openapi.Installation{Id: "inst-1", ProjectId: "proj-1"}
	inst.Config.Content.Subscribe = &openapi.SubscribeConfig{
		ProviderOptions: &openapi.SubscribeInstallationProviderOptions{
			UseSalesforceFlows: &openapi.UseSalesforceFlowsConfig{Enabled: useFlows},
		},
	}

	return inst
}

// withUsername puts the connect-time username on provider metadata, where the server stores
// post-auth values. That is the only place the builder can read it from.
func withUsername(inst *openapi.Installation) *openapi.Installation {
	inst.Connection.ProviderMetadata = &openapi.ProviderMetadata{
		salesforce.PostAuthCatalogVarUsername: {Value: "integration@example.com"},
	}

	return inst
}

// TestGetSalesforceRequestFlowMode verifies the flow-subscribe happy path: the builder assembles
// the FlowConfig from three different sources — the caller's webhook URL, the project app name via
// ProjectResolver, and the installation-resolved fields via SalesforceFlowResolver.
func TestGetSalesforceRequestFlowMode(t *testing.T) {
	t.Parallel()

	dependencies := deps.Dependencies{
		Project: stubProjectResolver{appName: "My App"},
		SalesforceFlow: stubSalesforceFlowResolver{selectedFields: map[common.ObjectName][]string{
			"Account": {"Name", "Industry"},
		}},
	}

	got, err := getSalesforceRequest(
		context.Background(), dependencies, withUsername(flowInstallation(true)),
		nil, nil, nil, "https://subscribe-webhook.example.com/v1/projects/p/integrations/i/installations/inst-1")
	if err != nil {
		t.Fatalf("getSalesforceRequest() error = %v", err)
	}

	req, ok := got.(*salesforce.SubscriptionRequest)
	if !ok {
		t.Fatalf("getSalesforceRequest() = %T, want *salesforce.SubscriptionRequest", got)
	}

	if !req.UseFlow {
		t.Error("UseFlow = false, want true")
	}

	if req.Flow == nil {
		t.Fatal("Flow = nil, want a config")
	}

	if req.Flow.EndpointURL != "https://subscribe-webhook.example.com/v1/projects/p/integrations/i/installations/inst-1" {
		t.Errorf("EndpointURL = %q, want the caller's webhook URL", req.Flow.EndpointURL)
	}

	if req.Flow.NamePrefix != "My App" {
		t.Errorf("NamePrefix = %q, want %q", req.Flow.NamePrefix, "My App")
	}

	if req.Flow.IntegrationUsername != "integration@example.com" {
		t.Errorf("IntegrationUsername = %q, want the value on the connection's provider metadata",
			req.Flow.IntegrationUsername)
	}

	if got, want := len(req.Flow.SelectedFields["Account"]), 2; got != want {
		t.Errorf("SelectedFields[Account] has %d entries, want %d", got, want)
	}

	// Flow mode is exclusive: the CDC quota-optimization payload must not be populated alongside,
	// or the subscribe would provision CDC artifacts for a subscription delivered by flows.
	if req.QuotaOptimizationObjectFields != nil {
		t.Errorf("QuotaOptimizationObjectFields = %v, want nil on a flow request",
			req.QuotaOptimizationObjectFields)
	}
}

// TestGetSalesforceRequestFlowModeTakesPrecedence verifies that flow mode wins when both resolvers
// would answer. A flow installation's events are delivered by outbound messages, so the CDC
// payload is not merely redundant — acting on it provisions the wrong artifacts.
func TestGetSalesforceRequestFlowModeTakesPrecedence(t *testing.T) {
	t.Parallel()

	dependencies := deps.Dependencies{
		Project:        stubProjectResolver{appName: "My App"},
		SalesforceFlow: stubSalesforceFlowResolver{},
		CDCOptimization: stubCDCOptimizationResolver{cfg: &deps.CDCOptimizationConfig{
			EnabledObjects: []common.ObjectName{"Account"},
		}},
	}

	got, err := getSalesforceRequest(
		context.Background(), dependencies, flowInstallation(true),
		nil, nil, nil, "https://subscribe-webhook.example.com/path")
	if err != nil {
		t.Fatalf("getSalesforceRequest() error = %v", err)
	}

	req, ok := got.(*salesforce.SubscriptionRequest)
	if !ok {
		t.Fatalf("getSalesforceRequest() = %T, want *salesforce.SubscriptionRequest", got)
	}

	if !req.UseFlow {
		t.Error("UseFlow = false, want true — flow mode must win over the CDC opt-in")
	}

	if req.QuotaOptimizationObjectFields != nil {
		t.Errorf("QuotaOptimizationObjectFields = %v, want nil", req.QuotaOptimizationObjectFields)
	}
}

// TestGetSalesforceRequestFlowNotEnabled verifies the mode comes from installation config, not
// from the resolver: an installation opted out falls through to CDC even with a flow resolver
// configured and returning inputs.
func TestGetSalesforceRequestFlowNotEnabled(t *testing.T) {
	t.Parallel()

	dependencies := deps.Dependencies{
		Project:        stubProjectResolver{appName: "My App"},
		SalesforceFlow: stubSalesforceFlowResolver{},
		CDCOptimization: stubCDCOptimizationResolver{cfg: &deps.CDCOptimizationConfig{
			EnabledObjects: []common.ObjectName{"Account"},
		}},
	}

	got, err := getSalesforceRequest(
		context.Background(), dependencies, flowInstallation(false),
		nil, nil, nil, "https://subscribe-webhook.example.com/path")
	if err != nil {
		t.Fatalf("getSalesforceRequest() error = %v", err)
	}

	req, ok := got.(*salesforce.SubscriptionRequest)
	if !ok {
		t.Fatalf("getSalesforceRequest() = %T, want *salesforce.SubscriptionRequest", got)
	}

	if req.UseFlow {
		t.Error("UseFlow = true, want false")
	}

	if len(req.QuotaOptimizationObjectFields) == 0 {
		t.Error("QuotaOptimizationObjectFields is empty, want the CDC payload to still be built")
	}
}

// TestGetSalesforceRequestFlowResolutionError verifies the contract that matters most: a failed
// mode resolution propagates. Falling through to CDC would provision CDC channels for a flow
// installation, and on an update tear down live flows and outbound messages.
func TestGetSalesforceRequestFlowResolutionError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("resolution failed")

	dependencies := deps.Dependencies{
		Project:        stubProjectResolver{appName: "My App"},
		SalesforceFlow: stubSalesforceFlowResolver{err: wantErr},
		CDCOptimization: stubCDCOptimizationResolver{cfg: &deps.CDCOptimizationConfig{
			EnabledObjects: []common.ObjectName{"Account"},
		}},
	}

	got, err := getSalesforceRequest(
		context.Background(), dependencies, flowInstallation(true),
		nil, nil, nil, "https://subscribe-webhook.example.com/path")
	if !errors.Is(err, wantErr) {
		t.Errorf("getSalesforceRequest() error = %v, want it to wrap %v", err, wantErr)
	}

	if got != nil {
		t.Errorf("getSalesforceRequest() = %v, want nil on a resolution failure", got)
	}
}

// TestGetSalesforceRequestFlowNoEndpoint verifies that flow mode without a webhook URL is an
// error. Salesforce rejects an outbound message with no endpoint, and deploying one that points
// nowhere leaves a subscription that silently delivers nothing.
func TestGetSalesforceRequestFlowNoEndpoint(t *testing.T) {
	t.Parallel()

	dependencies := deps.Dependencies{
		Project:        stubProjectResolver{appName: "My App"},
		SalesforceFlow: stubSalesforceFlowResolver{},
	}

	if _, err := getSalesforceRequest(
		context.Background(), dependencies, flowInstallation(true),
		nil, nil, nil, ""); err == nil {
		t.Error("getSalesforceRequest() error = nil, want an error when the webhook URL is empty")
	}
}

// TestGetSalesforceRequestFlowNoResolver verifies that a flow installation with no resolver is an
// error rather than a silent fall back to CDC. The config says events arrive by outbound message,
// so building a CDC request would provision artifacts delivering nothing it asked for.
func TestGetSalesforceRequestFlowNoResolver(t *testing.T) {
	t.Parallel()

	dependencies := deps.Dependencies{Project: stubProjectResolver{appName: "My App"}}

	if _, err := getSalesforceRequest(
		context.Background(), dependencies, flowInstallation(true),
		nil, nil, nil, "https://subscribe-webhook.example.com/path"); err == nil {
		t.Error("getSalesforceRequest() error = nil, want an error for a flow install with no resolver")
	}
}

// TestGetSalesforceRequestFlowNilSelectedFields verifies a nil map is permitted: the outbound
// messages still deploy, carrying only the always-included audit fields.
func TestGetSalesforceRequestFlowNilSelectedFields(t *testing.T) {
	t.Parallel()

	dependencies := deps.Dependencies{
		Project:        stubProjectResolver{appName: "My App"},
		SalesforceFlow: stubSalesforceFlowResolver{selectedFields: nil},
	}

	got, err := getSalesforceRequest(
		context.Background(), dependencies, flowInstallation(true),
		nil, nil, nil, "https://subscribe-webhook.example.com/path")
	if err != nil {
		t.Fatalf("getSalesforceRequest() error = %v", err)
	}

	req, ok := got.(*salesforce.SubscriptionRequest)
	if !ok {
		t.Fatalf("getSalesforceRequest() = %T, want *salesforce.SubscriptionRequest", got)
	}

	if !req.UseFlow || req.Flow == nil {
		t.Fatal("a nil map must still produce a flow request")
	}

	if req.Flow.NamePrefix != "My App" || req.Flow.EndpointURL == "" {
		t.Error("the values derived here must survive nil inputs")
	}

	if req.Flow.SelectedFields != nil {
		t.Error("a nil map must not be invented into fields")
	}
}
