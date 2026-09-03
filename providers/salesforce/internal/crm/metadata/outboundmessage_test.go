package metadata

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestGenerateOutboundMessageNameForSubscription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		object    string
		expected  string
		expectErr bool
	}{
		{
			name:     "Standard object",
			object:   "Account",
			expected: "amp_Account",
		},
		{
			name:     "Custom object collapses consecutive underscores",
			object:   "My_Object__c",
			expected: "amp_My_Object_c",
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

			got, err := GenerateOutboundMessageNameForSubscription(tt.object)
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
				t.Errorf("GenerateOutboundMessageNameForSubscription(%q) = %q, want %q", tt.object, got, tt.expected)
			}
		})
	}
}

func validOutboundMessageParams() OutboundMessageParams {
	return OutboundMessageParams{
		ObjectName:          "Account",
		Name:                "amp_Account",
		EndpointURL:         "https://example.com/webhook",
		IntegrationUsername: "integration@example.com",
	}
}

func TestValidateOutboundMessageParams(t *testing.T) {
	t.Parallel()

	if err := ValidateOutboundMessageParams(validOutboundMessageParams()); err != nil {
		t.Fatalf("unexpected error for valid params: %v", err)
	}

	missingEndpoint := validOutboundMessageParams()
	missingEndpoint.EndpointURL = ""

	if err := ValidateOutboundMessageParams(missingEndpoint); err == nil {
		t.Error("expected error for missing endpoint URL")
	}

	missingUser := validOutboundMessageParams()
	missingUser.IntegrationUsername = ""

	if err := ValidateOutboundMessageParams(missingUser); err == nil {
		t.Error("expected error for missing integration username")
	}

	missingName := validOutboundMessageParams()
	missingName.Name = ""

	if err := ValidateOutboundMessageParams(missingName); err == nil {
		t.Error("expected error for missing outbound message name")
	}
}

// readZipEntries unpacks an in-memory zip into name → content.
func readZipEntries(t *testing.T, zipData []byte) map[string]string {
	t.Helper()

	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}

	entries := make(map[string]string)

	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("failed to open zip entry %s: %v", file.Name, err)
		}

		content, err := io.ReadAll(rc)
		rc.Close()

		if err != nil {
			t.Fatalf("failed to read zip entry %s: %v", file.Name, err)
		}

		entries[file.Name] = string(content)
	}

	return entries
}

func TestConstructOutboundMessage(t *testing.T) {
	t.Parallel()

	params := validOutboundMessageParams()
	params.Fields = []string{"Id", "Name"}

	zipData, err := ConstructOutboundMessage(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := readZipEntries(t, zipData)

	pkg, ok := entries["package.xml"]
	if !ok {
		t.Fatalf("package.xml missing; entries: %v", keysOf(entries))
	}

	for _, want := range []string{
		"<name>WorkflowOutboundMessage</name>",
		"<members>Account.amp_Account</members>",
	} {
		if !strings.Contains(pkg, want) {
			t.Errorf("package.xml missing %q:\n%s", want, pkg)
		}
	}

	workflow, ok := entries["workflows/Account.workflow"]
	if !ok {
		t.Fatalf("workflow file missing; entries: %v", keysOf(entries))
	}

	for _, want := range []string{
		"<fullName>amp_Account</fullName>",
		"<endpointUrl>https://example.com/webhook</endpointUrl>",
		"<integrationUser>integration@example.com</integrationUser>",
		"<fields>Id</fields>",
		"<fields>Name</fields>",
		"<includeSessionId>false</includeSessionId>",
	} {
		if !strings.Contains(workflow, want) {
			t.Errorf("workflow XML missing %q:\n%s", want, workflow)
		}
	}
}

func TestConstructOutboundMessageAlwaysIncludesRequiredFields(t *testing.T) {
	t.Parallel()

	// No fields configured: the required set alone makes up the payload.
	zipData, err := ConstructOutboundMessage(validOutboundMessageParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := readZipEntries(t, zipData)

	for _, want := range []string{
		"<fields>Id</fields>",
		"<fields>CreatedDate</fields>",
		"<fields>LastModifiedDate</fields>",
	} {
		if !strings.Contains(entries["workflows/Account.workflow"], want) {
			t.Errorf("workflow XML missing required field %q", want)
		}
	}

	// Caller-configured fields must not displace the required set, and must
	// not be duplicated when they overlap with it.
	params := validOutboundMessageParams()
	params.Fields = []string{"Name", "CreatedDate"}

	zipData, err = ConstructOutboundMessage(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	workflow := readZipEntries(t, zipData)["workflows/Account.workflow"]

	for _, want := range []string{
		"<fields>Id</fields>",
		"<fields>LastModifiedDate</fields>",
		"<fields>Name</fields>",
	} {
		if !strings.Contains(workflow, want) {
			t.Errorf("workflow XML missing field %q:\n%s", want, workflow)
		}
	}

	if strings.Count(workflow, "<fields>CreatedDate</fields>") != 1 {
		t.Errorf("CreatedDate must appear exactly once:\n%s", workflow)
	}
}

func TestConstructDestructiveOutboundMessage(t *testing.T) {
	t.Parallel()

	zipData, err := ConstructDestructiveOutboundMessage("Account", "amp_Account")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := readZipEntries(t, zipData)

	destructive, ok := entries["destructiveChanges.xml"]
	if !ok {
		t.Fatalf("destructiveChanges.xml missing; entries: %v", keysOf(entries))
	}

	if !strings.Contains(destructive, "<members>Account.amp_Account</members>") ||
		!strings.Contains(destructive, "<name>WorkflowOutboundMessage</name>") {
		t.Errorf("destructiveChanges.xml missing outbound message member:\n%s", destructive)
	}

	if _, ok := entries["package.xml"]; !ok {
		t.Error("destructive zip requires an empty package.xml")
	}

	if _, err := ConstructDestructiveOutboundMessage("", "amp_Account"); err == nil {
		t.Error("expected error for empty object name")
	}

	if _, err := ConstructDestructiveOutboundMessage("Account", ""); err == nil {
		t.Error("expected error for empty outbound message name")
	}
}

func keysOf(entries map[string]string) []string {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}

	return keys
}
