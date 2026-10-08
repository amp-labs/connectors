package metadata

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func validOutboundMessageParams() OutboundMessageParams {
	return OutboundMessageParams{
		ObjectName:          "Account",
		Name:                "amp_Account",
		EndpointURL:         "https://example.com/webhook",
		IntegrationUsername: "integration@example.com",
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

func TestGenerateWorkflowXMLAlwaysIncludesRequiredFields(t *testing.T) {
	t.Parallel()

	// No fields configured: the required set alone makes up the payload.
	workflow, err := generateWorkflowXML(validOutboundMessageParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{
		"<fields>Id</fields>",
		"<fields>CreatedDate</fields>",
		"<fields>LastModifiedDate</fields>",
	} {
		if !strings.Contains(workflow, want) {
			t.Errorf("workflow XML missing required field %q", want)
		}
	}

	// Caller-configured fields must not displace the required set, and must
	// not be duplicated when they overlap with it.
	params := validOutboundMessageParams()
	params.Fields = []string{"Name", "CreatedDate"}

	workflow, err = generateWorkflowXML(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

func TestGenerateWorkflowXMLDedupesFieldsCaseInsensitively(t *testing.T) {
	t.Parallel()

	// Salesforce field names are case-insensitive, so a lowercase "id" next to
	// the required "Id" is a duplicate field and fails every delivery.
	params := validOutboundMessageParams()
	params.Fields = []string{"id", "createddate", "LASTMODIFIEDDATE", "Name"}

	workflow, err := generateWorkflowXML(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lower := strings.ToLower(workflow)

	for _, field := range []string{"id", "createddate", "lastmodifieddate", "name"} {
		if got := strings.Count(lower, "<fields>"+field+"</fields>"); got != 1 {
			t.Errorf("field %q appears %d times, want 1:\n%s", field, got, workflow)
		}
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

	pkg, ok := entries["package.xml"]
	if !ok {
		t.Fatal("destructive zip requires package.xml")
	}

	if strings.Contains(pkg, "<types>") || strings.Contains(pkg, "<members>") {
		t.Errorf("package.xml must list no components to deploy:\n%s", pkg)
	}
}

func keysOf(entries map[string]string) []string {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}

	return keys
}
