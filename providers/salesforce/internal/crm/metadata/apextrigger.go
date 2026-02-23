package metadata

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
)

var (
	errWatchFieldsEmpty  = errors.New("watchFields must not be empty")
	errRequiredParamsMet = errors.New("objectName, triggerName, and checkboxFieldName are required")
)

const (
	// triggerPrefix is the prefix used for generated APEX trigger names.
	triggerPrefix = "AmpersandTrack_"

	// apexTriggerMetadataAPIVersion is the Salesforce API version for trigger metadata.
	apexTriggerMetadataAPIVersion = "61.0"
)

// ApexTriggerParams contains the parameters for constructing and deploying an APEX trigger.
type ApexTriggerParams struct {
	// ObjectName is the Salesforce object the trigger runs on (e.g., "Lead").
	ObjectName string

	// TriggerName is the name of the APEX trigger (e.g., "AmpersandTrack_Lead").
	// Use GenerateApexTriggerName() to generate this.
	TriggerName string

	// CheckboxFieldName is the API name of the boolean field that the trigger sets
	// (e.g., "AmpTriggerSubscription__c").
	CheckboxFieldName string

	// WatchFields is the list of field API names to monitor for changes.
	WatchFields []string
}

// GenerateApexTriggerName returns the standard APEX trigger name for a given Salesforce object.
func GenerateApexTriggerName(objectName string) string {
	return triggerPrefix + objectName
}

// ConstructApexTrigger builds a zipped deployment package for an APEX trigger that sets
// a boolean checkbox field to true when any of the specified watch fields change.
//
// The trigger handles both insert and update events:
//   - On insert: sets checkbox to true if any watch field has a non-null/non-empty value.
//   - On update: sets checkbox to true if any watch field's value differs from the old record.
//
// The returned zip bytes are ready for DeployMetadataZip.
func ConstructApexTrigger(params ApexTriggerParams) ([]byte, error) {
	if len(params.WatchFields) == 0 {
		return nil, errWatchFieldsEmpty
	}

	if params.ObjectName == "" || params.TriggerName == "" || params.CheckboxFieldName == "" {
		return nil, errRequiredParamsMet
	}

	triggerCode := generateTriggerCode(params)
	triggerMetaXML := generateTriggerMetaXML()

	return createTriggerDeployZip(params.TriggerName, triggerCode, triggerMetaXML)
}

// ConstructDestructiveApexTrigger builds a zipped destructive changes package to delete
// an APEX trigger from Salesforce. The returned zip bytes are ready for DeployMetadataZip.
func ConstructDestructiveApexTrigger(triggerName string) ([]byte, error) {
	return createTriggerDestructiveZip(triggerName)
}

// generateTriggerCode dynamically generates APEX trigger code.
func generateTriggerCode(params ApexTriggerParams) string {
	// Build insert condition: field != null && field != ''
	insertConditions := make([]string, 0, len(params.WatchFields))
	for _, field := range params.WatchFields {
		insertConditions = append(insertConditions,
			fmt.Sprintf("(rec.%s != null && rec.%s != '')", field, field))
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

            rec.%s = fieldChanged;
        }
    }
}
`, params.TriggerName, params.ObjectName, params.ObjectName,
		insertExpr, params.ObjectName, updateExpr, params.CheckboxFieldName)
}

func generateTriggerMetaXML() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ApexTrigger xmlns="http://soap.sforce.com/2006/04/metadata">
    <apiVersion>%s</apiVersion>
    <status>Active</status>
</ApexTrigger>
`, apexTriggerMetadataAPIVersion)
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
		Version: apexTriggerMetadataAPIVersion,
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
		Version: apexTriggerMetadataAPIVersion,
		Types:   []triggerPackageType{},
	}

	emptyPkgXML, err := xml.MarshalIndent(emptyPkg, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal empty package.xml: %w", err)
	}

	destructivePkg := triggerPackageXML{
		Xmlns:   "http://soap.sforce.com/2006/04/metadata",
		Version: apexTriggerMetadataAPIVersion,
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
