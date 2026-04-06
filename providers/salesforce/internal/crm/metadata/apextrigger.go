package metadata

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

var (
	errWatchFieldsEmpty         = errors.New("watchFields must not be empty")
	errRequiredParamsMet        = errors.New("objectName, triggerName, and indicatorField are required")
	errUnsupportedIndicatorType = errors.New("unsupported indicator field type: only boolean and datetime are supported")
)

// ApexTriggerParams contains the parameters for constructing and deploying an APEX trigger.
type ApexTriggerParams struct {
	// ObjectName is the Salesforce object the trigger runs on (e.g., "Lead").
	ObjectName string

	// TriggerName is the name of the APEX trigger (e.g., "AmpersandTrack_Lead").
	// Use GenerateApexTriggerName() to generate this.
	TriggerName string

	// IndicatorField is the field definition for the indicator field that the trigger sets
	// when watched fields change. Supported types: boolean (sets to true/false) and
	// datetime (sets to System.now()).
	IndicatorField common.FieldDefinition

	// WatchFields is the list of field API names to monitor for changes.
	WatchFields []string
}

// GenerateApexTriggerName returns the standard APEX trigger name for a given Salesforce object.
func GenerateApexTriggerName(objectName string) string {
	return objectName
}

// ConstructApexTrigger builds a zipped deployment package for an APEX trigger that sets
// an indicator field when any of the specified watch fields change.
//
// The trigger handles both insert and update events:
//   - On insert: sets indicator if any watch field has a non-null value.
//   - On update: sets indicator if any watch field's value differs from the old record.
//
// Supported indicator field types:
//   - boolean: sets field to true/false based on whether fields changed.
//   - datetime: sets field to System.now() when fields changed.
//
// The returned zip bytes are ready for DeployMetadataZip.
func ConstructApexTrigger(params ApexTriggerParams) ([]byte, error) {
	if len(params.WatchFields) == 0 {
		return nil, errWatchFieldsEmpty
	}

	if params.ObjectName == "" || params.TriggerName == "" || params.IndicatorField.FieldName == "" {
		return nil, errRequiredParamsMet
	}

	triggerCode, err := generateTriggerCode(params)
	if err != nil {
		return nil, err
	}

	triggerMetaXML := generateTriggerMetaXML()

	return createTriggerDeployZip(params.TriggerName, triggerCode, triggerMetaXML)
}

// ConstructDestructiveApexTrigger builds a zipped destructive changes package to delete
// an APEX trigger from Salesforce. The returned zip bytes are ready for DeployMetadataZip.
func ConstructDestructiveApexTrigger(triggerName string) ([]byte, error) {
	return createTriggerDestructiveZip(triggerName)
}

// generateTriggerCode dynamically generates APEX trigger code.
// The indicator assignment varies based on the IndicatorField type:
//   - boolean: rec.field = fieldChanged
//   - datetime: rec.field = System.now() (only when fieldChanged is true)
func generateTriggerCode(params ApexTriggerParams) (string, error) {
	indicatorAssignment, err := buildIndicatorAssignment(params.IndicatorField)
	if err != nil {
		return "", err
	}

	// Build insert condition: field != null
	// We only check != null (not != '') because the empty-string check is invalid
	// for non-String Apex types (Boolean, Datetime, Number, etc.) and would cause
	// compilation errors. The null check is sufficient and type-safe for all field types.
	insertConditions := make([]string, 0, len(params.WatchFields))
	for _, field := range params.WatchFields {
		insertConditions = append(insertConditions,
			fmt.Sprintf("(rec.%s != null)", field))
	}

	insertExpr := strings.Join(insertConditions, " || ")

	// Build update condition: field changed compared to old record
	updateConditions := make([]string, 0, len(params.WatchFields))
	for _, field := range params.WatchFields {
		updateConditions = append(updateConditions,
			fmt.Sprintf("(rec.%s != oldRec.%s)", field, field))
	}

	updateExpr := strings.Join(updateConditions, " || ")

	return fmt.Sprintf(`trigger %s on %s (before insert, before update) {
    if (Trigger.isBefore) {
        for (%s rec : Trigger.new) {
            Boolean fieldChanged = false;

            if (Trigger.isInsert) {
                fieldChanged = %s;
            } else if (Trigger.isUpdate) {
                %s oldRec = Trigger.oldMap.get(rec.Id);
                fieldChanged = %s;
            }

            %s
        }
    }
}
`, params.TriggerName, params.ObjectName, params.ObjectName,
		insertExpr, params.ObjectName, updateExpr, indicatorAssignment), nil
}

// buildIndicatorAssignment returns the Apex code snippet that sets the indicator field
// based on whether watched fields changed.
func buildIndicatorAssignment(field common.FieldDefinition) (string, error) {
	switch field.ValueType { //nolint:exhaustive
	case common.FieldTypeBoolean:
		return fmt.Sprintf("rec.%s = fieldChanged;", field.FieldName), nil
	case common.FieldTypeDateTime:
		return fmt.Sprintf(`if (fieldChanged) {
                rec.%s = System.now();
            }`, field.FieldName), nil
	default:
		return "", fmt.Errorf("%w: %s", errUnsupportedIndicatorType, field.ValueType)
	}
}

func generateTriggerMetaXML() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ApexTrigger xmlns="http://soap.sforce.com/2006/04/metadata">
    <apiVersion>%s</apiVersion>
    <status>Active</status>
</ApexTrigger>
`, core.APIVersion)
}

// triggerPackageXML is the structure for Salesforce package.xml manifests.
type triggerPackageXML struct {
	XMLName xml.Name             `xml:"Package"`
	Xmlns   string               `xml:"xmlns,attr"`
	Types   []triggerPackageType `xml:"types"`
	Version string               `xml:"version"`
}

type triggerPackageType struct {
	Members []string `xml:"members"`
	Name    string   `xml:"name"`
}

func createTriggerDeployZip(triggerName, triggerCode, triggerMetaXML string) ([]byte, error) {
	pkg := triggerPackageXML{
		Xmlns:   "http://soap.sforce.com/2006/04/metadata",
		Version: core.APIVersion,
		Types: []triggerPackageType{
			{
				Members: []string{triggerName},
				Name:    "ApexTrigger",
			},
		},
	}

	pkgXML, err := xml.MarshalIndent(pkg, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package.xml: %w", err)
	}

	var buf bytes.Buffer

	zipWriter := zip.NewWriter(&buf)

	if err := addTriggerToZip(zipWriter, "package.xml", []byte(xml.Header+string(pkgXML))); err != nil {
		return nil, err
	}

	if err := addTriggerToZip(zipWriter, "triggers/"+triggerName+".trigger", []byte(triggerCode)); err != nil {
		return nil, err
	}

	if err := addTriggerToZip(zipWriter, "triggers/"+triggerName+".trigger-meta.xml", []byte(triggerMetaXML)); err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

func createTriggerDestructiveZip(triggerName string) ([]byte, error) {
	emptyPkg := triggerPackageXML{
		Xmlns:   "http://soap.sforce.com/2006/04/metadata",
		Version: core.APIVersion,
		Types:   []triggerPackageType{},
	}

	emptyPkgXML, err := xml.MarshalIndent(emptyPkg, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal empty package.xml: %w", err)
	}

	destructivePkg := triggerPackageXML{
		Xmlns:   "http://soap.sforce.com/2006/04/metadata",
		Version: core.APIVersion,
		Types: []triggerPackageType{
			{
				Members: []string{triggerName},
				Name:    "ApexTrigger",
			},
		},
	}

	destructiveXML, err := xml.MarshalIndent(destructivePkg, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal destructiveChanges.xml: %w", err)
	}

	var buf bytes.Buffer

	zipWriter := zip.NewWriter(&buf)

	if err := addTriggerToZip(zipWriter, "package.xml", []byte(xml.Header+string(emptyPkgXML))); err != nil {
		return nil, err
	}

	if err := addTriggerToZip(zipWriter, "destructiveChanges.xml", []byte(xml.Header+string(destructiveXML))); err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

func addTriggerToZip(zw *zip.Writer, name string, content []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create zip entry %s: %w", name, err)
	}

	if _, err := w.Write(content); err != nil {
		return fmt.Errorf("failed to write zip entry %s: %w", name, err)
	}

	return nil
}
