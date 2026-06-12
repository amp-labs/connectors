package metadata

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
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

func TestGenerateApexHandlerClassName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		trigger   string
		expected  string
		expectErr bool
	}{
		{name: "CDC trigger", trigger: "CDC_Lead", expected: "CDC_Lead_Handler"},
		{name: "Read trigger", trigger: "Read_Account", expected: "Read_Account_Handler"},
		{name: "Empty returns error", trigger: "", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := GenerateApexHandlerClassName(tt.trigger)
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
				t.Errorf("GenerateApexHandlerClassName(%q) = %q, want %q", tt.trigger, got, tt.expected)
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

				if !errors.Is(err, tt.expectErr) {
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

func TestConstructApexTriggerUnsupportedIndicatorType(t *testing.T) {
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Lead",
		TriggerName: "CDC_Lead",
		IndicatorField: common.FieldDefinition{
			FieldName: "AmpString__c",
			ValueType: common.FieldTypeString,
		},
		WatchFields: []string{"Email"},
	}

	if _, err := ConstructApexTrigger(t.Context(), params); !errors.Is(err, errUnsupportedIndicatorTy) {
		t.Errorf("expected errUnsupportedIndicatorTy, got %v", err)
	}
}

func TestConstructApexTriggerZipFileList(t *testing.T) {
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Lead",
		TriggerName: "CDC_Lead",
		IndicatorField: common.FieldDefinition{
			FieldName: "AmpTriggerSubscription__c",
			ValueType: common.FieldTypeBoolean,
		},
		WatchFields: []string{"Email"},
	}

	zipData, err := ConstructApexTrigger(t.Context(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertZipContainsFiles(t, zipData, []string{
		"package.xml",
		"triggers/CDC_Lead.trigger",
		"triggers/CDC_Lead.trigger-meta.xml",
		"classes/CDC_Lead_Handler.cls",
		"classes/CDC_Lead_Handler.cls-meta.xml",
		"classes/Test_CDC_Lead.cls",
		"classes/Test_CDC_Lead.cls-meta.xml",
	})
}

func TestConstructApexTriggerForCDCContent(t *testing.T) { //nolint:funlen
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Lead",
		TriggerName: "CDC_Lead",
		IndicatorField: common.FieldDefinition{
			FieldName: "AmpTriggerSubscription__c",
			ValueType: common.FieldTypeBoolean,
		},
		WatchFields: []string{"Email", "Phone"},
	}

	zipData, err := ConstructApexTrigger(t.Context(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := readZipFiles(t, zipData)

	// 1. The trigger is thin — it delegates to the handler. before-insert is
	//    kept so the companion test's Database.insert covers the trigger's
	//    delegation line; the handler's insert path is a no-op.
	expectedTrigger := `trigger CDC_Lead on Lead (before insert, before update) {
    CDC_Lead_Handler.process(Trigger.new, Trigger.old);
}
`
	if got := files["triggers/CDC_Lead.trigger"]; got != expectedTrigger {
		t.Errorf("trigger code mismatch.\nGot:\n%s\nWant:\n%s", got, expectedTrigger)
	}

	// 2. The handler holds the change-detection logic for the update path; the
	//    insert path is an early return because CREATE CDC events bypass the
	//    channel-member filter expression unconditionally.
	expectedHandler := `public class CDC_Lead_Handler {
    public static void process(List<Lead> newRecs, List<Lead> oldRecs) {
        if (oldRecs == null) {
            // Insert: no-op for CDC purposes. CREATE events bypass the channel
            // member's filter expression unconditionally, so the indicator's
            // value at insert time has no effect on event delivery.
            return;
        }
        for (Integer i = 0; i < newRecs.size(); i++) {
            Lead rec = newRecs[i];
            Lead oldRec = oldRecs[i];
            Boolean fieldChanged = (rec.Email != oldRec.Email) || (rec.Phone != oldRec.Phone);

            rec.AmpTriggerSubscription__c = fieldChanged;
        }
    }
}
`
	if got := files["classes/CDC_Lead_Handler.cls"]; got != expectedHandler {
		t.Errorf("handler code mismatch.\nGot:\n%s\nWant:\n%s", got, expectedHandler)
	}

	// 3. Trigger meta XML.
	expectedMetaXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ApexTrigger xmlns="http://soap.sforce.com/2006/04/metadata">
    <apiVersion>%s</apiVersion>
    <status>Active</status>
</ApexTrigger>
`, core.APIVersion)
	if got := files["triggers/CDC_Lead.trigger-meta.xml"]; got != expectedMetaXML {
		t.Errorf("trigger meta XML mismatch.\nGot:\n%s\nWant:\n%s", got, expectedMetaXML)
	}

	// 4. Package XML lists the trigger (1 ApexTrigger member) plus both Apex
	//    classes (handler + test) under a single ApexClass type.
	expectedPackageXML := xml.Header + fmt.Sprintf(`<Package xmlns="http://soap.sforce.com/2006/04/metadata">
    <types>
        <members>CDC_Lead</members>
        <name>ApexTrigger</name>
    </types>
    <types>
        <members>CDC_Lead_Handler</members>
        <members>Test_CDC_Lead</members>
        <name>ApexClass</name>
    </types>
    <version>%s</version>
</Package>`, core.APIVersion)
	if got := files["package.xml"]; got != expectedPackageXML {
		t.Errorf("package.xml mismatch.\nGot:\n%s\nWant:\n%s", got, expectedPackageXML)
	}
}

func TestConstructApexTriggerForFilteredReadContent(t *testing.T) {
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Lead",
		TriggerName: "Read_Lead",
		IndicatorField: common.FieldDefinition{
			FieldName: "AmpTimestamp__c",
			ValueType: common.FieldTypeDateTime,
		},
		WatchFields: []string{"Email", "Phone"},
	}

	zipData, err := ConstructApexTrigger(t.Context(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := readZipFiles(t, zipData)

	// Trigger is the same thin delegation regardless of CDC vs filtered-read variant.
	expectedTrigger := `trigger Read_Lead on Lead (before insert, before update) {
    Read_Lead_Handler.process(Trigger.new, Trigger.old);
}
`
	if got := files["triggers/Read_Lead.trigger"]; got != expectedTrigger {
		t.Errorf("trigger code mismatch.\nGot:\n%s\nWant:\n%s", got, expectedTrigger)
	}

	// The handler differs from CDC: the indicator assignment is conditional on
	// fieldChanged being true and writes System.now() instead of fieldChanged.
	expectedHandler := `public class Read_Lead_Handler {
    public static void process(List<Lead> newRecs, List<Lead> oldRecs) {
        if (oldRecs == null) {
            // Insert: no-op for CDC purposes. CREATE events bypass the channel
            // member's filter expression unconditionally, so the indicator's
            // value at insert time has no effect on event delivery.
            return;
        }
        for (Integer i = 0; i < newRecs.size(); i++) {
            Lead rec = newRecs[i];
            Lead oldRec = oldRecs[i];
            Boolean fieldChanged = (rec.Email != oldRec.Email) || (rec.Phone != oldRec.Phone);

            if (fieldChanged) {
                rec.AmpTimestamp__c = System.now();
            }
        }
    }
}
`
	if got := files["classes/Read_Lead_Handler.cls"]; got != expectedHandler {
		t.Errorf("handler code mismatch.\nGot:\n%s\nWant:\n%s", got, expectedHandler)
	}
}

func TestConstructApexTriggerHandlerSingleField(t *testing.T) {
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Contact",
		TriggerName: "CDC_Contact",
		IndicatorField: common.FieldDefinition{
			FieldName: "AmpChanged__c",
			ValueType: common.FieldTypeBoolean,
		},
		WatchFields: []string{"LastName"},
	}

	zipData, err := ConstructApexTrigger(t.Context(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := readZipFiles(t, zipData)
	handler := files["classes/CDC_Contact_Handler.cls"]

	for _, want := range []string{
		"public class CDC_Contact_Handler",
		"List<Contact> newRecs",
		"List<Contact> oldRecs",
		"if (oldRecs == null)",
		"(rec.LastName != oldRec.LastName)",
		"rec.AmpChanged__c = fieldChanged;",
	} {
		if !strings.Contains(handler, want) {
			t.Errorf("handler missing %q\nGot:\n%s", want, handler)
		}
	}
}

func TestConstructApexTriggerBundlesTestClass(t *testing.T) { //nolint:funlen
	t.Parallel()

	params := ApexTriggerParams{
		ObjectName:  "Lead",
		TriggerName: "CDC_Lead",
		IndicatorField: common.FieldDefinition{
			FieldName: "AmpTriggerSubscription__c",
			ValueType: common.FieldTypeBoolean,
		},
		WatchFields: []string{"Email", "Phone"},
	}

	zipData, err := ConstructApexTrigger(t.Context(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := readZipFiles(t, zipData)

	classCode, ok := files["classes/Test_CDC_Lead.cls"]
	if !ok {
		t.Fatal("expected classes/Test_CDC_Lead.cls in zip")
	}

	// The test class must:
	//   - Be marked @isTest so it's excluded from coverage requirements itself.
	//   - Invoke the handler directly (in-memory) for both the insert no-op
	//     and the update branch so handler coverage doesn't depend on whether
	//     DML against the real object succeeds.
	//   - Populate the first watch field on the in-memory newRec for the
	//     update branch so the change-detection condition evaluates true and
	//     the Read variant's conditional indicator-assignment line is covered.
	//   - In coverTriggerDelegation, use @isTest(SeeAllData=true) and a SOQL
	//     LIMIT 1 to find an existing record, then no-op-update it so the
	//     before-update trigger fires (covers the trigger's one line without
	//     needing makeRec() to construct a valid record). Fall back to a
	//     makeRec()-based insert for orgs with zero records of the type.
	for _, want := range []string{
		"@isTest",
		"private class Test_CDC_Lead",
		"static void exerciseHandlerInsertNoop",
		"static void exerciseHandlerUpdateBranch",
		"@isTest(SeeAllData=true)",
		"static void coverTriggerDelegation",
		"List<Lead> existing = [SELECT Id FROM Lead LIMIT 1]",
		"Database.update(existing[0], false)",
		"Database.insert(makeRec(), false)",
		"setWatchFieldValueIfPossible(newRec, 'Email')",
		"private static void setWatchFieldValueIfPossible(SObject rec, String fieldName)",
		"CDC_Lead_Handler.process",
		"Schema.getGlobalDescribe().get('Lead')",
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

	// The handler class meta XML should use the same template as the test class.
	handlerMeta, ok := files["classes/CDC_Lead_Handler.cls-meta.xml"]
	if !ok {
		t.Fatal("expected classes/CDC_Lead_Handler.cls-meta.xml in zip")
	}

	if handlerMeta != expectedClassMeta {
		t.Errorf("handler meta mismatch.\nGot:\n%s\nWant:\n%s", handlerMeta, expectedClassMeta)
	}
}

func TestConstructDestructiveApexTrigger(t *testing.T) { //nolint:funlen
	t.Parallel()

	triggerName := "CDC_Lead"
	expectedHandlerClassName := "CDC_Lead_Handler"
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

	// destructiveChanges.xml must reference the trigger, the handler, and
	// the companion test class.
	destructiveXML, ok := files["destructiveChanges.xml"]
	if !ok {
		t.Fatal("destructiveChanges.xml not found in zip")
	}

	for _, want := range []string{
		triggerName,
		expectedHandlerClassName,
		expectedTestClassName,
		"ApexTrigger",
		"ApexClass",
	} {
		if !strings.Contains(destructiveXML, want) {
			t.Errorf("destructiveChanges.xml missing %q\nGot:\n%s", want, destructiveXML)
		}
	}

	// package.xml must be empty (no types with members).
	packageXML, ok := files["package.xml"]
	if !ok {
		t.Fatal("package.xml not found in zip")
	}

	for _, unwanted := range []string{
		triggerName,
		expectedHandlerClassName,
		expectedTestClassName,
	} {
		if strings.Contains(packageXML, unwanted) {
			t.Errorf("package.xml should not contain %q for destructive changes", unwanted)
		}
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
