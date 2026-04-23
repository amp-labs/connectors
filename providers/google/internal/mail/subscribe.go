package mail

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Errors for subscribe validation failures.
var (
	errMissingParams              = errors.New("missing required parameters")
	errInvalidRequestType         = errors.New("invalid request type")
	errUnsupportedSubscribeObject = errors.New("unsupported subscribe object")
)

// Supported subscribe object names. These correspond to object names
// that can be subscribed to via the Gmail API. Each maps to a Gmail label ID
// used to filter which mailbox changes trigger push notifications.
const (
	messagesObject common.ObjectName = "messages"
	draftsObject   common.ObjectName = "drafts"
)

// objectLabelIDs maps supported subscribe objects to their Gmail label IDs.
// The Gmail watch API uses label IDs to filter which mailbox changes trigger
// push notifications. Each object corresponds to a specific label.
//
//nolint:gochecknoglobals
var objectLabelIDs = map[common.ObjectName]string{
	messagesObject: "INBOX",
	draftsObject:   "DRAFT",
}

// Compile-time interface conformance checks.
var (
	_ connectors.SubscribeConnector              = &Adapter{}
	_ connectors.SubscriptionMaintainerConnector = &Adapter{}
)

// WatchRequest represents the Gmail watch API request body.
// A single watch call covers all subscribed objects; the LabelIDs field is
// populated by mapping each requested object to its corresponding Gmail label.
type WatchRequest struct {
	// LabelIDs restricts which mailbox changes trigger push notifications.
	// Built from the subscribed objects (e.g. "messages" → "INBOX", "drafts" → "DRAFT").
	LabelIDs []string `json:"labelIds"`

	// LabelFilterBehavior controls how LabelIDs are applied (include vs. exclude).
	LabelFilterBehavior string `json:"labelFilterBehavior"`

	// TopicName is the fully qualified Google Cloud Pub/Sub topic to publish events to.
	TopicName string `json:"topicName"`
}

// WatchResponse represents the response from the Gmail watch API.
type WatchResponse struct {
	// HistoryID is the ID of the mailbox's current history record.
	HistoryID string `json:"historyId"`

	// Expiration is when Gmail will stop sending notifications (epoch millis).
	// Call watch again before this time to renew the subscription.
	Expiration string `json:"expiration"`
}

// validateRequest extracts and type-asserts the WatchRequest from subscribe params.
func validateRequest(params common.SubscribeParams) (*WatchRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(*WatchRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected '*WatchRequest', got '%T'", errInvalidRequestType, params.Request)
	}

	return req, nil
}

// buildLabelIDs iterates over the requested subscription objects and returns
// the corresponding Gmail label IDs. Returns an error if any object is unsupported.
func buildLabelIDs(events map[common.ObjectName]common.ObjectEvents) ([]string, error) {
	labels := make([]string, 0, len(events))

	for obj := range events {
		label, ok := objectLabelIDs[obj]
		if !ok {
			return nil, fmt.Errorf("%w: %q", errUnsupportedSubscribeObject, obj)
		}

		labels = append(labels, label)
	}

	return labels, nil
}

// EmptySubscriptionParams returns a zero-value SubscribeParams with a typed WatchRequest.
// Used by the server for JSON deserialization of persisted subscription params.
func (a *Adapter) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{
		Request: &WatchRequest{},
	}
}

// EmptySubscriptionResult returns a zero-value SubscriptionResult with a typed WatchResponse.
// Used by the server for JSON deserialization of persisted subscription results.
func (a *Adapter) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &WatchResponse{},
	}
}

// Subscribe creates a Gmail watch subscription for the requested objects.
//
// The Gmail watch API (users.watch) accepts a single call with a set of label IDs
// that filter which mailbox changes produce push notifications. This method:
//
//  1. Validates that all requested objects are supported ("messages", "drafts").
//  2. Maps each object to its Gmail label ID ("messages" → "INBOX", "drafts" → "DRAFT").
//  3. Extracts the WatchRequest from params, applying default label IDs and filter
//     behavior ("include") if the caller did not provide explicit overrides.
//  4. Issues a single POST to the Gmail watch endpoint.
//  5. On success, returns SubscriptionStatusSuccess with the WatchResponse (containing
//     the historyId and expiration) and the subscribed ObjectEvents.
//
// Error handling:
//   - If the POST fails (network error or non-2xx), returns SubscriptionStatusFailed.
//   - If the POST succeeds (2xx) but the response cannot be unmarshalled, the watch
//     was created at the provider but we have no usable result. In this case, we attempt
//     to roll back via stopWatch (users.stop). If rollback also fails, returns
//     SubscriptionStatusFailedToRollback so the caller knows an orphaned watch may exist.
//
// ref: https://developers.google.com/gmail/api/reference/rest/v1/users/watch
//
//nolint:cyclop
func (a *Adapter) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	// Step 1: Validate objects and resolve their Gmail label IDs.
	labels, err := buildLabelIDs(params.SubscriptionEvents)
	if err != nil {
		return nil, err
	}

	// Step 2: Extract and validate the WatchRequest from params.
	watchReq, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	// Step 3: Apply defaults. If the caller provided explicit LabelIDs or
	// LabelFilterBehavior in the request, those take precedence over the
	// object-derived defaults. This allows builders to customize filtering
	// beyond the standard object-to-label mapping.
	if len(watchReq.LabelIDs) == 0 {
		watchReq.LabelIDs = labels
	}

	if watchReq.LabelFilterBehavior == "" {
		watchReq.LabelFilterBehavior = "include"
	}

	// Step 4: Build the watch URL and call the Gmail API.
	watchURL, err := url.JoinPath(a.ModuleInfo().BaseURL, apiVersion, "users/me/watch")
	if err != nil {
		return nil, fmt.Errorf("subscribe: building watch URL: %w", err)
	}

	response, err := a.JSONHTTPClient().Post(ctx, watchURL, watchReq)
	if err != nil {
		return &common.SubscriptionResult{
			Status: common.SubscriptionStatusFailed,
		}, fmt.Errorf("subscribe: posting to gmail watch: %w", err)
	}

	// Step 5: Parse the response. If unmarshalling fails after a successful HTTP call,
	// the watch exists at the provider but we can't track it. Roll back to avoid orphans.
	result, unmarshalErr := common.UnmarshalJSON[WatchResponse](response)
	if unmarshalErr != nil {
		if response.Code >= http.StatusOK && response.Code < http.StatusMultipleChoices {
			// 2xx means the watch was created — attempt to stop it.
			if rollbackErr := a.stopWatch(ctx); rollbackErr != nil {
				return &common.SubscriptionResult{
					Status: common.SubscriptionStatusFailedToRollback,
				}, fmt.Errorf("subscribe: unmarshal failed: %w; rollback also failed: %w", unmarshalErr, rollbackErr)
			}
		}

		// Either non-2xx (watch was never created) or rollback succeeded.
		return &common.SubscriptionResult{
			Status: common.SubscriptionStatusFailed,
		}, fmt.Errorf("subscribe: unmarshal failed: %w", unmarshalErr)
	}

	return &common.SubscriptionResult{
		Result:       result,
		Status:       common.SubscriptionStatusSuccess,
		ObjectEvents: params.SubscriptionEvents,
	}, nil
}

// RunScheduledMaintenance renews the Gmail watch subscription.
// Gmail watch subscriptions expire after 7 days, so this re-issues the same watch call.
func (a *Adapter) RunScheduledMaintenance(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	return a.Subscribe(ctx, params)
}

// DeleteSubscription stops Gmail push notifications by calling the users.stop API.
// ref: https://developers.google.com/gmail/api/reference/rest/v1/users/stop
func (a *Adapter) DeleteSubscription(
	ctx context.Context,
	params common.SubscriptionResult,
) error {
	return a.stopWatch(ctx)
}

// stopWatch calls the Gmail users.stop API to terminate push notifications for
// the authenticated user's mailbox. This is used both by DeleteSubscription for
// normal unsubscribe flows and by Subscribe as a rollback when the watch call
// succeeds but the response cannot be parsed.
//
// The stop endpoint requires an empty request body and returns an empty response.
// ref: https://developers.google.com/gmail/api/reference/rest/v1/users/stop
func (a *Adapter) stopWatch(ctx context.Context) error {
	watchURL, err := url.JoinPath(a.ModuleInfo().BaseURL, apiVersion, "users/me/stop")
	if err != nil {
		return fmt.Errorf("stop watch: building URL: %w", err)
	}

	response, err := a.JSONHTTPClient().Post(ctx, watchURL, nil)
	if err != nil {
		return fmt.Errorf("stop watch: posting to gmail stop: %w", err)
	}

	// The stop endpoint returns an empty body on success. We still attempt to
	// unmarshal to let the HTTP client surface any non-2xx error responses.
	_, err = common.UnmarshalJSON[WatchResponse](response)
	if err != nil {
		return err
	}

	return nil
}

// GetRecordsByIds fetches full Gmail message payloads for the given IDs.
// Only the "messages" object is supported; other objects return ErrGetRecordNotSupportedForObject.
func (a *Adapter) GetRecordsByIds(ctx context.Context, // nolint: revive
	objectName string,
	recordIds []string, //nolint:revive
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	if objectName != objectNameMessages {
		return nil, common.ErrGetRecordNotSupportedForObject
	}

	if len(recordIds) == 0 {
		return []common.ReadResultRow{}, nil
	}

	messages, err := a.fetchMessagesByIDs(ctx, recordIds)
	if err != nil {
		return nil, fmt.Errorf("GetRecordsByIds: fetching messages: %w", err)
	}

	fieldSet := datautils.NewSetFromList(fields)
	rows := make([]common.ReadResultRow, 0, len(messages))

	for id, msg := range messages {
		row := common.ReadResultRow{
			Id:  id,
			Raw: msg,
		}

		if len(fieldSet) > 0 {
			row.Fields = readhelper.SelectFields(msg, fieldSet)
		} else {
			row.Fields = msg
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// UpdateSubscription re-issues the watch call with the updated params.
// Gmail does not support partial updates, so this is equivalent to a fresh subscribe.
func (a *Adapter) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	return a.Subscribe(ctx, params)
}

// HistoryListParams are inputs for Gmail users.history.list.
// Ref: https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.history/list
type HistoryListParams struct {
	// StartHistoryID is the checkpoint to fetch changes since. Required.
	StartHistoryID string

	// HistoryTypes restricts results to these record types
	// (messageAdded, messageDeleted, labelAdded, labelRemoved). Optional.
	HistoryTypes []string

	// LabelID restricts history records to those affecting the given label. Optional.
	LabelID string

	// MaxResults is the max number of history records per page. Optional (Gmail default: 100, max: 500).
	MaxResults int
}

// HistoryMessage is the minimal message shape embedded in history records.
type HistoryMessage struct {
	ID       string   `json:"id"`
	ThreadID string   `json:"threadId"`
	LabelIDs []string `json:"labelIds,omitempty"`
}

// HistoryMessageChange wraps a message reference for messagesAdded / messagesDeleted entries.
type HistoryMessageChange struct {
	Message HistoryMessage `json:"message"`
}

// HistoryLabelChange is a labelsAdded / labelsRemoved entry.
type HistoryLabelChange struct {
	Message  HistoryMessage `json:"message"`
	LabelIDs []string       `json:"labelIds,omitempty"`
}

// HistoryRecord is a single record from the history.list response.
type HistoryRecord struct {
	ID              string                 `json:"id"`
	Messages        []HistoryMessage       `json:"messages,omitempty"`
	MessagesAdded   []HistoryMessageChange `json:"messagesAdded,omitempty"`
	MessagesDeleted []HistoryMessageChange `json:"messagesDeleted,omitempty"`
	LabelsAdded     []HistoryLabelChange   `json:"labelsAdded,omitempty"`
	LabelsRemoved   []HistoryLabelChange   `json:"labelsRemoved,omitempty"`
}

// HistoryListResult is the aggregated result from paginating history.list.
type HistoryListResult struct {
	// HistoryID is the root historyId from the latest page — use as the next checkpoint.
	HistoryID string `json:"historyId"`

	// History is the concatenated records across all pages.
	History []HistoryRecord `json:"history"`
}

// historyListPage is a single page of the history.list response.
type historyListPage struct {
	History       []HistoryRecord `json:"history"`
	HistoryID     string          `json:"historyId"`
	NextPageToken string          `json:"nextPageToken"`
}

// buildHistoryListURL constructs the users.history URL for a single page fetch.
func (a *Adapter) buildHistoryListURL(params HistoryListParams, pageToken string) (string, error) {
	historyURL, err := urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, "users/me/history")
	if err != nil {
		return "", fmt.Errorf("history.list: building URL: %w", err)
	}

	historyURL.WithQueryParam("startHistoryId", params.StartHistoryID)

	if len(params.HistoryTypes) > 0 {
		historyURL.WithQueryParamList("historyTypes", params.HistoryTypes)
	}

	if params.LabelID != "" {
		historyURL.WithQueryParam("labelId", params.LabelID)
	}

	if params.MaxResults > 0 {
		historyURL.WithQueryParam("maxResults", strconv.Itoa(params.MaxResults))
	}

	if pageToken != "" {
		historyURL.WithQueryParam("pageToken", pageToken)
	}

	return historyURL.String(), nil
}

// HistoryList fetches Gmail mailbox history since params.StartHistoryID, paginating through all pages.
// Returns the aggregated records plus the root historyId — the new checkpoint to persist.
func (a *Adapter) HistoryList(
	ctx context.Context,
	params HistoryListParams,
) (*HistoryListResult, error) {
	if params.StartHistoryID == "" {
		return nil, fmt.Errorf("%w: startHistoryId is required", errMissingParams)
	}

	result := &HistoryListResult{}

	var pageToken string

	for {
		historyURL, err := a.buildHistoryListURL(params, pageToken)
		if err != nil {
			return nil, err
		}

		resp, err := a.JSONHTTPClient().Get(ctx, historyURL)
		if err != nil {
			return nil, fmt.Errorf("history.list: GET: %w", err)
		}

		page, err := common.UnmarshalJSON[historyListPage](resp)
		if err != nil {
			return nil, fmt.Errorf("history.list: unmarshaling response: %w", err)
		}

		result.History = append(result.History, page.History...)
		result.HistoryID = page.HistoryID

		if page.NextPageToken == "" {
			break
		}

		pageToken = page.NextPageToken
	}

	return result, nil
}

// VerifyWebhookMessage always returns true for Gmail. Gmail push notifications are
// delivered via Google Cloud Pub/Sub which handles authentication at the transport layer,
// so no application-level signature verification is needed.
func (a *Adapter) VerifyWebhookMessage(
	ctx context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	return true, nil
}
