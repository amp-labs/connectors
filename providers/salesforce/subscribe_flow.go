package salesforce

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/metadata"
)

// This file implements the flow-based subscription path, selected via
// SubscriptionRequest.UseFlow. Instead of CDC channel members + apex triggers,
// it deploys one record-triggered Flow and one Workflow Outbound Message per
// subscribed object: the flow fires after a record is created/updated and the
// outbound message POSTs a SOAP notification to the configured endpoint.
//
// subscribeWithFlow is self-sufficient: every Salesforce-side dependency is
// created inside the method itself —
//
//  1. Resolve the integration user (from config, or from the connected user's
//     identity when not configured).
//  2. Deploy ALL objects' outbound messages and flows in ONE Metadata API
//     package: one deploy() call and one status poll for the whole
//     subscription, regardless of object count.
//
// The package deploys with rollbackOnError=true, so the subscription is
// all-or-nothing: every component lands or none do. There is no subscribe-time
// rollback path at all — a failed deploy leaves nothing behind — and on
// success every component is recorded in SubscribeResult.Flows, so the result
// always faithfully describes what exists in Salesforce. DeleteSubscription
// tears those components down dependents-first (flow, then outbound message).
//
// Trade-offs vs the CDC path:
//   - No Apex is deployed, so there is no test-coverage gate and no
//     "Variable does not exist" compiler race.
//   - No registration (event channel / named credential / event relay) is
//     required; the endpoint URL is plain HTTPS.
//   - DELETE events are NOT supported: outbound messages only exist for
//     create/update (delete-triggered flows run before-delete and cannot call
//     the outbound message action). Requested delete events are dropped with
//     a warning.
//   - Watched-field filtering uses an ISCHANGED() entry condition instead of
//     an indicator field, and only applies to Update-only subscriptions.

var (
	errFlowConfigMissing = errors.New(
		"SubscriptionRequest.Flow config with a non-empty endpointURL is required when UseFlow is true")
	errFlowNoSupportedEvents = errors.New(
		"flow-based subscriptions support only create/update events")
	errFlowDeployFailed            = errors.New("flow subscription deployment failed")
	errFlowDestructiveDeployFailed = errors.New("flow subscription teardown deployment failed")
)

// FlowSubscriptionConfig carries the flow-path settings on SubscriptionRequest.
// Required when UseFlow is true.
type FlowSubscriptionConfig struct {
	// EndpointURL is the HTTPS endpoint Salesforce POSTs outbound messages to.
	// Required.
	EndpointURL string `json:"endpointUrl"`

	// IntegrationUsername is the Salesforce username recorded as the outbound
	// message's integration user (required by the metadata type). Optional:
	// when empty, subscribeWithFlow resolves the connected user's username via
	// the identity endpoint so the method stays self-sufficient.
	IntegrationUsername string `json:"integrationUsername,omitempty"`

	// Fields lists the object fields included in each outbound message payload.
	// Defaults to ["Id"] — consumers fetch the full record via the API.
	Fields []string `json:"fields,omitempty"`
}

// OutboundMessageMetadata records one deployed workflow outbound message —
// the dependency the object's flow invokes. Everything needed to regenerate
// its XML (for idempotent re-deploys) and to destructively remove it lives
// here, so teardown needs no access to the original request.
//
// The deployed component is a WorkflowOutboundMessage; official docs with
// example XML:
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_workflow.htm
type OutboundMessageMetadata struct {
	// Name is the outbound message developer name WITHOUT the object prefix;
	// the deployable full name is "<ObjectName>.<Name>".
	Name string `json:"name"`

	// EndpointURL, IntegrationUsername, and Fields mirror what was deployed.
	EndpointURL         string   `json:"endpointUrl"`
	IntegrationUsername string   `json:"integrationUsername"`
	Fields              []string `json:"fields,omitempty"`

	// DeployID is the Metadata API deploy id that created the component.
	DeployID string `json:"deployId,omitempty"`
}

// FlowMetadata records one deployed record-triggered flow.
//
// The deployed component is a Flow; official docs with example XML:
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_visual_workflow.htm
type FlowMetadata struct {
	// Name is the Flow API name (org-wide unique).
	Name string `json:"name"`

	// RecordTriggerType is the deployed flow's trigger (Create / Update /
	// CreateAndUpdate). Kept so teardown can regenerate the exact flow XML for
	// the deactivation fallback.
	RecordTriggerType string `json:"recordTriggerType"`

	// WatchFields are the fields baked into the flow's ISCHANGED entry
	// condition (Update-only subscriptions), empty otherwise.
	WatchFields []string `json:"watchFields,omitempty"`

	// DeployID is the Metadata API deploy id that created the component.
	DeployID string `json:"deployId,omitempty"`
}

// FlowSubscription describes the Salesforce components created for one object
// by subscribeWithFlow, in creation order. Artifacts are recorded as each
// creation step succeeds and set back to nil as teardown removes them, so the
// struct always describes what exists in Salesforce. All fields are exported
// to survive JSON round-trips of the persisted result.
type FlowSubscription struct {
	ObjectName common.ObjectName `json:"objectName"`

	// OutboundMessage is created first (the flow depends on it).
	OutboundMessage *OutboundMessageMetadata `json:"outboundMessage,omitempty"`

	// Flow is created second, referencing the outbound message.
	Flow *FlowMetadata `json:"flow,omitempty"`

	// Events are the normalized event types this subscription actually covers
	// (create/update; never delete).
	Events []common.SubscriptionEventType `json:"events"`
}

// subscribeWithFlow creates flow-based subscriptions for every object in
// params.SubscriptionEvents. Unlike the CDC path it needs no registration.
//
// All objects' components deploy in one atomic package, so there are only
// three outcomes: everything subscribed (Success), nothing created (Failed —
// the common failure shape, no rollback needed), or a poll timeout where the
// deploy is still queued org-side and may land later — the intended components
// are recorded on the Failed result so DeleteSubscription can clean up
// whatever eventually materialized.
func (c *Connector) subscribeWithFlow(
	ctx context.Context,
	params common.SubscribeParams,
	req *SubscriptionRequest,
) (*common.SubscriptionResult, error) {
	if req.Flow == nil || req.Flow.EndpointURL == "" {
		return nil, errFlowConfigMissing
	}

	config, err := c.resolveFlowConfig(ctx, req.Flow)
	if err != nil {
		return nil, err
	}

	// Build every object's component up-front: nothing has been deployed yet,
	// so any validation error aborts with zero Salesforce-side mutations.
	components := make([]metadata.FlowSubscriptionComponent, 0, len(params.SubscriptionEvents))
	pending := make([]*FlowSubscription, 0, len(params.SubscriptionEvents))

	for objName, objEvents := range params.SubscriptionEvents {
		component, flowSub, err := buildFlowSubscriptionComponent(ctx, objName, objEvents, config)
		if err != nil {
			return nil, fmt.Errorf("failed to build flow subscription for object %s: %w", objName, err)
		}

		components = append(components, component)
		pending = append(pending, flowSub)
	}

	zipData, err := metadata.ConstructFlowSubscription(components)
	if err != nil {
		return nil, fmt.Errorf("failed to construct flow subscription package: %w", err)
	}

	sfRes := &SubscribeResult{
		UseFlow: true,
		Flows:   make(map[common.ObjectName]*FlowSubscription),
	}

	deployID, err := c.deployFlowMetadataZip(ctx, "flow subscription package", zipData)
	if err != nil {
		// rollbackOnError=true makes the deploy all-or-nothing, so a deploy
		// failure means nothing was created. The exception is a poll timeout:
		// the deploy is still queued org-side and may land later, so record
		// the intended components for a later DeleteSubscription to clean up.
		if errors.Is(err, ErrDeployPollTimeout) {
			recordFlowSubscriptions(sfRes, pending, "")
		}

		return &common.SubscriptionResult{
			Status:  common.SubscriptionStatusFailed,
			Result:  sfRes,
			Objects: datautils.FromMap(sfRes.Flows).Keys(),
		}, err
	}

	recordFlowSubscriptions(sfRes, pending, deployID)

	return &common.SubscriptionResult{
		Status:  common.SubscriptionStatusSuccess,
		Result:  sfRes,
		Events:  flowCoveredEvents(sfRes),
		Objects: datautils.FromMap(sfRes.Flows).Keys(),
	}, nil
}

// recordFlowSubscriptions attaches the pending per-object records to the
// result, stamping the deploy id that created their components.
func recordFlowSubscriptions(sfRes *SubscribeResult, pending []*FlowSubscription, deployID string) {
	for _, flowSub := range pending {
		flowSub.OutboundMessage.DeployID = deployID
		flowSub.Flow.DeployID = deployID
		sfRes.Flows[flowSub.ObjectName] = flowSub
	}
}

// resolveFlowConfig fills in the config pieces the caller left out so the
// subscribe path is self-sufficient: currently the integration username,
// resolved from the connected user's identity (see authmetadata.go — callers
// holding a connection record can avoid this lookup entirely by passing the
// post-auth "username" provider metadata as IntegrationUsername). Returns a
// copy — the caller's request is never mutated.
func (c *Connector) resolveFlowConfig(
	ctx context.Context, config *FlowSubscriptionConfig,
) (*FlowSubscriptionConfig, error) {
	resolved := *config

	if resolved.IntegrationUsername == "" {
		username, err := c.GetCurrentUsername(ctx)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to resolve the integration user from the connected user's identity "+
					"(set SubscriptionRequest.Flow.IntegrationUsername to skip the lookup): %w", err)
		}

		logging.Logger(ctx).InfoContext(ctx,
			"no integration user configured; using the connected user's username",
			"username", username,
		)

		resolved.IntegrationUsername = username
	}

	return &resolved, nil
}

// buildFlowSubscriptionComponent assembles one object's package component and
// its (not yet deployed) result record.
func buildFlowSubscriptionComponent(
	ctx context.Context,
	objName common.ObjectName,
	objEvents common.ObjectEvents,
	config *FlowSubscriptionConfig,
) (metadata.FlowSubscriptionComponent, *FlowSubscription, error) {
	triggerType, coveredEvents, err := flowRecordTriggerType(ctx, objName, objEvents)
	if err != nil {
		return metadata.FlowSubscriptionComponent{}, nil, err
	}

	omParams, flowParams, err := buildFlowSubscriptionParams(objName, objEvents, config, triggerType)
	if err != nil {
		return metadata.FlowSubscriptionComponent{}, nil, err
	}

	component := metadata.FlowSubscriptionComponent{
		OutboundMessage: omParams,
		Flow:            flowParams,
	}

	flowSub := &FlowSubscription{
		ObjectName: objName,
		Events:     coveredEvents,
		OutboundMessage: &OutboundMessageMetadata{
			Name:                omParams.Name,
			EndpointURL:         config.EndpointURL,
			IntegrationUsername: config.IntegrationUsername,
			Fields:              config.Fields,
		},
		Flow: &FlowMetadata{
			Name:              flowParams.FlowName,
			RecordTriggerType: string(triggerType),
			WatchFields:       flowParams.WatchFields,
		},
	}

	return component, flowSub, nil
}

// deployFlowMetadataZip deploys one zip and polls the deploy to completion,
// returning the deploy id.
func (c *Connector) deployFlowMetadataZip(
	ctx context.Context, entity string, zipData []byte,
) (string, error) {
	deployID, err := c.DeployMetadataZip(ctx, zipData)
	if err != nil {
		return "", fmt.Errorf("failed to deploy %s: %w", entity, err)
	}

	deployResult, err := c.pollDeployStatus(ctx, deployID)
	if err != nil {
		return "", fmt.Errorf("failed to poll %s deploy status: %w", entity, err)
	}

	if !deployResult.Success {
		return "", fmt.Errorf("%w: %s: %s",
			errFlowDeployFailed, entity, formatDeployFailureDetails(deployResult))
	}

	return deployID, nil
}

// buildFlowSubscriptionParams assembles the metadata construction params for
// one object's outbound message + flow pair.
func buildFlowSubscriptionParams(
	objName common.ObjectName,
	objEvents common.ObjectEvents,
	config *FlowSubscriptionConfig,
	triggerType metadata.RecordTriggerType,
) (metadata.OutboundMessageParams, metadata.FlowParams, error) {
	flowName, err := metadata.GenerateFlowNameForSubscription(string(objName))
	if err != nil {
		return metadata.OutboundMessageParams{}, metadata.FlowParams{}, err
	}

	omName, err := metadata.GenerateOutboundMessageNameForSubscription(string(objName))
	if err != nil {
		return metadata.OutboundMessageParams{}, metadata.FlowParams{}, err
	}

	watchFields := objEvents.WatchFields
	if triggerType != metadata.RecordTriggerTypeUpdate {
		// The ISCHANGED entry condition is only generated for Update-only
		// flows; don't record watch fields the deployed flow won't enforce.
		watchFields = nil
	}

	omParams := metadata.OutboundMessageParams{
		ObjectName:          string(objName),
		Name:                omName,
		EndpointURL:         config.EndpointURL,
		IntegrationUsername: config.IntegrationUsername,
		Fields:              config.Fields,
	}

	flowParams := metadata.FlowParams{
		ObjectName:          string(objName),
		FlowName:            flowName,
		OutboundMessageName: omName,
		RecordTriggerType:   triggerType,
		WatchFields:         watchFields,
	}

	return omParams, flowParams, nil
}

// flowRecordTriggerType maps normalized subscription events onto a flow
// recordTriggerType. Delete events are dropped with a warning (outbound
// messages cannot fire on delete); an object with no create/update event
// is an error rather than a silent no-op subscription.
func flowRecordTriggerType(
	ctx context.Context, objName common.ObjectName, objEvents common.ObjectEvents,
) (metadata.RecordTriggerType, []common.SubscriptionEventType, error) {
	hasCreate, hasUpdate, hasDelete := scanFlowEventTypes(objEvents)

	if hasDelete {
		logging.Logger(ctx).WarnContext(ctx,
			"flow-based subscriptions cannot deliver delete events; dropping the delete event",
			"object", objName,
		)
	}

	switch {
	case hasCreate && hasUpdate:
		return metadata.RecordTriggerTypeCreateAndUpdate,
			[]common.SubscriptionEventType{
				common.SubscriptionEventTypeCreate,
				common.SubscriptionEventTypeUpdate,
			}, nil
	case hasCreate:
		return metadata.RecordTriggerTypeCreate,
			[]common.SubscriptionEventType{common.SubscriptionEventTypeCreate}, nil
	case hasUpdate:
		return metadata.RecordTriggerTypeUpdate,
			[]common.SubscriptionEventType{common.SubscriptionEventTypeUpdate}, nil
	default:
		return "", nil, fmt.Errorf("%w: object %s requested %v", errFlowNoSupportedEvents, objName, objEvents.Events)
	}
}

// scanFlowEventTypes reports which of the normalized event types the object's
// subscription requests.
func scanFlowEventTypes(objEvents common.ObjectEvents) (hasCreate, hasUpdate, hasDelete bool) {
	for _, event := range objEvents.Events {
		switch event { //nolint:exhaustive
		case common.SubscriptionEventTypeCreate:
			hasCreate = true
		case common.SubscriptionEventTypeUpdate:
			hasUpdate = true
		case common.SubscriptionEventTypeDelete:
			hasDelete = true
		}
	}

	return hasCreate, hasUpdate, hasDelete
}

// flowCoveredEvents returns the union of event types covered across all
// deployed flows, for the deprecated top-level SubscriptionResult.Events field.
func flowCoveredEvents(sfRes *SubscribeResult) []common.SubscriptionEventType {
	seen := make(map[common.SubscriptionEventType]bool)
	union := make([]common.SubscriptionEventType, 0, 2) //nolint:mnd

	for _, flowSub := range sfRes.Flows {
		if flowSub == nil {
			continue
		}

		for _, event := range flowSub.Events {
			if !seen[event] {
				seen[event] = true

				union = append(union, event)
			}
		}
	}

	return union
}

// deleteFlowSubscriptions tears down every flow-based subscription recorded in
// sfRes. Artifacts are removed dependents-first (flow, then outbound message),
// and each component is cleared from its FlowSubscription the moment it is gone
// — an object's map entry is removed only once every component is gone — so the
// surviving entries faithfully describe what is still in Salesforce (mirroring
// the CDC DeleteSubscription contract). The first failure aborts.
func (c *Connector) deleteFlowSubscriptions(ctx context.Context, sfRes *SubscribeResult) error {
	for objName, flowSub := range sfRes.Flows {
		if flowSub == nil {
			logging.Logger(ctx).WarnContext(ctx, "flow subscription entry is nil, skipping delete",
				"object", objName,
			)

			continue
		}

		if err := c.deleteFlowSubscription(ctx, flowSub); err != nil {
			logging.Logger(ctx).WarnContext(ctx,
				"flow subscription delete failed mid-teardown; aborting before remaining flows are touched",
				"failedObject", objName,
				"error", err,
				"remainingFlows", datautils.FromMap(sfRes.Flows).Keys(),
			)

			return fmt.Errorf("failed to delete flow subscription for object '%s': %w", objName, err)
		}

		delete(sfRes.Flows, objName)
	}

	return nil
}

// deleteFlowSubscription removes one object's components in inverse creation
// order, nil-ing each on flowSub as it is removed:
//
//  1. The flow. Salesforce rejects deletion of active flows, so it is
//     deactivated first (Tooling API FlowDefinition PATCH with
//     activeVersionNumber = 0; if the Tooling route fails, fall back to
//     redeploying the flow as Draft). A flow that is already gone skips both.
//  2. The outbound message the flow referenced.
func (c *Connector) deleteFlowSubscription(ctx context.Context, flowSub *FlowSubscription) error {
	if flowSub.Flow != nil {
		if err := c.deleteFlowMetadata(ctx, flowSub); err != nil {
			return err
		}

		flowSub.Flow = nil
	}

	if flowSub.OutboundMessage != nil {
		omZipData, err := metadata.ConstructDestructiveOutboundMessage(
			string(flowSub.ObjectName), flowSub.OutboundMessage.Name,
		)
		if err != nil {
			return fmt.Errorf("failed to construct destructive outbound message zip for %s: %w",
				flowSub.OutboundMessage.Name, err)
		}

		omFullName := metadata.OutboundMessageFullName(string(flowSub.ObjectName), flowSub.OutboundMessage.Name)
		if err := c.deployFlowDestructiveZip(ctx, omZipData, "outbound message "+omFullName); err != nil {
			return err
		}

		flowSub.OutboundMessage = nil
	}

	return nil
}

// deleteFlowMetadata deactivates and destructively removes the flow.
func (c *Connector) deleteFlowMetadata(ctx context.Context, flowSub *FlowSubscription) error {
	flowExists, err := c.deactivateFlow(ctx, flowSub)
	if err != nil {
		return err
	}

	if !flowExists {
		return nil
	}

	zipData, err := metadata.ConstructDestructiveFlow(flowSub.Flow.Name)
	if err != nil {
		return fmt.Errorf("failed to construct destructive flow zip for %s: %w", flowSub.Flow.Name, err)
	}

	return c.deployFlowDestructiveZip(ctx, zipData, "flow "+flowSub.Flow.Name)
}

// deployFlowDestructiveZip deploys one destructive-changes package with
// NoTestRun (no Apex is involved, so production orgs allow it) and polls to
// completion.
func (c *Connector) deployFlowDestructiveZip(ctx context.Context, zipData []byte, entity string) error {
	deployResult, err := c.deployDestructiveApex(ctx, zipData, metadata.TestLevelNoTestRun)
	if err != nil {
		return fmt.Errorf("failed to deploy destructive change for %s: %w", entity, err)
	}

	if !deployResult.Success {
		return fmt.Errorf("%w for %s: %s",
			errFlowDestructiveDeployFailed, entity, formatDeployFailureDetails(deployResult))
	}

	return nil
}

// deactivateFlow makes the flow deletable. Returns whether the flow still
// exists in Salesforce: false means it is already gone and the destructive
// flow deploy must be skipped (a destructive deploy referencing a missing
// component fails the whole package).
//
// Primary route: Tooling API PATCH on FlowDefinition with
// Metadata.activeVersionNumber = 0 — see the FlowDefinition Tooling API
// object, whose activeVersionNumber field controls which version is active:
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_flowdefinition.htm
// Fallback: redeploy the flow as Draft via the Metadata API, regenerated from
// the stored FlowSubscription components.
func (c *Connector) deactivateFlow(ctx context.Context, flowSub *FlowSubscription) (bool, error) {
	definitionID, err := c.findToolingEntityIDByDeveloperName(ctx, "FlowDefinition", flowSub.Flow.Name)
	if err != nil {
		if errors.Is(err, errToolingEntityNotFound) {
			logging.Logger(ctx).InfoContext(ctx, "flow already absent, skipping deactivation and deletion",
				"flow", flowSub.Flow.Name,
			)

			return false, nil
		}

		return false, fmt.Errorf("failed to look up FlowDefinition for %s: %w", flowSub.Flow.Name, err)
	}

	body := map[string]any{
		"Metadata": map[string]any{
			"activeVersionNumber": 0,
		},
	}

	_, patchErr := c.patchToSFAPI(
		ctx, body, "tooling/sobjects/FlowDefinition/"+definitionID, "FlowDefinition",
	)
	if patchErr == nil {
		return true, nil
	}

	logging.Logger(ctx).WarnContext(ctx,
		"tooling API flow deactivation failed; falling back to redeploying the flow as Draft",
		"flow", flowSub.Flow.Name,
		"error", patchErr,
	)

	zipData, err := metadata.ConstructFlow(flowParamsFromResult(flowSub), "Draft")
	if err != nil {
		return true, fmt.Errorf("failed to construct flow deactivation zip for %s: %w", flowSub.Flow.Name, err)
	}

	if err := c.deployFlowDestructiveZip(ctx, zipData, "flow deactivation "+flowSub.Flow.Name); err != nil {
		return true, err
	}

	return true, nil
}

// flowParamsFromResult reconstructs the flow metadata params from a stored
// FlowSubscription so teardown can regenerate the exact flow XML. Requires
// flowSub.Flow and flowSub.OutboundMessage to be present.
func flowParamsFromResult(flowSub *FlowSubscription) metadata.FlowParams {
	params := metadata.FlowParams{
		ObjectName: string(flowSub.ObjectName),
	}

	if flowSub.Flow != nil {
		params.FlowName = flowSub.Flow.Name
		params.RecordTriggerType = metadata.RecordTriggerType(flowSub.Flow.RecordTriggerType)
		params.WatchFields = flowSub.Flow.WatchFields
	}

	if flowSub.OutboundMessage != nil {
		params.OutboundMessageName = flowSub.OutboundMessage.Name
	}

	return params
}

// updateSubscriptionWithFlow handles UpdateSubscription whenever the flow path
// is involved on either side (previous state used flows, or the new request
// asks for them, including CDC↔flow transitions). Reconciliation is
// delete-then-recreate rather than incremental: the previous subscription is
// fully torn down (DeleteSubscription branches per mode), then Subscribe
// creates the new one. There is therefore a short window with no active
// subscription; events occurring in that window are lost. Acceptable for the
// exploration phase — an incremental reconcile can replace this later.
func (c *Connector) updateSubscriptionWithFlow(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
	prevState *SubscribeResult,
) (*common.SubscriptionResult, error) {
	deleteParams := *previousResult
	deleteParams.Result = prevState

	if err := c.DeleteSubscription(ctx, deleteParams); err != nil {
		// prevState was pruned in place by DeleteSubscription, so it reflects
		// exactly what is still in Salesforce.
		return &common.SubscriptionResult{
			Status: common.SubscriptionStatusFailed,
			Result: prevState,
		}, fmt.Errorf("failed to delete previous subscription before flow update: %w", err)
	}

	return c.Subscribe(ctx, params)
}
