package calendar

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/google/uuid"
)

// Compile-time interface conformance checks.
var (
	_ connectors.SubscribeConnector              = &Adapter{}
	_ connectors.SubscriptionMaintainerConnector = &Adapter{}
)

// objectWatchPaths maps supported subscribe objects to their watch URL paths.
//
//nolint:gochecknoglobals
var objectWatchPaths = map[common.ObjectName]string{
	objectNameEvents:       "calendars/primary/events/watch",
	objectNameCalendarList: "users/me/calendarList/watch",
	objectNameSettings:     "users/me/settings/watch",
	objectNameACL:          "calendars/primary/acl/watch",
}

// WatchRequest is the caller-provided config for creating watch channels.
// One WatchRequest is used for all subscribed objects; the adapter derives
// per-object channel IDs by appending the object name to ID.
// If ID is empty, the adapter generates a UUID base per Subscribe call.
type WatchRequest struct {
	// ID is the base channel identifier. The adapter appends "-{objectName}" per object.
	// If empty, a UUID is generated automatically.
	ID string `json:"id,omitempty"`

	// Address is the HTTPS URL that receives push notifications.
	Address string `json:"address"`

	// Token is an arbitrary string delivered in the X-Goog-Channel-Token header.
	// Optional; useful for verifying that a notification came from Google.
	Token string `json:"token,omitempty"`

	// Expiration is the requested channel lifetime in epoch milliseconds.
	// 0 means use the provider's default (up to ~30 days).
	Expiration int64 `json:"expiration,omitempty"`
}

// watchPayload is the request body sent to the Google Calendar watch endpoint.
type watchPayload struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Address    string `json:"address"`
	Token      string `json:"token,omitempty"`
	Expiration int64  `json:"expiration,omitempty"`
}

// WatchResponse is the response from the Calendar watch endpoint.
// Both ID and ResourceID are required to stop a channel later.
type WatchResponse struct {
	Kind        string `json:"kind"`
	ID          string `json:"id"`
	ResourceID  string `json:"resourceId"`
	ResourceURI string `json:"resourceUri"`
	Expiration  string `json:"expiration"` // epoch millis as string
}

// CalendarSubscriptionResult stores one WatchResponse per subscribed object.
// Persisted as SubscriptionResult.Result.
type CalendarSubscriptionResult struct {
	Channels map[common.ObjectName]*WatchResponse `json:"channels"`
}

// stopPayload is the request body for the channels/stop endpoint.
type stopPayload struct {
	ID         string `json:"id"`
	ResourceID string `json:"resourceId"`
}

// EmptySubscriptionParams returns a zero-value SubscribeParams with a typed WatchRequest.
func (a *Adapter) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{
		Request: &WatchRequest{},
	}
}

// EmptySubscriptionResult returns a zero-value SubscriptionResult with a typed CalendarSubscriptionResult.
func (a *Adapter) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &CalendarSubscriptionResult{
			Channels: make(map[common.ObjectName]*WatchResponse),
		},
	}
}

// Subscribe creates a watch channel for each requested object.
//
// Google Calendar requires a separate watch call per object — there is no
// multi-resource batch watch. For each object in params.SubscriptionEvents,
// this method POSTs to the object's watch endpoint and stores the resulting
// WatchResponse (which contains the channel ID and resourceId needed to stop it).
//
// # Event type filtering
//
// The ObjectEvents.Events field (SubscriptionEventTypeCreate, SubscriptionEventTypeUpdate,
// SubscriptionEventTypeDelete) is accepted for interface compliance but is NOT forwarded
// to the Google Calendar API. The Watch API does not support filtering by event type —
// every channel receives notifications for all changes (creates, updates, and deletes).
//
// Consumers must distinguish event types themselves using the X-Goog-Resource-State
// header delivered with each push notification:
//
//	"sync"       — initial handshake when the channel is created; ignore it
//	"exists"     — a resource was created or updated
//	"not_exists" — a resource was deleted
//
// If any watch call fails after some have succeeded, the successfully-created
// channels are rolled back via stopChannel before returning. If rollback also
// fails, returns SubscriptionStatusFailedToRollback so the caller knows orphaned
// channels may exist in Google Calendar.
//
// ref: https://developers.google.com/workspace/calendar/api/v3/reference/events/watch
// ref: https://developers.google.com/workspace/calendar/api/v3/reference/calendarList/watch
func (a *Adapter) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	watchReq, err := validateWatchRequest(params)
	if err != nil {
		return nil, err
	}

	baseID := watchReq.ID
	if baseID == "" {
		baseID = uuid.New().String()
	}

	result := &CalendarSubscriptionResult{
		Channels: make(map[common.ObjectName]*WatchResponse),
	}

	// Sort for deterministic processing order so rollback behaviour is predictable.
	objects := sortedKeys(params.SubscriptionEvents)

	for _, obj := range objects {
		resp, err := a.watchObject(ctx, obj, baseID, watchReq)
		if err != nil {
			if len(result.Channels) > 0 {
				rollbackErr := a.stopAllChannels(ctx, result.Channels)
				if rollbackErr != nil {
					return &common.SubscriptionResult{
						Status: common.SubscriptionStatusFailedToRollback,
						Result: result,
					}, fmt.Errorf("subscribe: watching %q failed: %w; rollback also failed: %w", obj, err, rollbackErr)
				}
			}

			return &common.SubscriptionResult{
				Status: common.SubscriptionStatusFailed,
			}, fmt.Errorf("subscribe: watching %q: %w", obj, err)
		}

		result.Channels[obj] = resp
	}

	return &common.SubscriptionResult{
		Result:       result,
		Status:       common.SubscriptionStatusSuccess,
		ObjectEvents: params.SubscriptionEvents,
	}, nil
}

// UpdateSubscription reconciles an existing subscription with the new desired state.
//
// Objects present in prev but absent from params are stopped. Objects in params
// but absent from prev are newly watched. Objects present in both are left untouched
// (Google Calendar does not support mutating an existing channel).
func (a *Adapter) UpdateSubscription( //nolint: cyclop,funlen
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	watchReq, err := validateWatchRequest(params)
	if err != nil {
		return nil, err
	}

	prevChannels := extractChannels(previousResult)

	toAdd := make(map[common.ObjectName]common.ObjectEvents)
	toRemove := make(map[common.ObjectName]*WatchResponse)

	for obj, evt := range params.SubscriptionEvents {
		if _, exists := prevChannels[obj]; !exists {
			toAdd[obj] = evt
		}
	}

	for obj, ch := range prevChannels {
		if _, exists := params.SubscriptionEvents[obj]; !exists {
			toRemove[obj] = ch
		}
	}

	// Start from a copy of the previous channels; we'll mutate it below.
	updatedChannels := make(map[common.ObjectName]*WatchResponse, len(prevChannels))
	maps.Copy(updatedChannels, prevChannels)

	// Stop removed channels.
	var stopErr error

	for obj, ch := range toRemove {
		if err := a.stopChannel(ctx, ch.ID, ch.ResourceID); err != nil {
			stopErr = errors.Join(stopErr, fmt.Errorf("update: stopping %q: %w", obj, err))
		} else {
			delete(updatedChannels, obj)
		}
	}

	if stopErr != nil {
		return &common.SubscriptionResult{
			Status: common.SubscriptionStatusFailed,
			Result: &CalendarSubscriptionResult{Channels: updatedChannels},
		}, stopErr
	}

	// Watch newly added objects.
	baseID := watchReq.ID
	if baseID == "" {
		baseID = uuid.New().String()
	}

	newChannels := make(map[common.ObjectName]*WatchResponse)

	for _, obj := range sortedKeys(toAdd) {
		resp, watchErr := a.watchObject(ctx, obj, baseID, watchReq)
		if watchErr != nil {
			if len(newChannels) > 0 {
				rollbackErr := a.stopAllChannels(ctx, newChannels)
				if rollbackErr != nil {
					return &common.SubscriptionResult{
						Status: common.SubscriptionStatusFailedToRollback,
						Result: &CalendarSubscriptionResult{Channels: updatedChannels},
					}, fmt.Errorf("update: watching %q failed: %w; rollback also failed: %w", obj, watchErr, rollbackErr)
				}

				// Rollback succeeded — prune the stopped channels from the result
				// so it accurately reflects what is still active in Google Calendar.
				for o := range newChannels {
					delete(updatedChannels, o)
				}
			}

			return &common.SubscriptionResult{
				Status: common.SubscriptionStatusFailed,
				Result: &CalendarSubscriptionResult{Channels: updatedChannels},
			}, fmt.Errorf("update: watching %q: %w", obj, watchErr)
		}

		newChannels[obj] = resp
		updatedChannels[obj] = resp
	}

	return &common.SubscriptionResult{
		Result:       &CalendarSubscriptionResult{Channels: updatedChannels},
		Status:       common.SubscriptionStatusSuccess,
		ObjectEvents: params.SubscriptionEvents,
	}, nil
}

// DeleteSubscription stops all active watch channels stored in result.
// Errors across individual channel stops are joined and returned together.
//
// ref: https://developers.google.com/workspace/calendar/api/v3/reference/channels/stop
func (a *Adapter) DeleteSubscription(
	ctx context.Context,
	result common.SubscriptionResult,
) error {
	channels := extractChannels(&result)

	return a.stopAllChannels(ctx, channels)
}

// RunScheduledMaintenance renews all watch channels.
//
// Google Calendar does not support extending an existing channel's lifetime.
// Renewal requires stopping each channel and re-watching the same objects.
// This is equivalent to DeleteSubscription followed by Subscribe with the same params.
//
// Callers should schedule maintenance before the earliest expiration across all channels.
func (a *Adapter) RunScheduledMaintenance(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	if err := a.DeleteSubscription(ctx, *previousResult); err != nil {
		return &common.SubscriptionResult{
			Status: common.SubscriptionStatusFailed,
		}, fmt.Errorf("maintenance: stopping old channels: %w", err)
	}

	return a.Subscribe(ctx, params)
}

// VerifyWebhookMessage always returns true for Google Calendar.
// Push notifications are delivered via Google's infrastructure and authenticated
// at the transport layer; no application-level signature verification is needed.
// Callers should verify the X-Goog-Channel-Token header against their stored token
// if they set one in WatchRequest.Token.
func (a *Adapter) VerifyWebhookMessage(
	ctx context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	return true, nil
}

// GetRecordsByIds is not implemented for Google Calendar.
// Calendar push notifications deliver an empty body — only headers are sent (X-Goog-Resource-State,
// X-Goog-Resource-ID, etc.). There is no record payload to enrich from; callers must issue
// a follow-up API read (e.g. events.list) to fetch what actually changed.
func (a *Adapter) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	recordIds []string, //nolint:revive
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	return nil, common.ErrGetRecordNotSupportedForObject
}

// watchObject POSTs a watch request for a single object and returns the channel response.
// The caller's ObjectEvents.Events (create/update/delete) are not included in the payload —
// the Google Calendar Watch API has no event-type filter; all change types are always delivered.
func (a *Adapter) watchObject(
	ctx context.Context,
	objectName common.ObjectName,
	baseID string,
	req *WatchRequest,
) (*WatchResponse, error) {
	watchPath, supported := objectWatchPaths[objectName]
	if !supported {
		return nil, fmt.Errorf("%w: %q", errUnsupportedSubscribeObject, objectName)
	}

	watchURL, err := url.JoinPath(a.ModuleInfo().BaseURL, apiVersion, watchPath)
	if err != nil {
		return nil, fmt.Errorf("watchObject: building URL for %q: %w", objectName, err)
	}

	payload := &watchPayload{
		ID:         perObjectChannelID(baseID, objectName),
		Type:       "web_hook",
		Address:    req.Address,
		Token:      req.Token,
		Expiration: req.Expiration,
	}

	response, err := a.JSONHTTPClient().Post(ctx, watchURL, payload)
	if err != nil {
		return nil, fmt.Errorf("watchObject: POST for %q: %w", objectName, err)
	}

	result, err := common.UnmarshalJSON[WatchResponse](response)
	if err != nil {
		return nil, fmt.Errorf("watchObject: unmarshalling response for %q: %w", objectName, err)
	}

	return result, nil
}

// stopChannel calls the Calendar channels/stop endpoint to terminate a single watch channel.
func (a *Adapter) stopChannel(ctx context.Context, id, resourceID string) error {
	stopURL, err := urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, "channels/stop")
	if err != nil {
		return fmt.Errorf("stopChannel: building URL: %w", err)
	}

	response, err := a.JSONHTTPClient().Post(ctx, stopURL.String(), &stopPayload{
		ID:         id,
		ResourceID: resourceID,
	})
	if err != nil {
		return fmt.Errorf("stopChannel: POST: %w", err)
	}

	// channels/stop returns 204 No Content on success with no body.
	if response.Code != http.StatusNoContent && response.Code != http.StatusOK {
		return fmt.Errorf("stopChannel: unexpected status %d", response.Code) //nolint: err113
	}

	return nil
}

// stopAllChannels stops every channel in the map, joining errors across failures.
func (a *Adapter) stopAllChannels(ctx context.Context, channels map[common.ObjectName]*WatchResponse) error {
	var errs error

	for obj, ch := range channels {
		if ch == nil {
			continue
		}

		if err := a.stopChannel(ctx, ch.ID, ch.ResourceID); err != nil {
			errs = errors.Join(errs, fmt.Errorf("stopping channel for %q: %w", obj, err))
		}
	}

	return errs
}

// perObjectChannelID appends the object name to the base ID to produce a unique channel ID.
// Google Calendar requires each channel to have a globally unique ID.
func perObjectChannelID(baseID string, objectName common.ObjectName) string {
	return baseID + "-" + string(objectName)
}

// extractChannels safely extracts the Channels map from a SubscriptionResult.
// Returns an empty map if the result is nil or has no CalendarSubscriptionResult.
func extractChannels(result *common.SubscriptionResult) map[common.ObjectName]*WatchResponse {
	if result == nil || result.Result == nil {
		return make(map[common.ObjectName]*WatchResponse)
	}

	sub, ok := result.Result.(*CalendarSubscriptionResult)
	if !ok || sub == nil {
		return make(map[common.ObjectName]*WatchResponse)
	}

	return sub.Channels
}

// validateWatchRequest type-asserts and validates the WatchRequest from params.
func validateWatchRequest(params common.SubscribeParams) (*WatchRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(*WatchRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected '*WatchRequest', got '%T'", errInvalidRequestType, params.Request)
	}

	if req.Address == "" {
		return nil, fmt.Errorf("%w: WatchRequest.Address is required", errMissingParams)
	}

	return req, nil
}

// sortedKeys returns the keys of m in ascending order so callers iterate deterministically.
func sortedKeys(m map[common.ObjectName]common.ObjectEvents) []common.ObjectName {
	keys := make([]common.ObjectName, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	return keys
}
