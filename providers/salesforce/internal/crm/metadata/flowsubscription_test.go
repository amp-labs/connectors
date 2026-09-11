package metadata

import (
	"strings"
	"testing"
)

func validFlowParams() FlowParams {
	return FlowParams{
		ObjectName:          "Account",
		FlowName:            "acme_Account",
		OutboundMessageName: "amp_Account",
		RecordTriggerType:   RecordTriggerTypeCreateAndUpdate,
	}
}

func TestValidateFlowParamsRejectsDelete(t *testing.T) {
	t.Parallel()

	params := validFlowParams()
	params.RecordTriggerType = "Delete"

	if err := ValidateFlowParams(params); err == nil {
		t.Error("expected error for Delete trigger; outbound messages cannot run on delete")
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

	// Create+update: ISNEW so creates still fire; ISCHANGED gates updates.
	params.RecordTriggerType = RecordTriggerTypeCreateAndUpdate
	wantCreateAndUpdate := "OR(ISNEW(), ISCHANGED({!$Record.Email}), ISCHANGED({!$Record.Phone}))"
	if got := buildWatchFieldsFilterFormula(params); got != wantCreateAndUpdate {
		t.Errorf("CreateAndUpdate formula = %q, want %q", got, wantCreateAndUpdate)
	}

	params.RecordTriggerType = RecordTriggerTypeCreate
	if got := buildWatchFieldsFilterFormula(params); got != "" {
		t.Errorf("expected no formula for Create-only, got %q", got)
	}

	params.RecordTriggerType = RecordTriggerTypeUpdate
	params.WatchFields = nil

	if got := buildWatchFieldsFilterFormula(params); got != "" {
		t.Errorf("expected no formula without watch fields, got %q", got)
	}
}

func TestConstructDraftFlow(t *testing.T) {
	t.Parallel()

	zipData, err := ConstructDraftFlow(validFlowParams())
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
		"<members>acme_Account</members>",
	} {
		if !strings.Contains(pkg, want) {
			t.Errorf("package.xml missing %q:\n%s", want, pkg)
		}
	}

	if strings.Contains(pkg, "WorkflowOutboundMessage") {
		t.Errorf("draft zip must not carry the outbound message:\n%s", pkg)
	}

	if _, exists := entries["workflows/Account.workflow"]; exists {
		t.Error("draft zip must not carry the workflow file")
	}

	flow, ok := entries["flows/acme_Account.flow"]
	if !ok {
		t.Fatalf("flow file missing; entries: %v", keysOf(entries))
	}

	for _, want := range []string{
		"<actionName>Account.amp_Account</actionName>",
		"<actionType>outboundMessage</actionType>",
		"<object>Account</object>",
		"<recordTriggerType>CreateAndUpdate</recordTriggerType>",
		"<triggerType>RecordAfterSave</triggerType>",
		"<status>Draft</status>",
		"<processType>AutoLaunchedFlow</processType>",
	} {
		if !strings.Contains(flow, want) {
			t.Errorf("flow XML missing %q:\n%s", want, flow)
		}
	}

	if strings.Contains(flow, "<filterFormula>") {
		t.Errorf("flow XML must not carry a filter formula without watch fields:\n%s", flow)
	}
}

func TestConstructDestructiveFlow(t *testing.T) {
	t.Parallel()

	zipData, err := ConstructDestructiveFlow("acme_Account", []int{1, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	destructive := readZipEntries(t, zipData)["destructiveChanges.xml"]
	for _, want := range []string{
		"<members>acme_Account-1</members>",
		"<members>acme_Account-2</members>",
		"<name>Flow</name>",
	} {
		if !strings.Contains(destructive, want) {
			t.Errorf("destructiveChanges.xml missing %q:\n%s", want, destructive)
		}
	}

	if _, err := ConstructDestructiveFlow("acme_Account", nil); err == nil {
		t.Error("expected error for empty version list")
	}
}

// subscriptionComponentFor builds a consistent component for the given object.
func subscriptionComponentFor(objectName string) FlowSubscriptionComponent {
	omName := "amp_" + objectName
	flowName := "acme_" + objectName

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
		"<members>acme_Account</members>",
		"<members>acme_Contact</members>",
	} {
		if !strings.Contains(pkg, want) {
			t.Errorf("package.xml missing %q:\n%s", want, pkg)
		}
	}

	for _, file := range []string{
		"workflows/Account.workflow",
		"workflows/Contact.workflow",
		"flows/acme_Account.flow",
		"flows/acme_Contact.flow",
	} {
		if _, ok := entries[file]; !ok {
			t.Errorf("package missing %s; entries: %v", file, keysOf(entries))
		}
	}

	if !strings.Contains(entries["flows/acme_Account.flow"], "<status>Active</status>") {
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
	mismatchedObject.Flow.FlowName = "acme_Contact"

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
