package calendar

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"sync"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/google/uuid"
)

// calendarMaxConcurrentWatches bounds the number of watch/stop calls issued in
// parallel so we don't trip Google Calendar's per-user rate limits.
const calendarMaxConcurrentWatches = 4

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

	// watchObjects returns whatever channels were created even on failure, so we can
	// roll them back here rather than mid-loop.
	channels, err := a.watchObjects(ctx, sortedKeys(params.SubscriptionEvents), baseID, watchReq)
	if err != nil {
		if len(channels) > 0 {
			if rollbackErr := a.stopAllChannels(ctx, channels); rollbackErr != nil {
				return &common.SubscriptionResult{
					Status: common.SubscriptionStatusFailedToRollback,
					Result: &CalendarSubscriptionResult{Channels: channels},
				}, fmt.Errorf("subscribe: %w; rollback also failed: %w", err, rollbackErr)
			}
		}

		return &common.SubscriptionResult{
			Status: common.SubscriptionStatusFailed,
		}, fmt.Errorf("subscribe: %w", err)
	}

	return &common.SubscriptionResult{
		Result:       &CalendarSubscriptionResult{Channels: channels},
		Status:       common.SubscriptionStatusSuccess,
		ObjectEvents: params.SubscriptionEvents,
	}, nil
}

// UpdateSubscription moves an existing subscription to the new desired object set.
//
// Google Calendar channels cannot be extended or mutated, so an "update" is really
// a recreate: we create a fresh set of channels for the full desired object set and
// then stop the previous channels. Recreating everything — not just the delta — keeps
// all channels on the same expiration clock so they renew together.
//
// Creating the new channels happens first; if that fails nothing has changed and the
// error is returned. Stopping the previous channels is best-effort: they expire on
// their own, so a failure there is logged and tolerated (the caller may briefly receive
// duplicate notifications until the old channels lapse).
func (a *Adapter) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	result, err := a.Subscribe(ctx, params)
	if err != nil {
		return result, fmt.Errorf("update: creating new subscription: %w", err)
	}

	if previousResult != nil {
		if err := a.DeleteSubscription(ctx, *previousResult); err != nil {
			logging.Logger(ctx).Warn(
				"update: failed to stop previous channels; they will expire automatically",
				"error", err,
			)
		}
	}

	return result, nil
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
// Google Calendar does not support extending an existing channel's lifetime, so renewal
// is the same recreate-then-stop flow as UpdateSubscription: create a fresh set of
// channels for the same objects, then stop the previous ones.
//
// Callers should schedule maintenance before the earliest expiration across all channels.
func (a *Adapter) RunScheduledMaintenance(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	return a.UpdateSubscription(ctx, params, previousResult)
}

// VerifyWebhookMessage is not yet implemented for Google Calendar.
//
// Verification (comparing the X-Goog-Channel-Token header against a stored token) is
// being delivered in a separate PR. Until then this refuses verification rather than
// asserting authenticity it never checked — callers must not treat unverified messages
// as trusted.
func (a *Adapter) VerifyWebhookMessage(
	ctx context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	return false, common.ErrNotImplemented
}

// GetRecordsByIds fetches the events that changed since a server-supplied checkpoint.
//
// # The recordIds convention (Calendar-specific)
//
// Calendar push notifications deliver an empty body — only headers are sent
// (X-Goog-Resource-State, X-Goog-Resource-ID, etc.) — so they carry no record IDs to
// enrich from. Rather than change the shared BatchRecordReaderConnector signature for a
// Calendar-only need, the subscribe pipeline reuses recordIds to deliver the fetch window:
//
//	recordIds[0] is NOT an event ID. It is an RFC3339 timestamp in UTC with milliseconds,
//	e.g. "2026-06-18T00:00:00.000Z" — the server's checkpoint of "changed since".
//
// It is used verbatim as the events.list updatedMin query param (it is already the exact
// format Google expects, so no parsing or reformatting is done here). Persisting and
// advancing the checkpoint, and de-duplicating across overlapping notifications, are the
// server's responsibility — this method is a stateless read of one window.
//
// # What is fetched
//
//   - Only the "events" object is supported (other objects return ErrGetRecordNotSupportedForObject).
//   - updatedMin filters by modification time, so edits to past or future events are all
//     captured (unlike timeMin/timeMax, which filter by scheduled time).
//   - showDeleted=true so deletions are returned as events with status:"cancelled".
//
// ref: https://developers.google.com/workspace/calendar/api/v3/reference/events/list
func (a *Adapter) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	recordIds []string, //nolint:revive
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	if objectName != objectNameEvents {
		return nil, common.ErrGetRecordNotSupportedForObject
	}

	if len(recordIds) == 0 || recordIds[0] == "" {
		return nil, fmt.Errorf("%w: recordIds[0] must be an updatedMin timestamp", errMissingParams)
	}

	updatedMin := recordIds[0]

	url, err := a.getURL(objectNameEvents)
	if err != nil {
		return nil, fmt.Errorf("GetRecordsByIds: building URL: %w", err)
	}

	url.WithQueryParam("maxResults", strconv.Itoa(defaultPageSize))
	url.WithQueryParam("showDeleted", "true")
	url.WithQueryParam("updatedMin", updatedMin)

	// Reuse the standard read parsing (record extraction, field marshaling, pagination).
	readParams := common.ReadParams{
		ObjectName: objectNameEvents,
		Fields:     datautils.NewSetFromList(fields),
	}

	rows := make([]common.ReadResultRow, 0)

	for pageURL := url.String(); pageURL != ""; {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
		if err != nil {
			return nil, fmt.Errorf("GetRecordsByIds: building request: %w", err)
		}

		response, err := a.JSONHTTPClient().Get(ctx, pageURL)
		if err != nil {
			return nil, fmt.Errorf("GetRecordsByIds: events.list: %w", err)
		}

		result, err := a.parseReadResponse(ctx, readParams, request, response)
		if err != nil {
			return nil, fmt.Errorf("GetRecordsByIds: parsing events: %w", err)
		}

		rows = append(rows, result.Data...)

		if result.Done {
			break
		}

		pageURL = result.NextPage.String()
	}

	return rows, nil
}

// watchResult pairs a created watch channel with the object it belongs to so the
// concurrent fan-in can rebuild the object→channel map.
// watchObjects creates a watch channel for each object concurrently, bounded by
// calendarMaxConcurrentWatches.
//
// Each job records its outcome under a mutex and returns nil regardless of failure, so
// every watch call runs to completion rather than cancelling its siblings on the first
// error. This mirrors the other subscribe connectors (outreach, salesloft, zoho) and
// guarantees the returned map holds *every* channel that was created — even when err is
// non-nil — so the caller has the full set to roll back. Per-object failures are joined
// into the returned error.
func (a *Adapter) watchObjects(
	ctx context.Context,
	objects []common.ObjectName,
	baseID string,
	req *WatchRequest,
) (map[common.ObjectName]*WatchResponse, error) {
	var (
		mutex    sync.Mutex
		watchErr error
	)

	channels := make(map[common.ObjectName]*WatchResponse, len(objects))
	callbacks := make([]simultaneously.Job, 0, len(objects))

	for _, obj := range objects {
		callbacks = append(callbacks, func(ctx context.Context) error {
			resp, err := a.watchObject(ctx, obj, baseID, req)

			mutex.Lock()
			defer mutex.Unlock()

			if err != nil {
				watchErr = errors.Join(watchErr, fmt.Errorf("watching %q: %w", obj, err))

				return nil
			}

			channels[obj] = resp

			return nil
		})
	}

	if err := simultaneously.DoCtx(ctx, calendarMaxConcurrentWatches, callbacks...); err != nil {
		return channels, err
	}

	return channels, watchErr
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
