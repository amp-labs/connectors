package mail

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/tools/debug"
)

var (
	errMissingParams              = errors.New("missing required parameters")
	errInvalidRequestType         = errors.New("invalid request type")
	errUnsupportedSubscribeObject = errors.New("unsupported subscribe object")
)

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

func (a *Adapter) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{
		Request: &WatchRequest{},
	}
}

func (a *Adapter) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &WatchResponse{},
	}
}

// Subscribe creates a Gmail watch subscription for the requested objects.
// It maps each object to a Gmail label ID (e.g. "messages" → "INBOX", "drafts" → "DRAFT"),
// then issues a single watch API call with all labels combined. Only "messages" and "drafts"
// are supported; any other object is rejected.
func (a *Adapter) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	labels, err := buildLabelIDs(params.SubscriptionEvents)
	if err != nil {
		return nil, err
	}

	watchReq, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	watchReq.LabelIDs = labels

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

	result, unmarshalErr := common.UnmarshalJSON[WatchResponse](response)
	if unmarshalErr != nil {
		// The watch call succeeded at the provider (2xx) but we can't parse the response.
		// Attempt to roll back by stopping the watch so we don't leave an orphaned subscription.
		if response.Code >= http.StatusOK && response.Code < http.StatusMultipleChoices {
			if rollbackErr := a.stopWatch(ctx); rollbackErr != nil {
				return &common.SubscriptionResult{
					Status: common.SubscriptionStatusFailedToRollback,
				}, fmt.Errorf("subscribe: unmarshal failed: %w; rollback also failed: %w", unmarshalErr, rollbackErr)
			}
		}

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

// stopWatch calls the Gmail users.stop API to stop push notifications.
func (a *Adapter) stopWatch(ctx context.Context) error {
	watchURL, err := url.JoinPath(a.ModuleInfo().BaseURL, apiVersion, "users/me/stop")
	if err != nil {
		return fmt.Errorf("stop watch: building URL: %w", err)
	}

	// The request body must be empty per the API spec.
	response, err := a.JSONHTTPClient().Post(ctx, watchURL, nil)
	if err != nil {
		return fmt.Errorf("stop watch: posting to gmail stop: %w", err)
	}

	_, err = common.UnmarshalJSON[WatchResponse](response)
	if err != nil {
		return err
	}

	return nil
}

func (a *Adapter) GetRecordsByIds(ctx context.Context, // nolint: revive
	objectName string,
	recordIds []string, //nolint:revive
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	return nil, common.ErrGetRecordNotSupportedForObject
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

func (a *Adapter) VerifyWebhookMessage(
	ctx context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	//nolint:forbidigo
	fmt.Println("VerifyWebhookMessage-----------------",
		"request-----------------",
		debug.PrettyFormatStringJSON(request),
		"params-----------------",
		debug.PrettyFormatStringJSON(params))

	return true, nil
}
