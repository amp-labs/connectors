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

func TestGenerateApexTriggerNameForCDC(t *testing.T) {
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
			expected: "CDC_Lead",
		},
		{
			name:     "Custom object",
			object:   "MyObject__c",
			expected: "CDC_MyObject__c",
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

			got, err := GenerateApexTriggerNameForCDC(tt.object)
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
				t.Errorf("GenerateApexTriggerNameForCDC(%q) = %q, want %q", tt.object, got, tt.expected)
			}
		})
	}
}

func TestGenerateApexTriggerNameForRead(t *testing.T) {
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
			expected: "Read_Lead",
		},
		{
			name:     "Custom object",
			object:   "MyObject__c",
			expected: "Read_MyObject__c",
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

			got, err := GenerateApexTriggerNameForRead(tt.object)
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
				t.Errorf("GenerateApexTriggerNameForRead(%q) = %q, want %q", tt.object, got, tt.expected)
			}
		})
	}
}

func TestValidateApexTriggerParams(t *testing.T) { //nolint:funlen,cyclop
	t.Parallel()

	tests := []struct {
		name               string
		params             ApexTriggerParams
		indicatorFieldName string
		expectErr          error
	}{
		{
			name: "Empty watch fields returns error",
			params: ApexTriggerParams{
				ObjectName:  "Lead",
				TriggerName: "Lead",
				WatchFields: nil,
			},
			indicatorFieldName: "AmpTriggerSubscription__c",
			expectErr:          errWatchFieldsEmpty,
		},
		{
			name: "Empty object name returns error",
			params: ApexTriggerParams{
				ObjectName:  "",
				TriggerName: "Lead",
				WatchFields: []string{"Email"},
			},
			indicatorFieldName: "AmpTriggerSubscription__c",
			expectErr:          errRequiredParamsMet,
		},
		{
			name: "Empty trigger name returns error",
			params: ApexTriggerParams{
				ObjectName:  "Lead",
				TriggerName: "",
				WatchFields: []string{"Email"},
			},
			indicatorFieldName: "AmpTriggerSubscription__c",
			expectErr:          errRequiredParamsMet,
		},
		{
			name: "Empty indicator field name returns error",
			params: ApexTriggerParams{
				ObjectName:  "Lead",
				TriggerName: "Lead",
				WatchFields: []string{"Email"},
			},
			indicatorFieldName: "",
			expectErr:          errRequiredParamsMet,
		},
		{
			name: "Valid params with single watch field",
			params: ApexTriggerParams{
				ObjectName:  "Lead",
				TriggerName: "Lead",
				WatchFields: []string{"Email"},
			},
			indicatorFieldName: "AmpTriggerSubscription__c",
		},
		{
			name: "Valid params with multiple watch fields",
			params: ApexTriggerParams{
				ObjectName:  "Contact",
				TriggerName: "Contact",
				WatchFields: []string{"Email", "Phone", "LastName"},
			},
			indicatorFieldName: "AmpTriggerSubscription__c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateApexTriggerParams(tt.params, tt.indicatorFieldName)

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
		})
	}
}

func TestConstructApexTriggerZip(t *testing.T) {
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Lead",
		TriggerName: "Lead",
		WatchFields: []string{"Email"},
	}

	triggerCode := GenerateTriggerCodeForCDC(params, "AmpTriggerSubscription__c")

	zipData, err := ConstructApexTrigger(params, triggerCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertZipContainsFiles(t, zipData, []string{
		"package.xml",
		"triggers/Lead.trigger",
		"triggers/Lead.trigger-meta.xml",
		"classes/Test_Lead.cls",
		"classes/Test_Lead.cls-meta.xml",
	})
}

func TestConstructApexTriggerForCDCContent(t *testing.T) { //nolint:funlen
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Lead",
		TriggerName: "Lead",
		WatchFields: []string{"Email", "Phone"},
	}

	triggerCode := GenerateTriggerCodeForCDC(params, "AmpTriggerSubscription__c")

	zipData, err := ConstructApexTrigger(params, triggerCode)
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
                fieldChanged = (rec.Email != null) || (rec.Phone != null);
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
    <types>
        <members>Test_Lead</members>
        <name>ApexClass</name>
    </types>
    <version>%s</version>
</Package>`, core.APIVersion)
	if packageXML != expectedPackageXML {
		t.Errorf("package.xml mismatch.\nGot:\n%s\nWant:\n%s", packageXML, expectedPackageXML)
	}
}

func TestConstructApexTriggerForFilteredReadContent(t *testing.T) { //nolint:funlen
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Lead",
		TriggerName: "Lead",
		WatchFields: []string{"Email", "Phone"},
	}

	triggerCode := GenerateTriggerCodeForFilteredRead(params, "AmpTimestamp__c")

	zipData, err := ConstructApexTrigger(params, triggerCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := readZipFiles(t, zipData)

	triggerCode, ok := files["triggers/Lead.trigger"]
	if !ok {
		t.Fatal("trigger file not found in zip")
	}

	expectedTriggerCode := `trigger Lead on Lead (before insert, before update) {
    if (Trigger.isBefore) {
        for (Lead rec : Trigger.new) {
            Boolean fieldChanged = false;

            if (Trigger.isInsert) {
                fieldChanged = (rec.Email != null) || (rec.Phone != null);
            } else if (Trigger.isUpdate) {
                Lead oldRec = Trigger.oldMap.get(rec.Id);
                fieldChanged = (rec.Email != oldRec.Email) || (rec.Phone != oldRec.Phone);
            }

            if (fieldChanged) {
                rec.AmpTimestamp__c = System.now();
            }
        }
    }
}
`
	if triggerCode != expectedTriggerCode {
		t.Errorf("trigger code mismatch.\nGot:\n%s\nWant:\n%s", triggerCode, expectedTriggerCode)
	}
}

func TestGenerateTriggerCodeForCDC(t *testing.T) {
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Lead",
		TriggerName: "Lead",
		WatchFields: []string{"Email", "Phone"},
	}

	got := GenerateTriggerCodeForCDC(params, "AmpTriggerSubscription__c")

	expected := `trigger Lead on Lead (before insert, before update) {
    if (Trigger.isBefore) {
        for (Lead rec : Trigger.new) {
            Boolean fieldChanged = false;

            if (Trigger.isInsert) {
                fieldChanged = (rec.Email != null) || (rec.Phone != null);
            } else if (Trigger.isUpdate) {
                Lead oldRec = Trigger.oldMap.get(rec.Id);
                fieldChanged = (rec.Email != oldRec.Email) || (rec.Phone != oldRec.Phone);
            }

            rec.AmpTriggerSubscription__c = fieldChanged;
        }
    }
}
`
	if got != expected {
		t.Errorf("trigger code mismatch.\nGot:\n%s\nWant:\n%s", got, expected)
	}
}

func TestGenerateTriggerCodeForCDCSingleField(t *testing.T) {
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Contact",
		TriggerName: "Contact",
		WatchFields: []string{"LastName"},
	}

	got := GenerateTriggerCodeForCDC(params, "AmpChanged__c")

	if !strings.Contains(got, "rec.AmpChanged__c = fieldChanged;") {
		t.Errorf("expected boolean assignment, got:\n%s", got)
	}

	if !strings.Contains(got, "(rec.LastName != null)") {
		t.Errorf("expected insert condition for LastName, got:\n%s", got)
	}

	if !strings.Contains(got, "(rec.LastName != oldRec.LastName)") {
		t.Errorf("expected update condition for LastName, got:\n%s", got)
	}
}

func TestGenerateTriggerCodeForFilteredRead(t *testing.T) {
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Lead",
		TriggerName: "Lead",
		WatchFields: []string{"Email", "Phone"},
	}

	got := GenerateTriggerCodeForFilteredRead(params, "AmpTimestamp__c")

	expected := `trigger Lead on Lead (before insert, before update) {
    if (Trigger.isBefore) {
        for (Lead rec : Trigger.new) {
            Boolean fieldChanged = false;

            if (Trigger.isInsert) {
                fieldChanged = (rec.Email != null) || (rec.Phone != null);
            } else if (Trigger.isUpdate) {
                Lead oldRec = Trigger.oldMap.get(rec.Id);
                fieldChanged = (rec.Email != oldRec.Email) || (rec.Phone != oldRec.Phone);
            }

            if (fieldChanged) {
                rec.AmpTimestamp__c = System.now();
            }
        }
    }
}
`
	if got != expected {
		t.Errorf("trigger code mismatch.\nGot:\n%s\nWant:\n%s", got, expected)
	}
}

func TestGenerateTriggerCodeForFilteredReadSingleField(t *testing.T) {
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Account",
		TriggerName: "Account",
		WatchFields: []string{"Name"},
	}

	got := GenerateTriggerCodeForFilteredRead(params, "AmpLastModified__c")

	if !strings.Contains(got, "rec.AmpLastModified__c = System.now();") {
		t.Errorf("expected timestamp assignment, got:\n%s", got)
	}

	if !strings.Contains(got, "if (fieldChanged)") {
		t.Errorf("expected conditional guard, got:\n%s", got)
	}

	if !strings.Contains(got, "(rec.Name != null)") {
		t.Errorf("expected insert condition for Name, got:\n%s", got)
	}
}

func TestGenerateApexTestClassName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		trigger   string
		expected  string
		expectErr bool
	}{
		{name: "CDC trigger", trigger: "CDC_Lead", expected: "Test_CDC_Lead"},
		{name: "Read trigger", trigger: "Read_Account", expected: "Test_Read_Account"},
		{name: "Empty returns error", trigger: "", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := GenerateApexTestClassName(tt.trigger)
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
				t.Errorf("GenerateApexTestClassName(%q) = %q, want %q", tt.trigger, got, tt.expected)
			}
		})
	}
}

func TestConstructApexTriggerBundlesTestClass(t *testing.T) {
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Lead",
		TriggerName: "CDC_Lead",
		WatchFields: []string{"Email", "Phone"},
	}

	triggerCode := GenerateTriggerCodeForCDC(params, "AmpTriggerSubscription__c")

	zipData, err := ConstructApexTrigger(params, triggerCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := readZipFiles(t, zipData)

	classCode, ok := files["classes/Test_CDC_Lead.cls"]
	if !ok {
		t.Fatal("expected classes/Test_CDC_Lead.cls in zip")
	}

	for _, want := range []string{
		"@isTest(SeeAllData=true)",
		"private class Test_CDC_Lead",
		"Schema.getGlobalDescribe().get('Lead')",
		"'SELECT Id FROM Lead LIMIT 1'",
		"'Email'",
		"'Phone'",
		"Database.insert(rec, false)",
		"Database.update(target, false)",
	} {
		if !strings.Contains(classCode, want) {
			t.Errorf("test class missing %q\nGot:\n%s", want, classCode)
		}
	}

	classMeta, ok := files["classes/Test_CDC_Lead.cls-meta.xml"]
	if !ok {
		t.Fatal("expected classes/Test_CDC_Lead.cls-meta.xml in zip")
	}

	expectedClassMeta := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ApexClass xmlns="http://soap.sforce.com/2006/04/metadata">
    <apiVersion>%s</apiVersion>
    <status>Active</status>
</ApexClass>
`, core.APIVersion)
	if classMeta != expectedClassMeta {
		t.Errorf("class meta mismatch.\nGot:\n%s\nWant:\n%s", classMeta, expectedClassMeta)
	}
}

func TestConstructDestructiveApexTrigger(t *testing.T) {
	t.Parallel()

	triggerName := "CDC_Lead"
	expectedTestClassName := "Test_CDC_Lead"

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

	// The destructiveChanges.xml must reference the trigger and the companion test class.
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

	if !strings.Contains(destructiveXML, expectedTestClassName) {
		t.Error("destructiveChanges.xml missing companion test class name")
	}

	if !strings.Contains(destructiveXML, "ApexClass") {
		t.Error("destructiveChanges.xml missing ApexClass type")
	}

	// The package.xml should be empty (no types with members).
	packageXML, ok := files["package.xml"]
	if !ok {
		t.Fatal("package.xml not found in zip")
	}

	if strings.Contains(packageXML, triggerName) {
		t.Error("package.xml should not contain the trigger name for destructive changes")
	}

	if strings.Contains(packageXML, expectedTestClassName) {
		t.Error("package.xml should not contain the test class name for destructive changes")
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
