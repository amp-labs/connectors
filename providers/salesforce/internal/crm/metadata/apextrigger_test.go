package metadata

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

func TestGenerateApexTriggerName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		objectName string
		expected   string
	}{
		{
			name:       "Standard object",
			objectName: "Lead",
			expected:   "Lead",
		},
		{
			name:       "Custom object",
			objectName: "MyObject__c",
			expected:   "MyObject__c",
		},
		{
			name:       "Empty object name",
			objectName: "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GenerateApexTriggerName(tt.objectName)
			if got != tt.expected {
				t.Errorf("GenerateApexTriggerName(%q) = %q, want %q", tt.objectName, got, tt.expected)
			}
		})
	}
}

func TestConstructApexTrigger(t *testing.T) { //nolint:funlen,cyclop
	t.Parallel()

	tests := []struct {
		name        string
		params      ApexTriggerParams
		expectErr   error
		expectFiles []string // expected file names inside the zip
	}{
		{
			name: "Empty watch fields returns error",
			params: ApexTriggerParams{
				ObjectName:        "Lead",
				TriggerName:       "Lead",
				CheckboxFieldName: "AmpTriggerSubscription__c",
				WatchFields:       nil,
			},
			expectErr: errWatchFieldsEmpty,
		},
		{
			name: "Empty object name returns error",
			params: ApexTriggerParams{
				ObjectName:        "",
				TriggerName:       "Lead",
				CheckboxFieldName: "AmpTriggerSubscription__c",
				WatchFields:       []string{"Email"},
			},
			expectErr: errRequiredParamsMet,
		},
		{
			name: "Empty trigger name returns error",
			params: ApexTriggerParams{
				ObjectName:        "Lead",
				TriggerName:       "",
				CheckboxFieldName: "AmpTriggerSubscription__c",
				WatchFields:       []string{"Email"},
			},
			expectErr: errRequiredParamsMet,
		},
		{
			name: "Empty checkbox field name returns error",
			params: ApexTriggerParams{
				ObjectName:        "Lead",
				TriggerName:       "Lead",
				CheckboxFieldName: "",
				WatchFields:       []string{"Email"},
			},
			expectErr: errRequiredParamsMet,
		},
		{
			name: "Valid params with single watch field",
			params: ApexTriggerParams{
				ObjectName:        "Lead",
				TriggerName:       "Lead",
				CheckboxFieldName: "AmpTriggerSubscription__c",
				WatchFields:       []string{"Email"},
			},
			expectFiles: []string{
				"package.xml",
				"triggers/Lead.trigger",
				"triggers/Lead.trigger-meta.xml",
			},
		},
		{
			name: "Valid params with multiple watch fields",
			params: ApexTriggerParams{
				ObjectName:        "Contact",
				TriggerName:       "Contact",
				CheckboxFieldName: "AmpTriggerSubscription__c",
				WatchFields:       []string{"Email", "Phone", "LastName"},
			},
			expectFiles: []string{
				"package.xml",
				"triggers/Contact.trigger",
				"triggers/Contact.trigger-meta.xml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			zipData, err := ConstructApexTrigger(tt.params)

			if tt.expectErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectErr)
				}

				if err != tt.expectErr {
					t.Fatalf("expected error %v, got %v", tt.expectErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertZipContainsFiles(t, zipData, tt.expectFiles)
		})
	}
}

func TestConstructApexTriggerContent(t *testing.T) { //nolint:funlen
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:        "Lead",
		TriggerName:       "Lead",
		CheckboxFieldName: "AmpTriggerSubscription__c",
		WatchFields:       []string{"Email", "Phone"},
	}

	zipData, err := ConstructApexTrigger(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := readZipFiles(t, zipData)

	// Verify trigger code content.
	triggerCode, ok := files["triggers/Lead.trigger"]
	if !ok {
		t.Fatal("trigger file not found in zip")
	}

	expectedTriggerCode := `trigger Lead on Lead (before insert, before update) {
    if (Trigger.isBefore) {
        for (Lead rec : Trigger.new) {
            Boolean fieldChanged = false;

            if (Trigger.isInsert) {
                fieldChanged = (rec.Email != null && rec.Email != '') || (rec.Phone != null && rec.Phone != '');
            } else if (Trigger.isUpdate) {
                Lead oldRec = Trigger.oldMap.get(rec.Id);
                fieldChanged = (rec.Email != oldRec.Email) || (rec.Phone != oldRec.Phone);
            }

            rec.AmpTriggerSubscription__c = fieldChanged;
        }
    }
}
`
	if triggerCode != expectedTriggerCode {
		t.Errorf("trigger code mismatch.\nGot:\n%s\nWant:\n%s", triggerCode, expectedTriggerCode)
	}

	// Verify meta XML content.
	metaXML, ok := files["triggers/Lead.trigger-meta.xml"]
	if !ok {
		t.Fatal("trigger meta XML file not found in zip")
	}

	expectedMetaXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ApexTrigger xmlns="http://soap.sforce.com/2006/04/metadata">
    <apiVersion>%s</apiVersion>
    <status>Active</status>
</ApexTrigger>
`, core.APIVersion)
	if metaXML != expectedMetaXML {
		t.Errorf("meta XML mismatch.\nGot:\n%s\nWant:\n%s", metaXML, expectedMetaXML)
	}

	// Verify package.xml content.
	packageXML, ok := files["package.xml"]
	if !ok {
		t.Fatal("package.xml not found in zip")
	}

	expectedPackageXML := xml.Header + fmt.Sprintf(`<Package xmlns="http://soap.sforce.com/2006/04/metadata">
    <types>
        <members>Lead</members>
        <name>ApexTrigger</name>
    </types>
    <version>%s</version>
</Package>`, core.APIVersion)
	if packageXML != expectedPackageXML {
		t.Errorf("package.xml mismatch.\nGot:\n%s\nWant:\n%s", packageXML, expectedPackageXML)
	}
}

func TestConstructDestructiveApexTrigger(t *testing.T) {
	t.Parallel()

	triggerName := "Lead"

	zipData, err := ConstructDestructiveApexTrigger(triggerName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFiles := []string{
		"package.xml",
		"destructiveChanges.xml",
	}
	assertZipContainsFiles(t, zipData, expectedFiles)

	files := readZipFiles(t, zipData)

	// The destructiveChanges.xml must reference the trigger.
	destructiveXML, ok := files["destructiveChanges.xml"]
	if !ok {
		t.Fatal("destructiveChanges.xml not found in zip")
	}

	if !strings.Contains(destructiveXML, triggerName) {
		t.Error("destructiveChanges.xml missing trigger name")
	}

	if !strings.Contains(destructiveXML, "ApexTrigger") {
		t.Error("destructiveChanges.xml missing ApexTrigger type")
	}

	// The package.xml should be empty (no types with members).
	packageXML, ok := files["package.xml"]
	if !ok {
		t.Fatal("package.xml not found in zip")
	}

	if strings.Contains(packageXML, triggerName) {
		t.Error("package.xml should not contain the trigger name for destructive changes")
	}
}

// assertZipContainsFiles verifies that the zip data contains exactly the expected files.
func assertZipContainsFiles(t *testing.T, zipData []byte, expectedFiles []string) {
	t.Helper()

	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}

	fileNames := make(map[string]bool)
	for _, f := range reader.File {
		fileNames[f.Name] = true
	}

	for _, expected := range expectedFiles {
		if !fileNames[expected] {
			t.Errorf("expected file %q not found in zip, got files: %v", expected, fileNames)
		}
	}

	if len(reader.File) != len(expectedFiles) {
		t.Errorf("expected %d files in zip, got %d", len(expectedFiles), len(reader.File))
	}
}

// readZipFiles reads all files from a zip and returns a map of filename -> content.
func readZipFiles(t *testing.T, zipData []byte) map[string]string {
	t.Helper()

	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}

	files := make(map[string]string)

	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("failed to open zip entry %s: %v", f.Name, err)
		}

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(rc); err != nil {
			rc.Close()
			t.Fatalf("failed to read zip entry %s: %v", f.Name, err)
		}

		rc.Close()

		files[f.Name] = buf.String()
	}

	return files
}
