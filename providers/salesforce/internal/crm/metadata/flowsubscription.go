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
// and/or updated and invokes a Workflow Outbound Message action, which POSTs
// a SOAP notification to the configured endpoint.
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

// validateFlowSubscriptionComponent checks one OM+flow pair before it is added
// to the combined subscribe zip: both param sets are valid, they name the same
// object and outbound message, and that object is not already in the package.
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

// RecordTriggerType maps to Flow.start.recordTriggerType in the Metadata API
// (Create, Update, CreateAndUpdate). We do not use Delete: record-triggered
// delete flows cannot invoke an outbound message.
type RecordTriggerType string

const (
	RecordTriggerTypeCreate          RecordTriggerType = "Create"
	RecordTriggerTypeUpdate          RecordTriggerType = "Update"
	RecordTriggerTypeCreateAndUpdate RecordTriggerType = "CreateAndUpdate"
)

// FlowParams contains the parameters for constructing the deploy package of
// one record-triggered flow that invokes an outbound message.
// See the official docs (with declaration details and example XML):
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_visual_workflow.htm
type FlowParams struct {
	// ObjectName is the Salesforce object the flow runs on (e.g., "Lead").
	ObjectName string

	// FlowName is the Flow API name (e.g., "acme_Lead").
	FlowName string

	// OutboundMessageName is the developer name (WITHOUT the object prefix)
	// of the outbound message the flow invokes.
	OutboundMessageName string

	// RecordTriggerType selects when the flow fires (Create / Update /
	// CreateAndUpdate).
	RecordTriggerType RecordTriggerType

	// WatchFields optionally narrows when the flow runs via an entry-condition
	// formula.
	WatchFields []string
}

// ValidateFlowParams checks that all required fields are present and the
// trigger type is valid.
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
// to one outbound message action call. Status is "Active" on subscribe and
// "Draft" on the deactivation fallback.
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
// applies WatchFields, or "" when none are configured or the trigger is
// Create-only.
//
// Salesforce does not allow a bare ISCHANGED() entry condition on flows that
// fire on create: there is no prior value, the Is Changed operator is
// documented as update-only, and ISCHANGED() on create fails as working as
// expected:
//
//	https://help.salesforce.com/s/articleView?id=platform.flow_ref_elements_start_object.htm
//	https://help.salesforce.com/s/issue?id=a028c00000j5k7eAAA
//
// For CreateAndUpdate, use OR with ISNEW() so creates still notify and updates
// stay gated on watched fields — the documented pattern:
//
//	https://admin.salesforce.com/blog/2022/three-powerful-formula-functions-you-can-use-in-flow
func buildWatchFieldsFilterFormula(params FlowParams) string {
	if len(params.WatchFields) == 0 || params.RecordTriggerType == RecordTriggerTypeCreate {
		return ""
	}

	changed := make([]string, 0, len(params.WatchFields))
	for _, field := range params.WatchFields {
		changed = append(changed, fmt.Sprintf("ISCHANGED({!$Record.%s})", field))
	}

	if params.RecordTriggerType == RecordTriggerTypeCreateAndUpdate {
		return "OR(ISNEW(), " + strings.Join(changed, ", ") + ")"
	}

	if len(changed) == 1 {
		return changed[0]
	}

	return "OR(" + strings.Join(changed, ", ") + ")"
}

// ConstructDraftFlow builds an UPDATE zip with status Draft so the flow is
// deactivated. Salesforce refuses to delete an Active flow, so teardown
// deactivates first. Prefer Tooling API PATCH of FlowDefinition
// (activeVersionNumber=0): it skips the org-wide Metadata deploy queue, does
// not create a new flow version, and does not need this XML rebuilt. Use this
// zip only when that PATCH fails. After the flow is inactive,
// ConstructDestructiveFlow removes the versions from the org.
func ConstructDraftFlow(params FlowParams) ([]byte, error) {
	if err := ValidateFlowParams(params); err != nil {
		return nil, err
	}

	flowContent, err := GenerateFlowXML(params, "Draft")
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

// FlowVersionMember returns the Metadata API member for one flow version
// (e.g. acme_Account-2). Destructive deploys must use this form:
// an unversioned Flow member fails with "insufficient access rights on
// cross-reference id" once a definition has more than one version
// (Salesforce known issue):
// https://help.salesforce.com/s/issue?id=a028c00000gAwixAAC
func FlowVersionMember(flowName string, version int) string {
	return fmt.Sprintf("%s-%d", flowName, version)
}

// ConstructDestructiveFlow builds a DELETE zip for the given flow versions
// (destructiveChanges.xml; no .flow file). The flow must already be inactive
// (Tooling or ConstructDraftFlow). Members are version-suffixed; see FlowVersionMember.
func ConstructDestructiveFlow(flowName string, versions []int) ([]byte, error) {
	if flowName == "" {
		return nil, errEmptyObjectName
	}

	if len(versions) == 0 {
		return nil, errNoFlowComponents
	}

	members := make([]string, 0, len(versions))
	for _, version := range versions {
		members = append(members, FlowVersionMember(flowName, version))
	}

	return constructDestructiveZip(triggerPackageType{
		Members: members,
		Name:    "Flow",
	})
}

// ConstructFlowSubscription builds ONE deploy package for a flow-based
// subscription: every component's workflow outbound message and Active
// record-triggered flow. Salesforce metadata deploys are atomic per package
// (our deploy options set rollbackOnError=true) — either every component
// lands or none do — so an entire subscription costs ONE deploy() call and
// one status poll regardless of object count, and a failed deploy leaves
// nothing to tear down.
//
// No dependency declaration is needed: Salesforce resolves references between
// components of the same deployment automatically — the flow's actionName reference
// to the outbound message is satisfied because the outbound message ships in the same
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
