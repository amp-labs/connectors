package metadata

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
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
			expected:   "amp_Lead",
		},
		{
			name:       "Custom object",
			objectName: "MyObject__c",
			expected:   "amp_MyObject__c",
		},
		{
			name:       "Empty object name",
			objectName: "",
			expected:   "amp_",
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
				TriggerName:       "amp_Lead",
				CheckboxFieldName: "AmpTriggerSubscription__c",
				WatchFields:       nil,
			},
			expectErr: errWatchFieldsEmpty,
		},
		{
			name: "Empty object name returns error",
			params: ApexTriggerParams{
				ObjectName:        "",
				TriggerName:       "amp_Lead",
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
				TriggerName:       "amp_Lead",
				CheckboxFieldName: "",
				WatchFields:       []string{"Email"},
			},
			expectErr: errRequiredParamsMet,
		},
		{
			name: "Valid params with single watch field",
			params: ApexTriggerParams{
				ObjectName:        "Lead",
				TriggerName:       "amp_Lead",
				CheckboxFieldName: "AmpTriggerSubscription__c",
				WatchFields:       []string{"Email"},
			},
			expectFiles: []string{
				"package.xml",
				"triggers/amp_Lead.trigger",
				"triggers/amp_Lead.trigger-meta.xml",
			},
		},
		{
			name: "Valid params with multiple watch fields",
			params: ApexTriggerParams{
				ObjectName:        "Contact",
				TriggerName:       "amp_Contact",
				CheckboxFieldName: "AmpTriggerSubscription__c",
				WatchFields:       []string{"Email", "Phone", "LastName"},
			},
			expectFiles: []string{
				"package.xml",
				"triggers/amp_Contact.trigger",
				"triggers/amp_Contact.trigger-meta.xml",
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
		TriggerName:       "amp_Lead",
		CheckboxFieldName: "AmpTriggerSubscription__c",
		WatchFields:       []string{"Email", "Phone"},
	}

	zipData, err := ConstructApexTrigger(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := readZipFiles(t, zipData)

	// Verify trigger code content.
	triggerCode, ok := files["triggers/amp_Lead.trigger"]
	if !ok {
		t.Fatal("trigger file not found in zip")
	}

	// Must reference the object name.
	if !strings.Contains(triggerCode, "trigger amp_Lead on Lead") {
		t.Error("trigger code missing trigger declaration with object name")
	}

	// Must contain both watch fields in insert conditions.
	if !strings.Contains(triggerCode, "rec.Email") {
		t.Error("trigger code missing Email field reference")
	}

	if !strings.Contains(triggerCode, "rec.Phone") {
		t.Error("trigger code missing Phone field reference")
	}

	// Must contain update conditions with oldRec references.
	if !strings.Contains(triggerCode, "oldRec.Email") {
		t.Error("trigger code missing oldRec.Email reference for update condition")
	}

	if !strings.Contains(triggerCode, "oldRec.Phone") {
		t.Error("trigger code missing oldRec.Phone reference for update condition")
	}

	// Must set the checkbox field.
	if !strings.Contains(triggerCode, "rec.AmpTriggerSubscription__c = fieldChanged") {
		t.Error("trigger code missing checkbox field assignment")
	}

	// Verify meta XML content.
	metaXML, ok := files["triggers/amp_Lead.trigger-meta.xml"]
	if !ok {
		t.Fatal("trigger meta XML file not found in zip")
	}

	if !strings.Contains(metaXML, "<apiVersion>61.0</apiVersion>") {
		t.Error("meta XML missing correct API version")
	}

	if !strings.Contains(metaXML, "<status>Active</status>") {
		t.Error("meta XML missing Active status")
	}

	// Verify package.xml content.
	packageXML, ok := files["package.xml"]
	if !ok {
		t.Fatal("package.xml not found in zip")
	}

	if !strings.Contains(packageXML, "amp_Lead") {
		t.Error("package.xml missing trigger member name")
	}

	if !strings.Contains(packageXML, "ApexTrigger") {
		t.Error("package.xml missing ApexTrigger type")
	}
}

func TestConstructDestructiveApexTrigger(t *testing.T) {
	t.Parallel()

	triggerName := "amp_Lead"

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
