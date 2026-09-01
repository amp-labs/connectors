package metadata

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

// This file builds Metadata API deploy packages for the record-triggered Flow
// of a flow-based subscription. The flow fires after a record is created
// and/or updated and invokes a Workflow Outbound Message action (see
// outboundmessage.go), which POSTs a SOAP notification to the configured
// endpoint.
//
// DELETE events are not supported on this path: record-triggered delete flows
// run before-delete and do not support the outbound message action.

var (
	errFlowRequiredParams = errors.New(
		"objectName, flowName, and outboundMessageName are required")
	errFlowInvalidTriggerType = errors.New(
		"recordTriggerType must be Create, Update, or CreateAndUpdate")
	errNoFlowComponents = errors.New(
		"at least one flow subscription component is required")
)

// FlowSubscriptionComponent pairs the two artifacts deployed for one object
// of a flow-based subscription: the workflow outbound message and the
// record-triggered flow that invokes it.
type FlowSubscriptionComponent struct {
	OutboundMessage OutboundMessageParams
	Flow            FlowParams
}

// validateFlowSubscriptionComponent checks one component's params, that the
// pair is internally consistent, and that its object was not already seen.
func validateFlowSubscriptionComponent(
	component FlowSubscriptionComponent, seenObjects map[string]bool,
) error {
	omParams, flowParams := component.OutboundMessage, component.Flow

	if err := ValidateOutboundMessageParams(omParams); err != nil {
		return err
	}

	if err := ValidateFlowParams(flowParams); err != nil {
		return err
	}

	if omParams.ObjectName != flowParams.ObjectName || omParams.Name != flowParams.OutboundMessageName {
		return fmt.Errorf("%w: outbound message %s.%s does not match flow reference %s.%s",
			errFlowRequiredParams,
			omParams.ObjectName, omParams.Name,
			flowParams.ObjectName, flowParams.OutboundMessageName)
	}

	if seenObjects[omParams.ObjectName] {
		return fmt.Errorf("%w: duplicate object %s", errFlowRequiredParams, omParams.ObjectName)
	}

	return nil
}

// RecordTriggerType maps to Flow.start.recordTriggerType in the Metadata API.
type RecordTriggerType string

const (
	RecordTriggerTypeCreate          RecordTriggerType = "Create"
	RecordTriggerTypeUpdate          RecordTriggerType = "Update"
	RecordTriggerTypeCreateAndUpdate RecordTriggerType = "CreateAndUpdate"
)

// FlowParams contains the parameters for constructing the deploy package of
// one record-triggered flow that invokes an outbound message.
//
// Field semantics mirror the Flow metadata type (FlowStart's recordTriggerType
// and filterFormula in particular); see the official docs (with declaration
// details and example XML):
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_visual_workflow.htm
type FlowParams struct {
	// ObjectName is the Salesforce object the flow runs on (e.g., "Lead").
	ObjectName string

	// FlowName is the Flow API name (e.g., "AmpSubscribe_Lead").
	// Use GenerateFlowNameForSubscription() to generate this.
	FlowName string

	// OutboundMessageName is the developer name (WITHOUT the object prefix)
	// of the outbound message the flow invokes. The referenced outbound
	// message must already exist in Salesforce when the flow deploys.
	OutboundMessageName string

	// RecordTriggerType selects when the flow fires (Create / Update /
	// CreateAndUpdate).
	RecordTriggerType RecordTriggerType

	// WatchFields optionally narrows Update triggers to changes of specific
	// fields via an ISCHANGED() entry-condition formula. Only applied when
	// RecordTriggerType is Update: ISCHANGED is not available in entry
	// conditions of flows that also fire on Create.
	WatchFields []string
}

// GenerateFlowNameForSubscription returns the Flow API name used for the
// flow-based subscription on a given Salesforce object. Flow API names must
// not contain consecutive underscores, so custom-object suffixes like "__c"
// are collapsed (e.g. "My_Object__c" → "AmpSubscribe_My_Object_c").
func GenerateFlowNameForSubscription(objectName string) (string, error) {
	if objectName == "" {
		return "", errEmptyObjectName
	}

	return "AmpSubscribe_" + sanitizeDeveloperName(objectName), nil
}

// ValidateFlowParams checks that all required fields are present and the
// trigger type is one Salesforce accepts for after-save flows.
func ValidateFlowParams(params FlowParams) error {
	if params.ObjectName == "" || params.FlowName == "" || params.OutboundMessageName == "" {
		return errFlowRequiredParams
	}

	switch params.RecordTriggerType {
	case RecordTriggerTypeCreate, RecordTriggerTypeUpdate, RecordTriggerTypeCreateAndUpdate:
		return nil
	default:
		return fmt.Errorf("%w: got %q", errFlowInvalidTriggerType, params.RecordTriggerType)
	}
}

// The XML structs below mirror the Metadata API WSDL. Element order matters:
// Salesforce validates deploy XML against an XSD sequence, so struct fields
// are declared in the order a metadata retrieve produces them.
//
// Official field reference and example XML for Flow (incl. FlowActionCall,
// FlowStart, and the outboundMessage action type):
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_visual_workflow.htm

type flowActionCallXML struct {
	Name                 string `xml:"name"`
	Label                string `xml:"label"`
	LocationX            int    `xml:"locationX"`
	LocationY            int    `xml:"locationY"`
	ActionName           string `xml:"actionName"`
	ActionType           string `xml:"actionType"`
	FlowTransactionModel string `xml:"flowTransactionModel"`
	NameSegment          string `xml:"nameSegment"`
}

type flowConnectorXML struct {
	TargetReference string `xml:"targetReference"`
}

type flowStartXML struct {
	LocationX         int               `xml:"locationX"`
	LocationY         int               `xml:"locationY"`
	Connector         *flowConnectorXML `xml:"connector,omitempty"`
	FilterFormula     string            `xml:"filterFormula,omitempty"`
	Object            string            `xml:"object"`
	RecordTriggerType string            `xml:"recordTriggerType"`
	TriggerType       string            `xml:"triggerType"`
}

type flowXML struct {
	XMLName        xml.Name          `xml:"Flow"`
	Xmlns          string            `xml:"xmlns,attr"`
	ActionCalls    flowActionCallXML `xml:"actionCalls"`
	APIVersion     string            `xml:"apiVersion"`
	Environments   string            `xml:"environments"`
	InterviewLabel string            `xml:"interviewLabel"`
	Label          string            `xml:"label"`
	ProcessType    string            `xml:"processType"`
	Start          flowStartXML      `xml:"start"`
	Status         string            `xml:"status"`
}

// flowSendActionName is the flow element name of the single outbound message
// action call inside a generated flow.
const flowSendActionName = "Send_Outbound_Message"

// GenerateFlowXML returns the .flow file content for the record-triggered
// flow of a flow-based subscription: a single after-save start element wired
// to one outbound message action call. Exported so the deactivation path can
// rebuild the same flow with a different status.
func GenerateFlowXML(params FlowParams, status string) (string, error) {
	omFullName := OutboundMessageFullName(params.ObjectName, params.OutboundMessageName)

	flow := flowXML{
		Xmlns: metadataXmlns,
		ActionCalls: flowActionCallXML{
			Name:       flowSendActionName,
			Label:      "Send Outbound Message",
			LocationX:  176, //nolint:mnd // canvas coordinates; arbitrary but required
			LocationY:  287, //nolint:mnd
			ActionName: omFullName,
			ActionType: "outboundMessage",
			// CurrentTransaction sends the message when the trigger's
			// transaction commits — a rolled-back save sends nothing.
			FlowTransactionModel: "CurrentTransaction",
			NameSegment:          omFullName,
		},
		APIVersion:     core.APIVersion,
		Environments:   "Default",
		InterviewLabel: params.FlowName + " {!$Flow.CurrentDateTime}",
		Label:          params.FlowName,
		ProcessType:    "AutoLaunchedFlow",
		Start: flowStartXML{
			LocationX:         50, //nolint:mnd
			LocationY:         0,
			Connector:         &flowConnectorXML{TargetReference: flowSendActionName},
			FilterFormula:     buildWatchFieldsFilterFormula(params),
			Object:            params.ObjectName,
			RecordTriggerType: string(params.RecordTriggerType),
			TriggerType:       "RecordAfterSave",
		},
		Status: status,
	}

	out, err := xml.MarshalIndent(flow, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal flow XML: %w", err)
	}

	return xml.Header + string(out) + "\n", nil
}

// buildWatchFieldsFilterFormula returns the flow entry-condition formula that
// restricts Update triggers to changes of the watched fields, or "" when the
// flow also fires on Create (ISCHANGED is rejected there) or no watch fields
// are configured.
func buildWatchFieldsFilterFormula(params FlowParams) string {
	if params.RecordTriggerType != RecordTriggerTypeUpdate || len(params.WatchFields) == 0 {
		return ""
	}

	changed := make([]string, 0, len(params.WatchFields))
	for _, field := range params.WatchFields {
		changed = append(changed, fmt.Sprintf("ISCHANGED({!$Record.%s})", field))
	}

	if len(changed) == 1 {
		return changed[0]
	}

	return "OR(" + strings.Join(changed, ", ") + ")"
}

// ConstructFlow builds a zip deployment package containing ONLY the
// record-triggered flow, deployed with the given status:
//
//   - "Active" for the subscribe path. On production orgs an Active deploy
//     requires either flow test coverage or the "Deploy processes and flows
//     as active" setting; an activation failure surfaces as a component
//     failure on the deploy. The outbound message the flow references must
//     already exist (deploy ConstructOutboundMessage first).
//   - "Draft" for the deactivation fallback used before deletion when the
//     Tooling API deactivation is unavailable.
func ConstructFlow(params FlowParams, status string) ([]byte, error) {
	if err := ValidateFlowParams(params); err != nil {
		return nil, err
	}

	flowContent, err := GenerateFlowXML(params, status)
	if err != nil {
		return nil, err
	}

	pkg := triggerPackageXML{
		Xmlns:   metadataXmlns,
		Version: core.APIVersion,
		Types: []triggerPackageType{
			{
				Members: []string{params.FlowName},
				Name:    "Flow",
			},
		},
	}

	pkgXML, err := xml.MarshalIndent(pkg, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package.xml: %w", err)
	}

	return buildZip(map[string][]byte{
		"package.xml":                        []byte(xml.Header + string(pkgXML)),
		"flows/" + params.FlowName + ".flow": []byte(flowContent),
	})
}

// ConstructDestructiveFlow builds a zipped destructive changes package that
// removes the flow. The flow must be inactive before this deploy — Salesforce
// rejects deletion of active flows.
func ConstructDestructiveFlow(flowName string) ([]byte, error) {
	if flowName == "" {
		return nil, errEmptyObjectName
	}

	return constructDestructiveZip(triggerPackageType{
		Members: []string{flowName},
		Name:    "Flow",
	})
}

// ConstructFlowSubscription builds ONE deploy package carrying both artifacts
// of a flow-based subscription for a single object: the workflow outbound
// message and the Active record-triggered flow that invokes it. Salesforce
// metadata deploys are atomic per package (our deploy options set
// rollbackOnError=true) — either every component lands or none do — so an
// entire subscription costs ONE deploy() call and one status poll regardless
// of object count, and a failed deploy leaves nothing to tear down.
//
// Each component's param sets must describe the same pair: same object, and
// the flow's OutboundMessageName must match the outbound message being
// deployed. Objects must be unique across components (one workflow file per
// object).
//
// No dependency declaration is needed (or possible): package.xml only lists
// members, and Salesforce resolves references between components of the same
// deployment automatically — the flow's actionName reference to the outbound
// message is satisfied because the outbound message ships in the same
// package. See "Deploying and Retrieving Metadata":
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_deploy.htm
func ConstructFlowSubscription(components []FlowSubscriptionComponent) ([]byte, error) {
	if len(components) == 0 {
		return nil, errNoFlowComponents
	}

	entries := make(map[string][]byte, 2*len(components)+1) //nolint:mnd
	omMembers := make([]string, 0, len(components))
	flowMembers := make([]string, 0, len(components))
	seenObjects := make(map[string]bool, len(components))

	for _, component := range components {
		omParams, flowParams := component.OutboundMessage, component.Flow

		if err := validateFlowSubscriptionComponent(component, seenObjects); err != nil {
			return nil, err
		}

		seenObjects[omParams.ObjectName] = true

		workflowContent, err := generateWorkflowXML(omParams)
		if err != nil {
			return nil, err
		}

		flowContent, err := GenerateFlowXML(flowParams, "Active")
		if err != nil {
			return nil, err
		}

		entries["workflows/"+omParams.ObjectName+".workflow"] = []byte(workflowContent)
		entries["flows/"+flowParams.FlowName+".flow"] = []byte(flowContent)

		omMembers = append(omMembers, OutboundMessageFullName(omParams.ObjectName, omParams.Name))
		flowMembers = append(flowMembers, flowParams.FlowName)
	}

	pkg := triggerPackageXML{
		Xmlns:   metadataXmlns,
		Version: core.APIVersion,
		Types: []triggerPackageType{
			{
				Members: omMembers,
				Name:    "WorkflowOutboundMessage",
			},
			{
				Members: flowMembers,
				Name:    "Flow",
			},
		},
	}

	pkgXML, err := xml.MarshalIndent(pkg, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package.xml: %w", err)
	}

	entries["package.xml"] = []byte(xml.Header + string(pkgXML))

	return buildZip(entries)
}
