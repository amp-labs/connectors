package salesforce

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"

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
	errFlowDeployFailed         = errors.New("flow subscription deployment failed")
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

	// SelectedFields is the installation's selected read-field list per object, which
	// are the fields that are delivered in the subscription payload.
	SelectedFields map[common.ObjectName][]string `json:"selectedFields,omitempty"`
}

// OutboundMessageMetadata is our record of a deployed OM, stored on a FlowPair
// (or, for the legacy shape, SubscribeResult.Flows[object].OutboundMessage). Delete uses Name for the
// destructive zip and as OutboundMessageName if the flow is Draft-redeployed.
// Flow updates delete-then-recreate, so they only read this on that delete.
// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_workflow.htm
type OutboundMessageMetadata struct {
	// Name is the outbound message developer name WITHOUT the object prefix;
	// the deployable full name is "<ObjectName>.<Name>".
	Name string `json:"name"`

	// Fields is the OM payload besides Id/CreatedDate/LastModifiedDate:
	// the installation's selected fields.
	Fields []string `json:"fields,omitempty"`

	// DeployID is the Metadata API deploy id that created the component.
	DeployID string `json:"deployId,omitempty"`
}

// FlowMetadata is our record of a deployed flow, stored on a FlowPair (or, for
// the legacy shape, SubscribeResult.Flows[object].Flow). Delete uses Name plus RecordTriggerType
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

	// WatchFields are baked into the flow's ISCHANGED entry-condition formula
	// (Update and CreateAndUpdate). Empty means no formula: any matching
	// trigger fires. Create-only ignores them.
	WatchFields []string `json:"watchFields,omitempty"`

	// DeployID is the Metadata API deploy id that created the component.
	DeployID string `json:"deployId,omitempty"`
}

// FlowPair is one deployed record-triggered flow and the outbound message it
// invokes. They are deployed and torn down together.
type FlowPair struct {
	// OutboundMessage is the OM deployed in the same package as Flow.
	OutboundMessage *OutboundMessageMetadata `json:"outboundMessage,omitempty"`

	// Flow is the record-triggered flow deployed in the same package as the OM.
	Flow *FlowMetadata `json:"flow,omitempty"`
}

// FlowSubscription is the persisted per-object record on SubscribeResult.Flows:
// the OM and flow bookmarks for one object, written after the combined deploy
// succeeds.
type FlowSubscription struct {
	ObjectName common.ObjectName `json:"objectName"`

	// Deprecated: OutboundMessage is the OM deployed in the same package as Flow (legacy shape).
	OutboundMessage *OutboundMessageMetadata `json:"outboundMessage,omitempty"`

	// Deprecated: Flow is the record-triggered flow deployed in the same package as the OM (legacy shape).
	Flow *FlowMetadata `json:"flow,omitempty"`

	// Pairs stores the flow + outbound message pair for each event type (create/updated) this config is subscribed to.
	Pairs map[common.SubscriptionEventType]*FlowPair `json:"pairs,omitempty"`

	// Events are the normalized event types this subscription covers
	// (create/update; never delete).
	Events []common.SubscriptionEventType `json:"events"`
}

// FlowPairs returns the flows + outbound message pairs recorded for the object.
func (f *FlowSubscription) FlowPairs() []*FlowPair {
	if f == nil {
		return nil
	}

	// If the subscription contains an OM/Flow in the legacy shape, return the single pair.
	if f.OutboundMessage != nil && f.Flow != nil {
		return []*FlowPair{{
			OutboundMessage: f.OutboundMessage,
			Flow:            f.Flow,
		}}
	}

	// Otherwise, look to see if the subscription contains a CreateEvent and/or UpdateEvent FlowPair.
	pairs := make([]*FlowPair, 0, len(f.Pairs))

	for _, eventType := range []common.SubscriptionEventType{
		common.SubscriptionEventTypeCreate,
		common.SubscriptionEventTypeUpdate,
	} {
		if pair := f.Pairs[eventType]; pair != nil && pair.Flow != nil && pair.OutboundMessage != nil {
			pairs = append(pairs, pair)
		}
	}

	return pairs
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
	prevState *SubscribeResult,
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

	flowComponents, subscriptions, err := buildFlowSubscription(params, req, prevState)
	if err != nil {
		return nil, err
	}

	zipData, err := metadata.ConstructFlowSubscription(flowComponents)
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

	// Set the deploy ID on each flow and outbound message pairs.
	for _, subscription := range subscriptions {
		for _, pair := range subscription.Pairs {
			pair.OutboundMessage.DeployID = deployID
			pair.Flow.DeployID = deployID
		}

		sfRes.Flows[subscription.ObjectName] = subscription
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

// buildFlowSubscription builds the flow and outbound-message pairs to deploy (one per event type)
// and the subscription records to persist once that deploy succeeds for each object in params.SubscriptionEvents.
//
//nolint:cyclop,funlen,gocognit
func buildFlowSubscription(
	params common.SubscribeParams,
	req *SubscriptionRequest,
	prevState *SubscribeResult,
) ([]metadata.FlowSubscriptionComponent, []*FlowSubscription, error) {
	flowComponents := make([]metadata.FlowSubscriptionComponent, 0, 2*len(params.SubscriptionEvents)) //nolint:mnd
	subscriptions := make([]*FlowSubscription, 0, len(params.SubscriptionEvents))

	for objName, objEvents := range params.SubscriptionEvents {
		// Outbound messages cannot fire on delete. Each requested create or update gets one pair.
		if slices.Contains(objEvents.Events, common.SubscriptionEventTypeDelete) {
			return nil, nil, fmt.Errorf(
				"failed to build flow subscription for object %s: %w: delete events are not supported",
				objName, errFlowNoSupportedEvents)
		}

		var eventTypes []common.SubscriptionEventType

		for _, eventType := range []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
			common.SubscriptionEventTypeUpdate,
		} {
			if slices.Contains(objEvents.Events, eventType) {
				eventTypes = append(eventTypes, eventType)
			}
		}

		if len(eventTypes) == 0 {
			return nil, nil, fmt.Errorf(
				"failed to build flow subscription for object %s: %w: requested %v",
				objName, errFlowNoSupportedEvents, objEvents.Events)
		}

		selectedFields := req.Flow.SelectedFields[objName]
		subscription := &FlowSubscription{
			ObjectName: objName,
			Events:     objEvents.Events,
			Pairs:      make(map[common.SubscriptionEventType]*FlowPair, len(eventTypes)),
		}

		for _, eventType := range eventTypes {
			triggerType, nameSuffix := metadata.RecordTriggerTypeUpdate, metadata.ArtifactSuffixUpdate
			if eventType == common.SubscriptionEventTypeCreate {
				triggerType, nameSuffix = metadata.RecordTriggerTypeCreate, metadata.ArtifactSuffixCreate
			}

			// A recorded name is reused so a prefix change upserts that flow and outbound message.
			var artifactName string
			if prevState != nil && prevState.Flows[objName] != nil {
				if pair := prevState.Flows[objName].Pairs[eventType]; pair != nil && pair.Flow != nil {
					artifactName = pair.Flow.Name
				}
			}

			if artifactName == "" {
				var err error
				artifactName, err = metadata.GenerateSubscriptionArtifactName(
					req.Flow.NamePrefix, string(objName), nameSuffix)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to build flow subscription for object %s: %w", objName, err)
				}
			}

			endpointURL, err := flowPairEndpointURL(req.Flow.EndpointURL, eventType)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to build flow subscription for object %s: %w", objName, err)
			}

			// Watch fields only gate updates: a create has no prior values to compare.
			// WatchFieldsAll means no ISCHANGED formula, so every update fires.
			var watchFields []string
			if eventType == common.SubscriptionEventTypeUpdate && !objEvents.WatchFieldsAll {
				watchFields = objEvents.WatchFields
			}

			// Deploy zip and saved record for this same pair.
			flowComponents = append(flowComponents, metadata.FlowSubscriptionComponent{
				OutboundMessage: metadata.OutboundMessageParams{
					ObjectName:          string(objName),
					Name:                artifactName,
					EndpointURL:         endpointURL,
					IntegrationUsername: req.Flow.IntegrationUsername,
					Fields:              selectedFields,
				},
				Flow: metadata.FlowParams{
					ObjectName:          string(objName),
					FlowName:            artifactName,
					OutboundMessageName: artifactName,
					RecordTriggerType:   triggerType,
					WatchFields:         watchFields,
				},
			})

			subscription.Pairs[eventType] = &FlowPair{
				OutboundMessage: &OutboundMessageMetadata{
					Name:   artifactName,
					Fields: selectedFields,
				},
				Flow: &FlowMetadata{
					Name:              artifactName,
					RecordTriggerType: string(triggerType),
					WatchFields:       watchFields,
				},
			}
		}

		subscriptions = append(subscriptions, subscription)
	}

	return flowComponents, subscriptions, nil
}

// flowPairEndpointURL returns the endpoint URL for one event type's outbound message.
func flowPairEndpointURL(endpointURL string, eventType common.SubscriptionEventType) (string, error) {
	parsed, err := url.Parse(endpointURL)
	if err != nil {
		return "", fmt.Errorf("invalid flow endpoint URL: %w", err)
	}

	query := parsed.Query()
	query.Set(omQueryParamEventType, string(eventType))
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
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

// reconcileFlowSubscription updates a flow-mode subscription in place.
//
// Salesforce replaces the outbound message in place, and for the flow it
// activates a new version while deactivating the previous one as part of the
// same deploy. Deactivated versions of the surviving flows are deleted after
// the upsert to avoid hitting Salesforce's 50-versions-per-flow cap.
//
// Objects that dropped out of the subscription also have their artifacts removed
// after the upsert.
func (c *Connector) reconcileFlowSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	req *SubscriptionRequest,
	prevState *SubscribeResult,
) (*common.SubscriptionResult, error) {
	// Upsert the new subscription.
	result, err := c.subscribeWithFlow(ctx, params, req, prevState)
	if err != nil {
		return result, err
	}

	sfRes, ok := result.Result.(*SubscribeResult)
	if !ok {
		return result, fmt.Errorf("%w: expected *SubscribeResult from the reconcile deploy, got '%T'",
			errInvalidRequestType, result.Result)
	}

	// Delete deactivated versions of flows that stayed subscribed (Active stays).
	c.deleteDeactivatedFlows(ctx, sfRes)

	// Delete the flows + outbound messages the new subscription no longer has.
	for objName, pairs := range droppedFlowPairs(prevState, sfRes) {
		for _, pair := range pairs {
			if err := c.deleteFlowPair(ctx, objName, pair); err != nil {
				return result, fmt.Errorf("reconcile deployed but could not remove a dropped flow pair: %w", err)
			}
		}
	}

	return result, nil
}

// droppedFlowPairs diffs the previous and new subscription states and returns the
// previous pairs whose flow the new state no longer has, keyed by object. This covers
// an object being dropped from the subscription and an event being dropped from an object.
func droppedFlowPairs(prevState, newState *SubscribeResult) map[common.ObjectName][]*FlowPair {
	if prevState == nil {
		return nil
	}

	dropped := make(map[common.ObjectName][]*FlowPair)

	for objName, prevSub := range prevState.Flows {
		kept := make(map[string]bool)

		if newState != nil {
			for _, pair := range newState.Flows[objName].FlowPairs() {
				kept[pair.Flow.Name] = true
			}
		}

		for _, pair := range prevSub.FlowPairs() {
			if !kept[pair.Flow.Name] {
				dropped[objName] = append(dropped[objName], pair)
			}
		}
	}

	return dropped
}

const flowVersionStatusActive = "Active"

// deleteDeactivatedFlows deletes inactive versions of every surviving flow
// after a reconcile upsert. The Active version is left in place. Failures
// are logged only: the new version is already active, so a missed delete
// must not fail the update.
func (c *Connector) deleteDeactivatedFlows(ctx context.Context, sfRes *SubscribeResult) {
	for _, flowSub := range sfRes.Flows {
		for _, pair := range flowSub.FlowPairs() {
			flowName := pair.Flow.Name

			definitionID, err := c.findToolingEntityIDByDeveloperName(ctx, "FlowDefinition", flowName)
			if err != nil {
				logging.Logger(ctx).WarnContext(ctx,
					"failed to delete deactivated flow versions during subscription update",
					"flow", flowName,
					"error", err,
				)

				continue
			}

			versions, err := c.listFlowVersions(ctx, definitionID)
			if err != nil {
				logging.Logger(ctx).WarnContext(ctx,
					"failed to delete deactivated flow versions during subscription update",
					"flow", flowName,
					"error", err,
				)

				continue
			}

			for _, version := range versions {
				if version.Status == flowVersionStatusActive {
					continue
				}

				_, delErr := c.deleteToSFAPI(ctx,
					"tooling/sobjects/Flow/"+version.ID,
					fmt.Sprintf("flow %s version %d", flowName, version.VersionNumber),
				)
				if delErr != nil {
					logging.Logger(ctx).WarnContext(ctx,
						"tooling delete of deactivated flow version failed; leaving it in place",
						"flow", flowName,
						"version", version.VersionNumber,
						"status", version.Status,
						"error", delErr,
					)
				}
			}
		}
	}
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
		for _, pair := range flowSub.FlowPairs() {
			if err := c.deleteFlowPair(ctx, objName, pair); err != nil {
				return err
			}
		}
	}

	return nil
}

// deleteFlowPair removes one flow pair dependents-first: the flow (deactivated,
// then every version) and then the outbound message it referenced.
func (c *Connector) deleteFlowPair(ctx context.Context, objName common.ObjectName, pair *FlowPair) error {
	abort := func(err error) error {
		logging.Logger(ctx).WarnContext(ctx,
			"flow subscription delete failed mid-teardown; aborting before remaining flows are touched",
			"failedObject", objName,
			"error", err,
		)

		return fmt.Errorf("failed to delete flow subscription for object '%s': %w", objName, err)
	}

	if err := c.deleteFlow(ctx, objName, pair); err != nil {
		return abort(err)
	}

	omZipData, err := metadata.ConstructDestructiveOutboundMessage(string(objName), pair.OutboundMessage.Name)
	if err != nil {
		return fmt.Errorf("failed to construct destructive outbound message zip for %s: %w",
			pair.OutboundMessage.Name, err)
	}

	omEntity := "outbound message " + pair.OutboundMessage.Name

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
func (c *Connector) deleteFlow(ctx context.Context, objName common.ObjectName, pair *FlowPair) error {
	definitionID, err := c.deactivateFlow(ctx, objName, pair)
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
			fmt.Sprintf("flow %s version %d", pair.Flow.Name, version.VersionNumber),
		)
		if delErr != nil {
			logging.Logger(ctx).WarnContext(ctx,
				"tooling delete of flow version failed; will try versioned metadata delete",
				"flow", pair.Flow.Name,
				"version", version.VersionNumber,
				"error", delErr,
			)

			remaining = append(remaining, version.VersionNumber)
		}
	}

	if len(remaining) == 0 {
		return nil
	}

	zipData, err := metadata.ConstructDestructiveFlow(pair.Flow.Name, remaining)
	if err != nil {
		return fmt.Errorf("failed to construct versioned destructive flow zip for %s: %w",
			pair.Flow.Name, err)
	}

	entity := "flow versions of " + pair.Flow.Name

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
func (c *Connector) deactivateFlow(ctx context.Context, objName common.ObjectName, pair *FlowPair) (string, error) {
	definitionID, err := c.findToolingEntityIDByDeveloperName(ctx, "FlowDefinition", pair.Flow.Name)
	if err != nil {
		if errors.Is(err, errToolingEntityNotFound) {
			logging.Logger(ctx).InfoContext(ctx, "flow already absent, skipping deactivation and deletion",
				"flow", pair.Flow.Name,
			)

			return "", nil
		}

		return "", fmt.Errorf("failed to look up FlowDefinition for %s: %w", pair.Flow.Name, err)
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
		"flow", pair.Flow.Name,
		"error", patchErr,
	)

	// Fallback: redeploy the flow as Draft via the Metadata API
	params := metadata.FlowParams{
		ObjectName:          string(objName),
		FlowName:            pair.Flow.Name,
		RecordTriggerType:   metadata.RecordTriggerType(pair.Flow.RecordTriggerType),
		WatchFields:         pair.Flow.WatchFields,
		OutboundMessageName: pair.OutboundMessage.Name,
	}

	zipData, err := metadata.ConstructDraftFlow(params)
	if err != nil {
		return "", fmt.Errorf("failed to construct flow deactivation zip for %s: %w", pair.Flow.Name, err)
	}

	entity := "flow deactivation " + pair.Flow.Name

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
