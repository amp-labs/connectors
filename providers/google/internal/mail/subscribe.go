package mail

import (
	"context"
	"encoding/json"

	"github.com/amp-labs/connectors/common"
)

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

// Subscribe subscribes to the mailboxes events for the given params.
// It returns subscriptions expiry timestamp with the history id.
func (a *Adapter) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	var (
		watchURL = a.ModuleInfo().BaseURL + "/" + apiVersion + "/" + "users/me/watch"
		payload  watchRequest
	)

	req, err := json.Marshal(params.Request)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(req, &payload); err != nil {
		return nil, err
	}

	response, err := a.JSONHTTPClient().Post(ctx, watchURL, payload)
	if err != nil {
		return nil, err
	}

	result, err := common.UnmarshalJSON[watchResponse](response)
	if err != nil {
		return nil, err
	}

	return &common.SubscriptionResult{
		Result: result,
	}, nil
}
