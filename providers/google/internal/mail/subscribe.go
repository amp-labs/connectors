package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

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
	watchURL, err := url.JoinPath(a.ModuleInfo().BaseURL, apiVersion, "users/me/watch")
	if err != nil {
		return nil, fmt.Errorf("building watch URL: %w", err)
	}

	raw, err := json.Marshal(params.Request)
	if err != nil {
		return nil, fmt.Errorf("subscribe: marshaling request: %w", err)
	}

	var watchReq watchRequest
	if err := json.Unmarshal(raw, &watchReq); err != nil {
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
