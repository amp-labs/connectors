package zohocrm

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
)

type SubscribeResult struct {
	Notifications map[common.ObjectName]*Notification
}

// Notification represents a Zoho CRM notification subscription.
type Notification struct {
	ChannelID     string
	NotifyURL     string
	Events        []string
	WatchFields   []string
	Token         string `json:"token,omitempty"`
	ChannelExpiry string
}

// nolin:funlen
// Subscribe subscribes to events for the given objects
// This is where the actual API calls to Zoho CRM happen to create notification subscriptions
// Zoho CRM doesn't require registration - we directly subscribe to events.
//
//nolint:funlen, cyclop
func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	// Generate a unique channel ID
	channelID := strconv.FormatInt(time.Now().UnixNano(), 10)

	// The expiry date can be a maximum of one week from the time of subscribe.
	//  If it is not specified or set for more than a week, the default expiry time is for one hour.
	// Setting this 6 days just to be on safe side.
	channelExpiryTime := datautils.Time.FormatRFC3339inUTC(time.Now().Add(time.Hour * 24 * 6)) //nolint:mnd

	notifyURL := "https://play.svix.com/in/e_Z4PpxWo75NamyQ2qBOJkrN7SsM6/"
	token := "test_token"

	zohoRes := &SubscribeResult{
		Notifications: make(map[common.ObjectName]*Notification),
	}

	var failError error

	objectsSubscribed := []common.ObjectName{}

	for objName, objEvents := range params.SubscriptionEvents {
		// Convert object name to proper case for Zoho CRM API
		zohoObjName := getZohoObjectName(string(objName))

		events := []string{}

		for _, eventType := range objEvents.Events {
			//nolint:exhaustive
			switch eventType {
			case common.SubscriptionEventTypeCreate:
				events = append(events, zohoObjName+".create")
			case common.SubscriptionEventTypeUpdate:
				events = append(events, zohoObjName+".edit")
			case common.SubscriptionEventTypeDelete:
				events = append(events, zohoObjName+".delete")
			default:
				events = append(events, zohoObjName+".all")
			}
		}

		if len(events) == 0 {
			events = append(events, zohoObjName+".all")
		}

		notification := &Notification{
			ChannelID:     channelID,
			NotifyURL:     notifyURL,
			Events:        events,
			WatchFields:   objEvents.WatchFields,
			Token:         token,
			ChannelExpiry: channelExpiryTime,
		}

		// Create notification in Zoho CRM
		newNotification, err := c.CreateNotification(ctx, notification)
		if err != nil {
			failError = fmt.Errorf("failed to create notification for object %s: %w", objName, err)

			break
		}

		zohoRes.Notifications[objName] = newNotification

		objectsSubscribed = append(objectsSubscribed, objName)
	}

	if failError != nil {
		channelIDs := []string{}

		for _, notification := range zohoRes.Notifications {
			channelIDs = append(channelIDs, notification.ChannelID)
		}

		chnanelIDStr := strings.Join(channelIDs, ",")

		err := c.DeleteNotifications(ctx, chnanelIDStr)
		if err != nil {
			return &common.SubscriptionResult{
				Status:  common.SubscriptionStatusFailedToRollback,
				Result:  zohoRes,
				Objects: objectsSubscribed,
				Events:  getRequstedEventTypes(params.SubscriptionEvents),
			}, fmt.Errorf("failed to rollback: %w, original erro :%w", err, failError)
		}

		return nil, failError
	}

	return &common.SubscriptionResult{
		Status:  common.SubscriptionStatusSuccess,
		Result:  zohoRes,
		Events:  getRequstedEventTypes(params.SubscriptionEvents),
		Objects: objectsSubscribed,
	}, nil
}

// UpdateSubscription will update subscription by :
// 1. Removing objects from the previous subscription that are not in the new subscription
// 2. Adding new objects to the subscription that in the new subscription but not in the previous subscription
// 3. Returing the updated subscription result.
//
// nolint:funlen,lll,cyclop
func (c *Connector) UpdateSubscription(ctx context.Context, params common.SubscribeParams, previousResult *common.SubscriptionResult) (*common.SubscriptionResult, error) {
	if previousResult.Result == nil {
		return nil, fmt.Errorf("%w, missing previousResult.Result", errMissingParams)
	}

	prevState, ok := previousResult.Result.(*SubscribeResult)

	if !ok {
		return nil, fmt.Errorf("%w: expected previousResult.Result to be type '%T', but got '%T'",
			errInvalidRequestType,
			prevState,
			previousResult.Result)
	}

	objectsToDelete := []common.ObjectName{}
	objectsToAdd := []common.ObjectName{}

	// collect objects to exclude from subscription
	for objName := range prevState.Notifications {
		_, ok := params.SubscriptionEvents[objName]
		if !ok {
			objectsToDelete = append(objectsToDelete, objName)
		}
	}

	// collect new objects to add to the subscription
	for objName := range params.SubscriptionEvents {
		_, ok := prevState.Notifications[objName]
		if !ok {
			objectsToAdd = append(objectsToAdd, objName)
		}
	}

	// remove objects that is to be exluded from subscription and delete
	for _, objName := range objectsToAdd {
		_, ok := params.SubscriptionEvents[objName]
		if !ok {
			delete(params.SubscriptionEvents, objName)
		}
	}

	NotificatiosToKeep := make(map[common.ObjectName]*Notification)

	// Remove objects to exclue from delete
	for _, objName := range objectsToDelete {
		_, ok := prevState.Notifications[objName]
		if !ok {
			NotificatiosToKeep[objName] = prevState.Notifications[objName]
			delete(prevState.Notifications, objName)
		}
	}

	deleteParams := *previousResult
	deleteParams.Objects = objectsToDelete
	deleteParams.Result = prevState.Notifications

	err := c.DeleteSubscription(ctx, deleteParams)
	if err != nil {
		return nil, fmt.Errorf("failed to delete previous subscription: %w", err)
	}

	zohRes, err := c.Subscribe(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to new objets: %w", err)
	}

	newState := prevState

	newState.Notifications = NotificatiosToKeep

	for objName, notification := range zohRes.Result.(*SubscribeResult).Notifications { //nolint:forcetypeassert
		newState.Notifications[objName] = notification
	}

	objectsSubscribed := []common.ObjectName{}

	for objName := range newState.Notifications {
		objectsSubscribed = append(objectsSubscribed, objName)
	}

	res := &common.SubscriptionResult{
		Status:  common.SubscriptionStatusSuccess,
		Result:  newState,
		Objects: objectsSubscribed,
		Events:  getRequstedEventTypes(params.SubscriptionEvents),
	}

	return res, nil
}

// DeleteSubscription deletes a subscription by deleting all the notifications.
// If any of the notification is failed to delete, it will return an error.
func (c *Connector) DeleteSubscription(ctx context.Context, params common.SubscriptionResult) error {
	if params.Result == nil {
		return errors.New("missing SubscriptionResult") //nolint:err113
	}

	zohoRes, ok := params.Result.(*SubscribeResult)

	if !ok {
		return fmt.Errorf("%w: expected SubscriptionResult to be type '%T', but got '%T'", errInvalidRequestType, zohoRes, params.Result) //nolint:err113,lll
	}

	channelIDs := []string{}

	for _, notification := range zohoRes.Notifications {
		channelIDs = append(channelIDs, notification.ChannelID)
	}

	channelIDStr := strings.Join(channelIDs, ",")

	err := c.DeleteNotifications(ctx, channelIDStr)
	if err != nil {
		return fmt.Errorf("failed to delete notification channel: %w", err)
	}

	return nil
}

// CreateNotification subscribe to the webhook
// https://www.zoho.com/crm/developer/docs/api/v7/notifications/enable.html
func (c *Connector) CreateNotification(ctx context.Context, notification *Notification) (*Notification, error) {
	url, err := c.getAPIURL("actions/watch")
	if err != nil {
		return nil, err
	}

	requestBody := map[string]any{
		"watch": []map[string]any{
			{
				"channel_id":                   notification.ChannelID,
				"events":                       notification.Events,
				"channel_expiry":               notification.ChannelExpiry,
				"return_affected_field_values": true,
				"notify_url":                   notification.NotifyURL,
			},
		},
	}

	resp, err := c.Client.Post(ctx, url.String(), requestBody)
	if err != nil {
		return nil, err
	}

	responsePtr, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return nil, err
	}

	response := *responsePtr

	watchResponse, ok := response["watch"].([]any)
	if !ok || len(watchResponse) == 0 {
		return nil, errInvalidResponse
	}

	watchResult, ok := watchResponse[0].(map[string]any)
	if !ok {
		return nil, errInvalidResponse
	}

	if watchResult["code"] != "SUCCESS" {
		return nil, fmt.Errorf("failed to create notification: %v", watchResult["message"]) //nolint:err113
	}

	return notification, nil
}

// DeleteNotifcations disable all notification for list of channelIDs
// https://www.zoho.com/crm/developer/docs/api/v7/notifications/update-details.html
func (c *Connector) DeleteNotifications(ctx context.Context, channelIDs string) error {
	url, err := c.getAPIURL("actions/watch")
	if err != nil {
		return err
	}

	url.WithQueryParam("channel_ids", channelIDs)

	_, err = c.Client.Delete(ctx, url.String())
	if err != nil {
		return err
	}

	return nil
}

func getZohoObjectName(objName string) string {
	return naming.CapitalizeFirstLetterEveryWord(objName)
}

func getRequstedEventTypes(events map[common.ObjectName]common.ObjectEvents) []common.SubscriptionEventType {
	uniqueEvents := make(map[common.SubscriptionEventType]bool)

	for _, objEvents := range events {
		for _, eventType := range objEvents.Events {
			uniqueEvents[eventType] = true
		}
	}

	if len(uniqueEvents) == 0 {
		return []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
			common.SubscriptionEventTypeDelete,
			common.SubscriptionEventTypeUpdate,
		}
	}

	result := make([]common.SubscriptionEventType, 0, len(uniqueEvents))

	for eventType := range uniqueEvents {
		result = append(result, eventType)
	}

	return result
}
