package mail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

var (
	errMissingParams      = errors.New("missing required parameters")
	errInvalidRequestType = errors.New("invalid request type")
)

var (
	_ connectors.SubscribeConnector              = &Adapter{}
	_ connectors.SubscriptionMaintainerConnector = &Adapter{}
)

// watchRequest represents the Subscription.Request data expected from the builder.
type watchRequest struct {
	// LabelIDs is a list of labelIds to restrict notifications about.
	// By default, if unspecified, all changes are pushed out.
	// If specified then dictates which labels are required for a push notification to be generated.
	LabelIDs []string `json:"labelIds"`

	// LabelFilterBehavior represents filtering behavior of labelIds list specified.
	LabelFilterBehavior string `json:"labelFilterBehavior"`

	// TopicName represents a fully qualified Google Cloud Pub/Sub API topic name to publish the events to
	TopicName string `json:"topicName"`
}

type watchResponse struct {
	// HistoryID is the ID of the mailbox's current history record.
	HistoryID string `json:"historyId"`

	// When Gmail will stop sending notifications for mailbox updates (epoch millis).
	// Call watch again before this time to renew the watch.
	Expiration string `json:"expiration"`
}

func validateRequest(params common.SubscribeParams) ([]byte, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected '%T', got '%T'", errInvalidRequestType, req, params.Request)
	}

	raw, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("subscribe: marshaling request: %w", err)
	}

	return raw, nil
}

func (a *Adapter) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{}
}

func (a *Adapter) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &watchResponse{},
	}
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

	req, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	var watchReq watchRequest
	if err := json.Unmarshal(req, &watchReq); err != nil {
		return nil, fmt.Errorf("subscribe: unmarshaling into watchRequest: %w", err)
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
	return a.Subscribe(ctx, params)
}

func (a *Adapter) DeleteSubscription(
	ctx context.Context,
	params common.SubscriptionResult,
) error {
	watchURL, err := url.JoinPath(a.ModuleInfo().BaseURL, apiVersion, "users/me/stop")
	if err != nil {
		return fmt.Errorf("delete subscribe: building watch URL: %w", err)
	}

	// The request body. must be empty.
	// ref: https://developers.google.com/workspace/gmail/api/reference/rest/v1/users/stop
	response, err := a.JSONHTTPClient().Post(ctx, watchURL, nil)
	if err != nil {
		return fmt.Errorf("delete subscribe: posting to gmail watch: %w", err)
	}

	_, err = common.UnmarshalJSON[watchResponse](response)
	if err != nil {
		return err
	}

	return nil
}

func (a *Adapter) GetRecordsByIds(ctx context.Context, // nolint: revive
	objectName string,
	recordIds []string, //nolint:revive
	fields []string,
	associations []string) ([]common.ReadResultRow, error) {
	return nil, common.ErrGetRecordNotSupportedForObject
}

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
	return true, nil
}
