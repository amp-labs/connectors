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
// success every component is recorded in SubscribeResult.Flows. DeleteSubscription
// tears those components down dependents-first (flow, then outbound message).
//
// Trade-offs vs the CDC path:
//   - No Apex is deployed, so there is no test-coverage gate and no
//     "Variable does not exist" compiler race.
//   - No registration (event channel / named credential / event relay) is
//     required; the endpoint URL is plain HTTPS.
//   - DELETE events are NOT supported: outbound messages only exist for
//     create/update (delete-triggered flows run before-delete and cannot call
//     the outbound message action). Callers must not request delete; this
//     path errors if they do.

var (
	errFlowEndpointURLRequired = errors.New(
		"SubscriptionRequest.Flow.EndpointURL is required when UseFlow is true")
	errFlowNoSupportedEvents = errors.New(
		"flow-based subscriptions support only create/update events")
	errFlowDeployFailed            = errors.New("flow subscription deployment failed")
	errFlowTeardownDeployFailed = errors.New("flow subscription teardown deployment failed")
)

// FlowConfig carries the flow-path settings on SubscriptionRequest.
// Required when UseFlow is true.
type FlowConfig struct {
	// EndpointURL is the HTTPS endpoint Salesforce POSTs outbound messages to.
	// Required.
	EndpointURL string `json:"endpointUrl"`

	// IntegrationUsername is the Salesforce username recorded as the outbound
	// message's integration user (required by the metadata type). Optional:
	// when empty, subscribeWithFlow resolves the connected user's username via
	// the identity endpoint so the method stays self-sufficient.
	IntegrationUsername string `json:"integrationUsername,omitempty"`

	// NamePrefix is prepended to the deployed artifact names, giving
	// "<NamePrefix>_<Object>" for both the flow and its outbound message. Pass
	// the project's app name: the flow appears in the org's Flow list, which the
	// end customer browses.
	NamePrefix string `json:"namePrefix"`

	// Fields lists additional object fields to include in each outbound message
	// payload. Optional: Id, CreatedDate and LastModifiedDate are always included
	// on top of whatever is listed here, so leaving this empty yields a payload of
	// exactly those three. Id is how consumers fetch the full record; the audit timestamps
	// let the parser infer create vs update.
	Fields []string `json:"fields,omitempty"`
}

// OutboundMessageMetadata is our record of a deployed OM, stored on
// SubscribeResult.Flows[object].OutboundMessage. Delete uses Name for the
// destructive zip and as OutboundMessageName if the flow is Draft-redeployed.
// Flow updates delete-then-recreate, so they only read this on that delete.
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_workflow.htm
type OutboundMessageMetadata struct {
	// Name is the outbound message developer name WITHOUT the object prefix;
	// the deployable full name is "<ObjectName>.<Name>".
	Name string `json:"name"`

	// DeployID is the Metadata API deploy id that created the component.
	DeployID string `json:"deployId,omitempty"`
}

// FlowMetadata is our record of a deployed flow, stored on
// SubscribeResult.Flows[object].Flow. Delete uses Name plus RecordTriggerType
// and WatchFields to rebuild XML for the Draft deactivation fallback, then
// deletes versions. Update only reads this on that delete half.
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_visual_workflow.htm
type FlowMetadata struct {
	// Name is the Flow API name (org-wide unique).
	Name string `json:"name"`

	// RecordTriggerType is the deployed flow's trigger (Create / Update /
	// CreateAndUpdate). Kept so teardown can regenerate the exact flow XML for
	// the deactivation fallback.
	RecordTriggerType string `json:"recordTriggerType"`

	// WatchFields are the fields baked into the flow's entry-condition
	// formula (Update and CreateAndUpdate). Create-only ignores them.
	WatchFields []string `json:"watchFields,omitempty"`

	// DeployID is the Metadata API deploy id that created the component.
	DeployID string `json:"deployId,omitempty"`
}

// FlowSubscription is the persisted per-object record on SubscribeResult.Flows:
// the OM and flow bookmarks for one object, written after the combined deploy
// succeeds.
type FlowSubscription struct {
	ObjectName common.ObjectName `json:"objectName"`

	// OutboundMessage is the OM deployed in the same package as Flow.
	OutboundMessage *OutboundMessageMetadata `json:"outboundMessage,omitempty"`

	// Flow is the record-triggered flow deployed in the same package as the OM.
	Flow *FlowMetadata `json:"flow,omitempty"`

	// Events are the normalized event types this subscription covers
	// (create/update; never delete).
	Events []common.SubscriptionEventType `json:"events"`
}

// subscribeWithFlow creates flow-based subscriptions for every object in
// params.SubscriptionEvents.
//
// Unlike the CDC path, this is one Metadata API deploy (all objects' OMs and
// flows in one zip, rollbackOnError=true). Salesforce either commits the
// whole package or rolls it back itself. The connector has no extra steps to
// undo, so there is no rollbackSubscribe and no FailedToRollback.
//
//nolint:cyclop,funlen
func (c *Connector) subscribeWithFlow(
	ctx context.Context,
	params common.SubscribeParams,
	req *SubscriptionRequest,
) (*common.SubscriptionResult, error) {
	if req.Flow == nil || req.Flow.EndpointURL == "" {
		return nil, errFlowEndpointURLRequired
	}

	if req.Flow.IntegrationUsername == "" {
		username, err := c.GetConnectedUsername(ctx)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to resolve the integration user from the connected user's identity: %w", err)
		}

		logging.Logger(ctx).InfoContext(ctx,
			"no integration user configured; using the connected user's username",
			"username", username,
		)

		req.Flow.IntegrationUsername = username
	}

	components := make([]metadata.FlowSubscriptionComponent, 0, len(params.SubscriptionEvents))
	pending := make([]*FlowSubscription, 0, len(params.SubscriptionEvents))

	for objName, objEvents := range params.SubscriptionEvents {
		triggerType, err := flowRecordTriggerType(objEvents)
		if err != nil {
			return nil, fmt.Errorf("failed to build flow subscription for object %s: %w", objName, err)
		}

		artifactName, err := metadata.GenerateSubscriptionArtifactName(req.Flow.NamePrefix, string(objName))
		if err != nil {
			return nil, fmt.Errorf("failed to build flow subscription for object %s: %w", objName, err)
		}

		components = append(components, metadata.FlowSubscriptionComponent{
			OutboundMessage: metadata.OutboundMessageParams{
				ObjectName:          string(objName),
				Name:                artifactName,
				EndpointURL:         req.Flow.EndpointURL,
				IntegrationUsername: req.Flow.IntegrationUsername,
				Fields:              req.Flow.Fields,
			},
			Flow: metadata.FlowParams{
				ObjectName:          string(objName),
				FlowName:            artifactName,
				OutboundMessageName: artifactName,
				RecordTriggerType:   triggerType,
				WatchFields:         objEvents.WatchFields,
			},
		})
		pending = append(pending, &FlowSubscription{
			ObjectName: objName,
			Events:     objEvents.Events,
			OutboundMessage: &OutboundMessageMetadata{
				Name: artifactName,
			},
			Flow: &FlowMetadata{
				Name:              artifactName,
				RecordTriggerType: string(triggerType),
				WatchFields:       objEvents.WatchFields,
			},
		})
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
		if !errors.Is(err, ErrDeployPollTimeout) {
			return &common.SubscriptionResult{
				Status:  common.SubscriptionStatusFailed,
				Result:  sfRes,
				Objects: datautils.FromMap(sfRes.Flows).Keys(),
			}, err
		}

		// Treat poll timeout as Success, matching CDC (deployTriggersToleratingTimeouts).
		// The Metadata deploy is still queued/running org-side (deploys serialize
		// per org, so queue wait can exceed the poll window) and will land on its
		// own. Returning Failed would make Temporal retry Subscribe and stack
		// another package. Events start when this zip commits.
		// TODO: surface a poll-timed-out marker on SubscribeResult so the server
		// can tell "Done" from "still queued" and watch Deployment Status.
		logging.Logger(ctx).WarnContext(ctx,
			"flow subscription deploy still running after poll timeout; proceeding without waiting — "+
				"events start when the package lands",
			"deployId", deployID,
		)
	}

	for _, flowSub := range pending {
		flowSub.OutboundMessage.DeployID = deployID
		flowSub.Flow.DeployID = deployID
		sfRes.Flows[flowSub.ObjectName] = flowSub
	}

	return &common.SubscriptionResult{
		Status:  common.SubscriptionStatusSuccess,
		Result:  sfRes,
		Events:  flowCoveredEvents(sfRes),
		Objects: datautils.FromMap(sfRes.Flows).Keys(),
	}, nil
}

// deployFlowMetadataZip submits one zip and polls until Salesforce reports a
// verdict. Poll timeout returns the deploy id and ErrDeployPollTimeout (the
// job is still running org-side). A completed Salesforce failure returns no
// id. Submit failures also return no id.
func (c *Connector) deployFlowMetadataZip(
	ctx context.Context, entity string, zipData []byte,
) (string, error) {
	deployID, err := c.DeployMetadataZip(ctx, zipData)
	if err != nil {
		return "", fmt.Errorf("failed to deploy %s: %w", entity, err)
	}

	deployResult, err := c.pollDeployStatus(ctx, deployID)
	if err != nil {
		return deployID, fmt.Errorf("failed to poll %s deploy status: %w", entity, err)
	}

	if !deployResult.Success {
		return "", fmt.Errorf("%w: %s: %s",
			errFlowDeployFailed, entity, formatDeployFailureDetails(deployResult))
	}

	return deployID, nil
}

// flowRecordTriggerType maps normalized subscription events onto a flow
// recordTriggerType. Delete is rejected (OM cannot fire on before-delete).
func flowRecordTriggerType(
	objEvents common.ObjectEvents,
) (metadata.RecordTriggerType, error) {
	var hasCreate, hasUpdate bool

	for _, event := range objEvents.Events {
		switch event { //nolint:exhaustive
		case common.SubscriptionEventTypeCreate:
			hasCreate = true
		case common.SubscriptionEventTypeUpdate:
			hasUpdate = true
		case common.SubscriptionEventTypeDelete: // rejected at API validation
			return "", fmt.Errorf("%w: delete events are not supported", errFlowNoSupportedEvents)
		}
	}

	switch {
	case hasCreate && hasUpdate:
		return metadata.RecordTriggerTypeCreateAndUpdate, nil
	case hasCreate:
		return metadata.RecordTriggerTypeCreate, nil
	case hasUpdate:
		return metadata.RecordTriggerTypeUpdate, nil
	default:
		return "", fmt.Errorf("%w: requested %v", errFlowNoSupportedEvents, objEvents.Events)
	}
}

// flowCoveredEvents returns the union of event types covered across all
// deployed flows.
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

// deleteAllFlowSubscriptions tears down every flow-based subscription in
// sfRes.Flows. Salesforce rejects deletion of active flows, so each flow is deactivated
// first (Tooling API FlowDefinition PATCH with activeVersionNumber = 0; if
// that fails, fall back to redeploying the flow as Draft). Versions are then
// deleted individually (unversioned Metadata delete fails once more than one
// version exists). A flow that is already gone skips both. Then the outbound
// message the flow referenced is removed.
func (c *Connector) deleteAllFlowSubscriptions(ctx context.Context, sfRes *SubscribeResult) error {
	for objName, flowSub := range sfRes.Flows {
		if err := c.deleteFlowSubscription(ctx, objName, flowSub); err != nil {
			return err
		}
	}

	return nil
}

// deleteFlowSubscription removes one object's artifacts dependents-first: the
// flow (deactivated, then every version) and then the outbound message it
// referenced. Shared by full teardown and by reconcile, which uses it for the
// objects that dropped out of a subscription.
func (c *Connector) deleteFlowSubscription(
	ctx context.Context, objName common.ObjectName, flowSub *FlowSubscription,
) error {
	abort := func(err error) error {
		logging.Logger(ctx).WarnContext(ctx,
			"flow subscription delete failed mid-teardown; aborting before remaining flows are touched",
			"failedObject", objName,
			"error", err,
		)

		return fmt.Errorf("failed to delete flow subscription for object '%s': %w", objName, err)
	}

	if err := c.deleteFlow(ctx, flowSub); err != nil {
		return abort(err)
	}

	omZipData, err := metadata.ConstructDestructiveOutboundMessage(
		string(flowSub.ObjectName), flowSub.OutboundMessage.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to construct destructive outbound message zip for %s: %w",
			flowSub.OutboundMessage.Name, err)
	}

	omEntity := "outbound message " + flowSub.OutboundMessage.Name

	omDeploy, err := c.deployDestructiveApex(ctx, omZipData, metadata.TestLevelNoTestRun)
	if err != nil {
		return abort(fmt.Errorf("failed to deploy destructive change for %s: %w", omEntity, err))
	}

	if !omDeploy.Success {
		return abort(fmt.Errorf("%w for %s: %s",
			errFlowTeardownDeployFailed, omEntity, formatDeployFailureDetails(omDeploy)))
	}

	return nil
}

// deleteFlow deactivates and removes every version of the flow.
//
// An unversioned Metadata destructive member (AmpSubscribe_Account) fails
// with "insufficient access rights on cross-reference id" when the definition
// has more than one version — Salesforce known issue:
// https://help.salesforce.com/s/issue?id=a028c00000gAwixAAC
//
// Teardown therefore deletes each Tooling Flow version by Id (the supported
// path for inactive versions), then falls back to a version-suffixed
// destructive deploy if any Tooling delete fails.
func (c *Connector) deleteFlow(ctx context.Context, flowSub *FlowSubscription) error {
	definitionID, err := c.deactivateFlow(ctx, flowSub)
	if err != nil {
		return err
	}

	if definitionID == "" { // flow is already gone
		return nil
	}

	versions, err := c.listFlowVersions(ctx, definitionID)
	if err != nil {
		return err
	}

	var remaining []int

	for _, version := range versions {
		_, delErr := c.deleteToSFAPI(ctx,
			"tooling/sobjects/Flow/"+version.ID,
			fmt.Sprintf("flow %s version %d", flowSub.Flow.Name, version.VersionNumber),
		)
		if delErr != nil {
			logging.Logger(ctx).WarnContext(ctx,
				"tooling delete of flow version failed; will try versioned metadata delete",
				"flow", flowSub.Flow.Name,
				"version", version.VersionNumber,
				"error", delErr,
			)

			remaining = append(remaining, version.VersionNumber)
		}
	}

	if len(remaining) == 0 {
		return nil
	}

	zipData, err := metadata.ConstructDestructiveFlow(flowSub.Flow.Name, remaining)
	if err != nil {
		return fmt.Errorf("failed to construct versioned destructive flow zip for %s: %w",
			flowSub.Flow.Name, err)
	}

	entity := "flow versions of " + flowSub.Flow.Name

	deployResult, err := c.deployDestructiveApex(ctx, zipData, metadata.TestLevelNoTestRun)
	if err != nil {
		return fmt.Errorf("failed to deploy destructive change for %s: %w", entity, err)
	}

	if !deployResult.Success {
		return fmt.Errorf("%w for %s: %s",
			errFlowTeardownDeployFailed, entity, formatDeployFailureDetails(deployResult))
	}

	return nil
}

// deactivateFlow makes the flow deletable. Returns the FlowDefinition Id, or
// empty if the flow is already gone.
//
// Primary route: Tooling API PATCH on FlowDefinition setting
// Metadata.activeVersionNumber to 0, which deactivates the flow (no version
// is active). Salesforce documents that FlowDefinition.Metadata activates and
// deactivates flows; 0 as the deactivate value is the established convention:
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_flowdefinition.htm
// https://salesforce.stackexchange.com/questions/396198/deactivating-a-salesforce-flow-via-the-metadata-api
// Fallback: redeploy the flow as Draft via the Metadata API, regenerated from
// the stored FlowSubscription components.
func (c *Connector) deactivateFlow(ctx context.Context, flowSub *FlowSubscription) (string, error) {
	definitionID, err := c.findToolingEntityIDByDeveloperName(ctx, "FlowDefinition", flowSub.Flow.Name)
	if err != nil {
		if errors.Is(err, errToolingEntityNotFound) {
			logging.Logger(ctx).InfoContext(ctx, "flow already absent, skipping deactivation and deletion",
				"flow", flowSub.Flow.Name,
			)

			return "", nil
		}

		return "", fmt.Errorf("failed to look up FlowDefinition for %s: %w", flowSub.Flow.Name, err)
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
		return definitionID, nil
	}

	logging.Logger(ctx).WarnContext(ctx,
		"tooling API flow deactivation failed; falling back to redeploying the flow as Draft",
		"flow", flowSub.Flow.Name,
		"error", patchErr,
	)

	// Fallback: redeploy the flow as Draft via the Metadata API
	params := metadata.FlowParams{
		ObjectName:          string(flowSub.ObjectName),
		FlowName:            flowSub.Flow.Name,
		RecordTriggerType:   metadata.RecordTriggerType(flowSub.Flow.RecordTriggerType),
		WatchFields:         flowSub.Flow.WatchFields,
		OutboundMessageName: flowSub.OutboundMessage.Name,
	}

	zipData, err := metadata.ConstructDraftFlow(params)
	if err != nil {
		return "", fmt.Errorf("failed to construct flow deactivation zip for %s: %w", flowSub.Flow.Name, err)
	}

	entity := "flow deactivation " + flowSub.Flow.Name

	deployResult, err := c.deployDestructiveApex(ctx, zipData, metadata.TestLevelNoTestRun)
	if err != nil {
		return "", fmt.Errorf("failed to deploy destructive change for %s: %w", entity, err)
	}

	if !deployResult.Success {
		return "", fmt.Errorf("%w for %s: %s",
			errFlowTeardownDeployFailed, entity, formatDeployFailureDetails(deployResult))
	}

	return definitionID, nil
}

// flowVersion is one Tooling API Flow row (a single version of a definition).
type flowVersion struct {
	ID            string `json:"Id"`
	Status        string `json:"Status"`
	VersionNumber int    `json:"VersionNumber"`
}

// listFlowVersions returns every version of a FlowDefinition. Used so teardown
// can delete AmpSubscribe_<Object>-N rather than the unversioned API name.
func (c *Connector) listFlowVersions(ctx context.Context, definitionID string) ([]flowVersion, error) {
	location, err := c.getRestApiURL("tooling/query")
	if err != nil {
		return nil, err
	}

	soql := fmt.Sprintf(
		"SELECT Id, Status, VersionNumber FROM Flow WHERE DefinitionId = '%s'",
		escapeSOQLString(definitionID),
	)
	location.WithQueryParam("q", soql)

	resp, err := c.Client.Get(ctx, location.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list flow versions for definition %s: %w", definitionID, err)
	}

	result, err := common.UnmarshalJSON[struct {
		Records []flowVersion `json:"records"`
	}](resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse flow version query for definition %s: %w", definitionID, err)
	}

	if result == nil {
		return nil, nil
	}

	return result.Records, nil
}
