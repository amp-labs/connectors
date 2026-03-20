package mail

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/go-playground/validator"
)

var (
	errMissingParams      = errors.New("missing required parameters")
	errInvalidRequestType = errors.New("invalid request type")
)

const twoDaysHr = 48

// WatchRequest represents the Subscription.Request data expected from the builder.
type WatchRequest struct {
	// LabelIDs is a list of labelIds to restrict notifications about.
	// By default, if unspecified, all changes are pushed out.
	// If specified then dictates which labels are required for a push notification to be generated.
	LabelIDs []string `json:"labelIds" validate:"required"`

	// LabelFilterBehavior represents filtering behavior of labelIds list specified.
	LabelFilterBehavior string `json:"labelFilterBehavior" validate:"required"`

	// TopicName represents a fully qualified Google Cloud Pub/Sub API topic name to publish the events to
	TopicName string `json:"topicName" validate:"required"`
}

type watchResponse struct {
	// HistoryID is the ID of the mailbox's current history record.
	HistoryID string `json:"historyId"`

	// When Gmail will stop sending notifications for mailbox updates (epoch millis).
	// Call watch again before this time to renew the watch.
	Expiration string `json:"expiration"`
}

func validateRequest(params common.SubscribeParams) (*WatchRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(*WatchRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected '%T', got '%T'", errInvalidRequestType, req, params.Request)
	}

	validate := validator.New()

	if validate.Struct(req) != nil {
		return nil, fmt.Errorf("%w: request is invalid", errInvalidRequestType)
	}

	return req, nil
}

// Subscribe subscribes to the mailboxes events for the given params.
// It returns subscriptions expiry timestamp with the history id.
func (a *Adapter) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	watchURL, err := url.JoinPath(a.ModuleInfo().BaseURL, apiVersion, "users/me/watch")
	if err != nil {
		return nil, fmt.Errorf("subscribe: building watch URL: %w", err)
	}

	watchReq, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	response, err := a.JSONHTTPClient().Post(ctx, watchURL, watchReq)
	if err != nil {
		return nil, fmt.Errorf("subscribe: posting to gmail watch: %w", err)
	}

	result, err := common.UnmarshalJSON[watchResponse](response)
	if err != nil {
		return nil, err
	}

	return &common.SubscriptionResult{
		Result: result,
	}, nil
}

// RunScheduledMaintenance runs the schedule for the connector to maintain the subscription.
// gmail expects the same watch call that was used subscribing.
func (a *Adapter) RunScheduledMaintenance(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	response, ok := previousResult.Result.(watchResponse)
	if !ok {
		return nil, fmt.Errorf("%w: expected watchResponse, got %T", errInvalidRequestType, previousResult.Result)
	}

	expiration := response.Expiration

	// We don't to make this call necessarily,
	// if the subscription is still active, and has more than 2 days to go, we skip.
	ms, err := strconv.ParseInt(expiration, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("RunScheduledMaintenance: parsing expiration %q: %w", expiration, err)
	}

	exp := time.UnixMilli(ms)
	now := time.Now()

	inTwoDays := now.Add(twoDaysHr * time.Hour)

	// Renew if already expired, or expiring within the next 2 days
	if exp.Before(inTwoDays) {
		return a.Subscribe(ctx, params)
	}

	return previousResult, nil
}
