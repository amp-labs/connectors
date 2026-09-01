package metadata

import (
	"strings"
	"testing"
)

func TestGenerateFlowNameForSubscription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		object    string
		expected  string
		expectErr bool
	}{
		{
			name:     "Standard object",
			object:   "Lead",
			expected: "AmpSubscribe_Lead",
		},
		{
			name:     "Custom object collapses consecutive underscores",
			object:   "My_Object__c",
			expected: "AmpSubscribe_My_Object_c",
		},
		{
			name:      "Empty object name returns error",
			object:    "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := GenerateFlowNameForSubscription(tt.object)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("GenerateFlowNameForSubscription(%q) = %q, want %q", tt.object, got, tt.expected)
			}
		})
	}
}

func validFlowParams() FlowParams {
	return FlowParams{
		ObjectName:          "Account",
		FlowName:            "AmpSubscribe_Account",
		OutboundMessageName: "amp_Account",
		RecordTriggerType:   RecordTriggerTypeCreateAndUpdate,
	}
}

func TestValidateFlowParams(t *testing.T) {
	t.Parallel()

	if err := ValidateFlowParams(validFlowParams()); err != nil {
		t.Fatalf("unexpected error for valid params: %v", err)
	}

	missingFlowName := validFlowParams()
	missingFlowName.FlowName = ""

	if err := ValidateFlowParams(missingFlowName); err == nil {
		t.Error("expected error for missing flow name")
	}

	missingOMName := validFlowParams()
	missingOMName.OutboundMessageName = ""

	if err := ValidateFlowParams(missingOMName); err == nil {
		t.Error("expected error for missing outbound message name")
	}

	badTrigger := validFlowParams()
	badTrigger.RecordTriggerType = "Delete"

	if err := ValidateFlowParams(badTrigger); err == nil {
		t.Error("expected error for unsupported record trigger type")
	}
}

func TestBuildWatchFieldsFilterFormula(t *testing.T) {
	t.Parallel()

	params := validFlowParams()
	params.RecordTriggerType = RecordTriggerTypeUpdate
	params.WatchFields = []string{"Email"}

	if got := buildWatchFieldsFilterFormula(params); got != "ISCHANGED({!$Record.Email})" {
		t.Errorf("single watch field formula = %q", got)
	}

	params.WatchFields = []string{"Email", "Phone"}

	expected := "OR(ISCHANGED({!$Record.Email}), ISCHANGED({!$Record.Phone}))"
	if got := buildWatchFieldsFilterFormula(params); got != expected {
		t.Errorf("multi watch field formula = %q, want %q", got, expected)
	}

	// ISCHANGED is invalid in flows that also fire on Create.
	params.RecordTriggerType = RecordTriggerTypeCreateAndUpdate
	if got := buildWatchFieldsFilterFormula(params); got != "" {
		t.Errorf("expected no formula for CreateAndUpdate, got %q", got)
	}

	params.RecordTriggerType = RecordTriggerTypeUpdate
	params.WatchFields = nil

	if got := buildWatchFieldsFilterFormula(params); got != "" {
		t.Errorf("expected no formula without watch fields, got %q", got)
	}
}

func TestConstructFlowActive(t *testing.T) {
	t.Parallel()

	zipData, err := ConstructFlow(validFlowParams(), "Active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := readZipEntries(t, zipData)

	pkg, ok := entries["package.xml"]
	if !ok {
		t.Fatalf("package.xml missing; entries: %v", keysOf(entries))
	}

	for _, want := range []string{
		"<name>Flow</name>",
		"<members>AmpSubscribe_Account</members>",
	} {
		if !strings.Contains(pkg, want) {
			t.Errorf("package.xml missing %q:\n%s", want, pkg)
		}
	}

	if strings.Contains(pkg, "WorkflowOutboundMessage") {
		t.Errorf("flow package.xml must not carry the outbound message:\n%s", pkg)
	}

	if _, exists := entries["workflows/Account.workflow"]; exists {
		t.Error("flow zip must not carry the workflow file")
	}

	flow, ok := entries["flows/AmpSubscribe_Account.flow"]
	if !ok {
		t.Fatalf("flow file missing; entries: %v", keysOf(entries))
	}

	for _, want := range []string{
		"<actionName>Account.amp_Account</actionName>",
		"<actionType>outboundMessage</actionType>",
		"<object>Account</object>",
		"<recordTriggerType>CreateAndUpdate</recordTriggerType>",
		"<triggerType>RecordAfterSave</triggerType>",
		"<status>Active</status>",
		"<processType>AutoLaunchedFlow</processType>",
	} {
		if !strings.Contains(flow, want) {
			t.Errorf("flow XML missing %q:\n%s", want, flow)
		}
	}

	if strings.Contains(flow, "<filterFormula>") {
		t.Errorf("flow XML must not carry a filter formula for CreateAndUpdate:\n%s", flow)
	}
}

func TestConstructFlowUpdateOnlyWatchFields(t *testing.T) {
	t.Parallel()

	params := validFlowParams()
	params.RecordTriggerType = RecordTriggerTypeUpdate
	params.WatchFields = []string{"Email", "Phone"}

	zipData, err := ConstructFlow(params, "Active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := readZipEntries(t, zipData)

	flow := entries["flows/AmpSubscribe_Account.flow"]
	want := "<filterFormula>OR(ISCHANGED({!$Record.Email}), ISCHANGED({!$Record.Phone}))</filterFormula>"

	if !strings.Contains(flow, want) {
		t.Errorf("flow XML missing %q:\n%s", want, flow)
	}
}

func TestConstructFlowDraft(t *testing.T) {
	t.Parallel()

	zipData, err := ConstructFlow(validFlowParams(), "Draft")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := readZipEntries(t, zipData)

	flow := entries["flows/AmpSubscribe_Account.flow"]
	if !strings.Contains(flow, "<status>Draft</status>") {
		t.Errorf("flow XML missing Draft status:\n%s", flow)
	}
}

func TestConstructDestructiveFlow(t *testing.T) {
	t.Parallel()

	zipData, err := ConstructDestructiveFlow("AmpSubscribe_Account")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := readZipEntries(t, zipData)

	destructive, ok := entries["destructiveChanges.xml"]
	if !ok {
		t.Fatalf("destructiveChanges.xml missing; entries: %v", keysOf(entries))
	}

	if !strings.Contains(destructive, "<members>AmpSubscribe_Account</members>") ||
		!strings.Contains(destructive, "<name>Flow</name>") {
		t.Errorf("destructiveChanges.xml missing flow member:\n%s", destructive)
	}

	if _, ok := entries["package.xml"]; !ok {
		t.Error("destructive zip requires an empty package.xml")
	}

	if _, err := ConstructDestructiveFlow(""); err == nil {
		t.Error("expected error for empty flow name")
	}
}

// subscriptionComponentFor builds a consistent component for the given object.
func subscriptionComponentFor(objectName string) FlowSubscriptionComponent {
	omName := "amp_" + objectName
	flowName := "AmpSubscribe_" + objectName

	return FlowSubscriptionComponent{
		OutboundMessage: OutboundMessageParams{
			ObjectName:          objectName,
			Name:                omName,
			EndpointURL:         "https://example.com/webhook",
			IntegrationUsername: "integration@example.com",
		},
		Flow: FlowParams{
			ObjectName:          objectName,
			FlowName:            flowName,
			OutboundMessageName: omName,
			RecordTriggerType:   RecordTriggerTypeCreateAndUpdate,
		},
	}
}

func TestConstructFlowSubscription(t *testing.T) {
	t.Parallel()

	// The whole subscription (multiple objects) ships as ONE package.
	zipData, err := ConstructFlowSubscription([]FlowSubscriptionComponent{
		subscriptionComponentFor("Account"),
		subscriptionComponentFor("Contact"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := readZipEntries(t, zipData)

	pkg := entries["package.xml"]
	for _, want := range []string{
		"<name>WorkflowOutboundMessage</name>",
		"<members>Account.amp_Account</members>",
		"<members>Contact.amp_Contact</members>",
		"<name>Flow</name>",
		"<members>AmpSubscribe_Account</members>",
		"<members>AmpSubscribe_Contact</members>",
	} {
		if !strings.Contains(pkg, want) {
			t.Errorf("package.xml missing %q:\n%s", want, pkg)
		}
	}

	for _, file := range []string{
		"workflows/Account.workflow",
		"workflows/Contact.workflow",
		"flows/AmpSubscribe_Account.flow",
		"flows/AmpSubscribe_Contact.flow",
	} {
		if _, ok := entries[file]; !ok {
			t.Errorf("package missing %s; entries: %v", file, keysOf(entries))
		}
	}

	if !strings.Contains(entries["flows/AmpSubscribe_Account.flow"], "<status>Active</status>") {
		t.Error("package must deploy flows as Active")
	}
}

func TestConstructFlowSubscriptionValidation(t *testing.T) {
	t.Parallel()

	if _, err := ConstructFlowSubscription(nil); err == nil {
		t.Error("expected error for empty component list")
	}

	mismatchedName := subscriptionComponentFor("Account")
	mismatchedName.Flow.OutboundMessageName = "amp_Other"

	if _, err := ConstructFlowSubscription([]FlowSubscriptionComponent{mismatchedName}); err == nil {
		t.Error("expected error for outbound message name mismatch")
	}

	mismatchedObject := subscriptionComponentFor("Account")
	mismatchedObject.Flow.ObjectName = "Contact"
	mismatchedObject.Flow.FlowName = "AmpSubscribe_Contact"

	if _, err := ConstructFlowSubscription([]FlowSubscriptionComponent{mismatchedObject}); err == nil {
		t.Error("expected error for object name mismatch")
	}

	if _, err := ConstructFlowSubscription([]FlowSubscriptionComponent{
		subscriptionComponentFor("Account"),
		subscriptionComponentFor("Account"),
	}); err == nil {
		t.Error("expected error for duplicate object")
	}
}
