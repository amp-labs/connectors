package metadata

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

// This file builds Metadata API deploy packages for Workflow Outbound Messages
// — the delivery half of the flow-based subscription path. An outbound message
// POSTs a SOAP notification to a configured HTTPS endpoint whenever the
// automation that references it (a record-triggered flow) fires.
//
// Outbound messages are a create/update mechanism only — Salesforce offers no
// way to fire one on record deletion.

var errOutboundMessageRequiredParams = errors.New(
	"objectName, name, endpointURL, and integrationUsername are required")

const metadataXmlns = "http://soap.sforce.com/2006/04/metadata"

// OutboundMessageParams contains the parameters for constructing the deploy
// package of one workflow outbound message.
//
// Field semantics mirror the WorkflowOutboundMessage metadata type; see the
// official docs (with declaration and package examples):
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_workflow.htm
type OutboundMessageParams struct {
	// ObjectName is the Salesforce object the outbound message belongs to
	// (e.g., "Lead").
	ObjectName string

	// Name is the outbound message developer name WITHOUT the object prefix
	// (e.g., "amp_Lead"). The Metadata API addresses the component as
	// "<ObjectName>.<Name>". Use GenerateOutboundMessageNameForSubscription()
	// to generate this.
	Name string

	// EndpointURL is where Salesforce POSTs the SOAP outbound message.
	// Must be an HTTPS endpoint reachable from Salesforce.
	EndpointURL string

	// IntegrationUsername is the Salesforce username whose permissions the
	// outbound message payload is evaluated under. Required by the
	// WorkflowOutboundMessage metadata type even when IncludeSessionID is
	// false, and must resolve to an active user in the org.
	IntegrationUsername string

	// Fields lists the object fields included in the outbound message payload.
	// Id, CreatedDate, and LastModifiedDate are always included on top of
	// whatever the caller lists (see ensureOutboundMessageFields): Id is how
	// consumers fetch the full record, and the audit timestamps let the event
	// parser infer create vs update — outbound messages carry no event type
	// of their own.
	Fields []string
}

// GenerateOutboundMessageNameForSubscription returns the outbound message
// developer name (without object prefix) for a given Salesforce object.
// Developer names must not contain consecutive underscores, so custom-object
// suffixes like "__c" are collapsed (e.g. "My_Object__c" → "amp_My_Object_c").
func GenerateOutboundMessageNameForSubscription(objectName string) (string, error) {
	if objectName == "" {
		return "", errEmptyObjectName
	}

	return "amp_" + sanitizeDeveloperName(objectName), nil
}

// OutboundMessageFullName returns the Metadata API full name of the outbound
// message component, which is prefixed by the object it belongs to.
func OutboundMessageFullName(objectName, outboundMessageName string) string {
	return objectName + "." + outboundMessageName
}

// sanitizeDeveloperName converts an object API name into a string that is
// valid inside a developer name: consecutive underscores collapse to one and
// trailing underscores are trimmed.
func sanitizeDeveloperName(name string) string {
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}

	return strings.Trim(name, "_")
}

// ValidateOutboundMessageParams checks that all required fields are present.
func ValidateOutboundMessageParams(params OutboundMessageParams) error {
	if params.ObjectName == "" || params.Name == "" ||
		params.EndpointURL == "" || params.IntegrationUsername == "" {
		return errOutboundMessageRequiredParams
	}

	return nil
}

// The XML structs below mirror the Metadata API WSDL. Element order matters:
// Salesforce validates deploy XML against an XSD sequence, so struct fields
// are declared in the order a metadata retrieve produces them.
//
// Official field reference and example XML for Workflow / WorkflowOutboundMessage:
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_workflow.htm

type workflowOutboundMessageXML struct {
	FullName           string   `xml:"fullName"`
	APIVersion         string   `xml:"apiVersion"`
	Description        string   `xml:"description"`
	EndpointURL        string   `xml:"endpointUrl"`
	Fields             []string `xml:"fields"`
	IncludeSessionID   bool     `xml:"includeSessionId"`
	IntegrationUser    string   `xml:"integrationUser"`
	Name               string   `xml:"name"`
	Protected          bool     `xml:"protected"`
	UseDeadLetterQueue bool     `xml:"useDeadLetterQueue"`
}

type workflowXML struct {
	XMLName          xml.Name                     `xml:"Workflow"`
	Xmlns            string                       `xml:"xmlns,attr"`
	OutboundMessages []workflowOutboundMessageXML `xml:"outboundMessages"`
}

// ensureOutboundMessageFields returns the caller's field list with Id,
// CreatedDate, and LastModifiedDate guaranteed present (prepended when
// missing, original order otherwise preserved). Every standard and custom
// object carries these audit fields; forcing them keeps the payload usable
// downstream — Id for record fetch, the timestamps for create-vs-update
// inference — regardless of what the caller configured.
func ensureOutboundMessageFields(fields []string) []string {
	required := []string{"Id", "CreatedDate", "LastModifiedDate"}

	present := make(map[string]bool, len(fields))
	for _, field := range fields {
		present[field] = true
	}

	missing := make([]string, 0, len(required))

	for _, field := range required {
		if !present[field] {
			missing = append(missing, field)
		}
	}

	return append(missing, fields...)
}

// generateWorkflowXML returns the workflows/<Object>.workflow file content
// carrying the single outbound message component.
func generateWorkflowXML(params OutboundMessageParams) (string, error) {
	fields := ensureOutboundMessageFields(params.Fields)

	workflow := workflowXML{
		Xmlns: metadataXmlns,
		OutboundMessages: []workflowOutboundMessageXML{
			{
				FullName:   params.Name,
				APIVersion: core.APIVersion,
				Description: "THIS IS AN AUTOMATED OUTBOUND MESSAGE. DO NOT EDIT. " +
					"It delivers subscription events for the object to Ampersand.",
				EndpointURL:      params.EndpointURL,
				Fields:           fields,
				IncludeSessionID: false,
				IntegrationUser:  params.IntegrationUsername,
				Name:             params.Name,
				Protected:        false,
				// Salesforce retries failed deliveries for up to 24 hours on
				// its own; the dead letter queue is unnecessary for our
				// at-least-once pipeline.
				UseDeadLetterQueue: false,
			},
		},
	}

	out, err := xml.MarshalIndent(workflow, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal workflow XML: %w", err)
	}

	return xml.Header + string(out) + "\n", nil
}

// ConstructOutboundMessage builds a zip deployment package containing ONLY the
// workflow outbound message for one object. The returned zip bytes are ready
// for DeployMetadataZip. Deployed as its own step (before any automation that
// references it) so the caller can record the created dependency in its result
// before moving on to the next artifact.
func ConstructOutboundMessage(params OutboundMessageParams) ([]byte, error) {
	if err := ValidateOutboundMessageParams(params); err != nil {
		return nil, err
	}

	workflowContent, err := generateWorkflowXML(params)
	if err != nil {
		return nil, err
	}

	pkg := triggerPackageXML{
		Xmlns:   metadataXmlns,
		Version: core.APIVersion,
		Types: []triggerPackageType{
			{
				Members: []string{OutboundMessageFullName(params.ObjectName, params.Name)},
				Name:    "WorkflowOutboundMessage",
			},
		},
	}

	pkgXML, err := xml.MarshalIndent(pkg, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package.xml: %w", err)
	}

	return buildZip(map[string][]byte{
		"package.xml": []byte(xml.Header + string(pkgXML)),
		"workflows/" + params.ObjectName + ".workflow": []byte(workflowContent),
	})
}

// ConstructDestructiveOutboundMessage builds a zipped destructive changes
// package that removes the outbound message. Deploy this AFTER every
// automation that references it is gone (dependents removed before dependency).
func ConstructDestructiveOutboundMessage(objectName, outboundMessageName string) ([]byte, error) {
	if objectName == "" || outboundMessageName == "" {
		return nil, errEmptyObjectName
	}

	return constructDestructiveZip(triggerPackageType{
		Members: []string{OutboundMessageFullName(objectName, outboundMessageName)},
		Name:    "WorkflowOutboundMessage",
	})
}

// constructDestructiveZip builds a destructive-changes zip for the given
// package types with the empty package.xml required by the Metadata API.
func constructDestructiveZip(types ...triggerPackageType) ([]byte, error) {
	emptyPkg := triggerPackageXML{
		Xmlns:   metadataXmlns,
		Version: core.APIVersion,
		Types:   []triggerPackageType{},
	}

	emptyPkgXML, err := xml.MarshalIndent(emptyPkg, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal empty package.xml: %w", err)
	}

	destructivePkg := triggerPackageXML{
		Xmlns:   metadataXmlns,
		Version: core.APIVersion,
		Types:   types,
	}

	destructiveXML, err := xml.MarshalIndent(destructivePkg, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal destructiveChanges.xml: %w", err)
	}

	return buildZip(map[string][]byte{
		"package.xml":            []byte(xml.Header + string(emptyPkgXML)),
		"destructiveChanges.xml": []byte(xml.Header + string(destructiveXML)),
	})
}

// buildZip writes the given name→content entries into an in-memory zip.
func buildZip(entries map[string][]byte) ([]byte, error) {
	var buf bytes.Buffer

	zipWriter := zip.NewWriter(&buf)

	for name, content := range entries {
		if err := addTriggerToZip(zipWriter, name, content); err != nil {
			return nil, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}
