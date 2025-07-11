package zohocrm

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/go-playground/validator"
	"github.com/mitchellh/hashstructure"
)

const (
	maxGroupSize = 2
)

var (
	_ connectors.SubscribeConnector              = &Connector{}
	_ connectors.SubscriptionMaintainerConnector = &Connector{}
)

func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{}
}

func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &WatchResult{},
	}
}

// Subscribe subscribes to the events for the given params.
// It returns a subscription result with the channel id.
func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	req, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	hashedChannelId, err := hashString(req.UniqueRef)
	if err != nil {
		return nil, err
	}

	return c.putOrPostSubscribe(ctx, params, req, c.Client.Post, hashedChannelId)
}

func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	req, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	hashedChannelId, err := hashString(req.UniqueRef)
	if err != nil {
		return nil, err
	}

	if err := validateChannelId(previousResult, hashedChannelId); err != nil {
		return nil, err
	}

	return c.putOrPostSubscribe(ctx, params, req, c.Client.Put, hashedChannelId)
}

// DeleteSubscription deletes a subscription with channel id which is extracted from the previous result.
// previousResult is validated to make sure that there is only 1 channel id in the result.
func (c *Connector) DeleteSubscription(ctx context.Context, result common.SubscriptionResult) error {
	if result.Result == nil {
		return fmt.Errorf("%w: Result cannot be null", errMissingParams) //nolint:err113,lll
	}

	watchRes, ok := result.Result.(*WatchResult)
	if !ok {
		return fmt.Errorf("%w: expected SubscriptionResult to be type '%T', but got '%T'", errInvalidRequestType, watchRes, result.Result) //nolint:err113,lll
	}

	if len(watchRes.Details.Events) == 0 {
		return fmt.Errorf("%w: events cannot be empty", errMissingParams) //nolint:err113,lll
	}

	//nolint:revive
	channelIds := datautils.NewSet[string]()

	var channelId string

	for _, event := range watchRes.Details.Events {
		channelIds.AddOne(event.ChannelID)
		channelId = event.ChannelID
	}

	if len(channelIds) == 0 {
		return fmt.Errorf("%w: no channel ids found", errMissingParams)
	}

	if len(channelIds) != 1 {
		return fmt.Errorf("%w: %s", errInconsistentChannelIdsMismatch, channelIds.List())
	}

	err := c.deleteNotifications(ctx, channelId)
	if err != nil {
		return fmt.Errorf("failed to delete notification channel: %w", err)
	}

	return nil
}

// RunScheduledMaintenance runs the schedule for the connector to maintain the subscription.
func (c *Connector) RunScheduledMaintenance(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	// In order to maintain the subscription, we need to
	// update the expiration time of the subscription
	// Available API is PUT or PATCH endpoint.
	// Using PATCH will require parsing exisiting subscription and
	// reformulating the request body which is complicated and error prone.
	// Our UpdateSubscription uses PUT endpoint and it will automatically extend the expiry time.
	// So we just use UpdateSubscription to maintain the subscription.
	return c.UpdateSubscription(ctx, params, previousResult)
}

// DeleteNotifcations disable all notification for list of channelIDs
// https://www.zoho.com/crm/developer/docs/api/v7/notifications/update-details.html
func (c *Connector) deleteNotifications(ctx context.Context, channelIDs string) error {
	url, err := c.getSubscribeURL()
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

//nolint:funlen,cyclop
func (c *Connector) putOrPostSubscribe(
	ctx context.Context,
	params common.SubscribeParams,
	req *SubscriptionRequest,
	putOrPost common.WriteMethod,
	channelId string,
) (*common.SubscriptionResult, error) {
	if req.Duration != nil && *req.Duration > defaultDuration {
		return nil, errInvalidDuration
	}

	payload := &SubscriptionPayload{
		Watch: make([]Watch, 0),
	}

	// in order to build the payload, we need to get the module metadata to get the object name and object type id
	moduleMetadataMap, err := c.getModuleMetadata(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error getting module metadata map: %w", err)
	}

	//nolint:varnamelen
	var wg sync.WaitGroup

	var subscriptionErr error

	var dur time.Duration
	if req.Duration != nil {
		dur = *req.Duration
	} else {
		// default 1 week
		dur = defaultDuration
	}

	exp := time.Now().Add(dur)

	expiryStr := datautils.Time.FormatRFC3339inUTC(exp)

	watchObject := Watch{
		ChannelID:                 channelId,
		NotifyURL:                 req.WebhookEndPoint,
		Token:                     req.UniqueRef,
		ReturnAffectedFieldValues: true,
		NotifyOnRelatedAction:     false, // TODO: [ENG-2229] Enable this when association update is enabled
		ChannelExpiry:             expiryStr,
	}

	var mutex sync.Mutex

	// iterate over all objects and events to build the payload
	for obj, evt := range params.SubscriptionEvents {
		wg.Add(1)

		go func(objName common.ObjectName, event common.ObjectEvents) {
			defer wg.Done()

			mappedEvents, err := mapEvents(string(objName), event.Events)
			if err != nil {
				subscriptionErr = errors.Join(
					subscriptionErr,
					fmt.Errorf("error mapping events: %w", err),
				)

				return
			}

			moduleMetadata := moduleMetadataMap[string(objName)]

			// get notification conditions per object
			notificationConditions, goroutineErr := c.getNotificationConditions(ctx, moduleMetadata, event)
			if goroutineErr != nil {
				subscriptionErr = errors.Join(
					subscriptionErr,
					fmt.Errorf("error getting notification conditions for object %s: %w", objName, goroutineErr),
				)

				return
			}

			mutex.Lock()
			watchObject.Events = append(watchObject.Events, mappedEvents...)
			watchObject.NotificationCondition = append(watchObject.NotificationCondition, notificationConditions...)

			mutex.Unlock()
		}(obj, evt)
	}

	wg.Wait()

	payload.Watch = append(payload.Watch, watchObject)

	if subscriptionErr != nil {
		return nil, fmt.Errorf("error preparing to subscribe: %w", subscriptionErr)
	}

	res, err := c.putOrPostSubscription(ctx, payload, putOrPost)
	if err != nil {
		return nil, fmt.Errorf("error enabling subscription: %w", err)
	}

	subscriptionResult := &common.SubscriptionResult{
		Result:       res,
		ObjectEvents: params.SubscriptionEvents,
		Status:       common.SubscriptionStatusSuccess,
	}

	return subscriptionResult, nil
}

func (c *Connector) putOrPostSubscription(
	ctx context.Context,
	payload *SubscriptionPayload,
	updater common.WriteMethod,
) (*WatchResult, error) {
	url, err := c.getSubscribeURL()
	if err != nil {
		return nil, fmt.Errorf("error getting subscribe URL: %w", err)
	}

	resp, err := updater(ctx, url.String(), payload)
	if err != nil {
		return nil, fmt.Errorf("error creating subscription: %w", err)
	}

	body, err := common.UnmarshalJSON[Result](resp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling subscription response: %w", err)
	}

	if len(body.Watch) == 0 {
		return nil, errNoSubscriptionCreated
	}

	if body.Watch[0].Code != ResultStatusSuccess {
		return nil, fmt.Errorf("%w: %s", errSubscriptionFailed, body.Watch[0].Message)
	}

	return &body.Watch[0], nil
}

func mapEvents(apiName string, events []common.SubscriptionEventType) ([]ModuleEvent, error) {
	moduleEvents := make([]ModuleEvent, 0)

	for _, event := range events {
		//nolint:exhaustive
		switch event {
		case common.SubscriptionEventTypeCreate:
			moduleEvents = append(moduleEvents, ModuleEvent(apiName+"."+OperationCreate))
		case common.SubscriptionEventTypeUpdate:
			moduleEvents = append(moduleEvents, ModuleEvent(apiName+"."+OperationEdit))
		case common.SubscriptionEventTypeDelete:
			moduleEvents = append(moduleEvents, ModuleEvent(apiName+"."+OperationDelete))
		default:
			return nil, fmt.Errorf("%w: %s", errUnsupportedEventType, event)
		}
	}

	return moduleEvents, nil
}

type ModuleMetadata struct {
	Modules []map[string]any `json:"modules"`
}

func (c *Connector) fetchModuleMetadata(ctx context.Context, metadataURL string) (*ModuleMetadata, error) {
	resp, err := c.Client.Get(ctx, metadataURL)
	if err != nil {
		return nil, fmt.Errorf("error requesting module metadata: %w", err)
	}

	response, err := common.UnmarshalJSON[ModuleMetadata](resp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling module metadata: %w", err)
	}

	return response, nil
}

func (c *Connector) getModuleMetadata(
	ctx context.Context,
	params common.SubscribeParams,
) (map[string]map[string]any, error) {
	objectNames := make([]string, 0)
	for obj := range params.SubscriptionEvents {
		objectNames = append(objectNames, string(obj))
	}

	var metadataURL string

	var err error

	//nolint:gocritic
	if len(objectNames) == 0 {
		return nil, fmt.Errorf("%w: no subscription events provided", errMissingParams)
	} else if len(objectNames) > 1 {
		metadataURL, err = c.getModulesMetadataURL()
	} else {
		metadataURL, err = c.getModuleMetadataURL(objectNames[0])
	}

	if err != nil {
		return nil, fmt.Errorf("error getting metadata URL for object(s) '%v': %w", objectNames, err)
	}

	modulesMetadata, err := c.fetchModuleMetadata(ctx, metadataURL)
	if err != nil {
		return nil, fmt.Errorf("error getting module metadata: %w", err)
	}

	objectNameMatchedModule := make(map[string]map[string]any)

	for _, objName := range objectNames {
		found := false

		for _, module := range modulesMetadata.Modules {
			//nolint:forcetypeassert
			if naming.PluralityAndCaseIgnoreEqual(objName, module["module_name"].(string)) {
				objectNameMatchedModule[objName] = module
				found = true

				break
			}
		}

		if !found {
			return nil, fmt.Errorf("%w: %s", errObjectNameNotFound, objName)
		}
	}

	return objectNameMatchedModule, nil
}

func (c *Connector) getfieldsMetadata(ctx context.Context, moduleName string) (*metadataFields, error) {
	resp, err := c.fetchFieldMetadata(ctx, moduleName)
	if err != nil {
		return nil, fmt.Errorf("error getting metadata for module '%s': %w", moduleName, err)
	}

	response, err := common.UnmarshalJSON[metadataFields](resp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling metadata for module '%s': %w", moduleName, err)
	}

	return response, nil
}

// getNotificationConditions builds the notification conditions for the given event.
// it fetches the field metadata for the given module and event to build the notification condition.
//
//nolint:cyclop,funlen,gocognit
func (c *Connector) getNotificationConditions(
	ctx context.Context,
	moduleMetadata map[string]any,
	event common.ObjectEvents,
) ([]NotificationCondition, error) {
	if event.WatchFieldsAll {
		return nil, errWatchFieldsAll
	}

	if len(event.WatchFields) > maxWatchFields {
		return nil, fmt.Errorf("%w: maximum 10 fields can be watched", errTooManyWatchFields)
	}

	if len(event.WatchFields) == 0 {
		return nil, nil
	}

	var fieldMetadata *metadataFields

	var err error

	//nolint:varnamelen
	moduleName, ok := moduleMetadata["module_name"].(string)
	if !ok {
		return nil, errModuleNameNotString
	}

	if len(event.WatchFields) > 0 {
		fieldMetadata, err = c.getfieldsMetadata(ctx, moduleName)
		if err != nil {
			return nil, fmt.Errorf("error getting fields metadata: %w", err)
		}
	}

	watchFieldsMetadata := make(map[string]map[string]any, 0)

	for _, field := range event.WatchFields {
		found := false

		for _, fm := range fieldMetadata.Fields {
			apiName, ok := fm["api_name"].(string)
			if !ok {
				continue
			}

			if naming.PluralityAndCaseIgnoreEqual(apiName, field) {
				watchFieldsMetadata[field] = fm
				found = true

				break
			}
		}

		if !found {
			return nil, fmt.Errorf("%w: %s", errFieldNotFound, field)
		}
	}

	fieldNames := make([]string, 0, len(watchFieldsMetadata))
	for fieldName := range watchFieldsMetadata {
		fieldNames = append(fieldNames, fieldName)
	}

	fieldSelection, err := buildFieldSelection(fieldNames, watchFieldsMetadata)
	if err != nil {
		return nil, err
	}

	apiName, ok := moduleMetadata["api_name"].(string)
	if !ok {
		return nil, errAPINameNotString
	}

	id, ok := moduleMetadata["id"].(string)
	if !ok {
		return nil, errIDNotString
	}

	return []NotificationCondition{
		{
			Type: "field_selection",
			Module: Module{
				APIName: apiName, // this is object name
				Id:      id,      // this is object type id
			},
			FieldSelection: fieldSelection,
		},
	}, nil
}

// according to Zoho CRM API requirements (max 2 objects per group). So we need to nest the fields
// to subscribe for more fields.
//
//nolint:cyclop,funlen
func buildFieldSelection(
	fieldNames []string,
	watchFieldsMetadata map[string]map[string]any,
) (FieldSelection, error) {
	var result FieldSelection

	var err error

	switch len(fieldNames) {
	case 0:
		result = FieldSelection{}
	case 1:
		fieldName := fieldNames[0]
		fm := watchFieldsMetadata[fieldName]

		id, ok := fm["id"].(string)
		if !ok {
			return FieldSelection{}, fmt.Errorf("%w: %s", errFieldIDNotString, fieldName)
		}

		result = FieldSelection{
			Field: &Field{
				APIName: formatAPIName(fieldName),
				ID:      id,
			},
		}
	case maxGroupSize:
		fieldGroups := make([]FieldGroup, 0)

		for _, fieldName := range fieldNames {
			fm := watchFieldsMetadata[fieldName]

			id, ok := fm["id"].(string)
			if !ok {
				return FieldSelection{}, fmt.Errorf("%w: %s", errFieldIDNotString, fieldName)
			}

			fieldGroups = append(fieldGroups, FieldGroup{
				Field: &Field{
					APIName: formatAPIName(fieldName),
					ID:      id,
				},
			})
		}

		result = FieldSelection{
			Group:         fieldGroups,
			GroupOperator: GroupOperatorOr,
		}
	default:
		// More than 2 fields - binary nest: first field, and the rest as a nested group
		firstField := fieldNames[0]
		fm := watchFieldsMetadata[firstField]

		id, ok := fm["id"].(string)
		if !ok {
			return FieldSelection{}, fmt.Errorf("%w: %s", errFieldIDNotString, firstField)
		}

		firstGroup := FieldGroup{
			Field: &Field{
				APIName: formatAPIName(firstField),
				ID:      id,
			},
		}

		nestedSelection, err := buildFieldSelection(fieldNames[1:], watchFieldsMetadata)
		if err != nil {
			return FieldSelection{}, err
		}

		var nestedGroup FieldGroup
		if nestedSelection.Field != nil {
			nestedGroup = FieldGroup{Field: nestedSelection.Field}
		} else {
			nestedGroup = FieldGroup{
				Group:         nestedSelection.Group,
				GroupOperator: string(nestedSelection.GroupOperator),
			}
		}

		result = FieldSelection{
			Group:         []FieldGroup{firstGroup, nestedGroup},
			GroupOperator: GroupOperatorOr,
		}
	}

	return result, err
}

func (c *Connector) getSubscribeURL() (*urlbuilder.URL, error) {
	url, err := urlbuilder.New(c.BaseURL, "crm/v7/actions/watch")
	if err != nil {
		return nil, err
	}

	return url, nil
}

func (c *Connector) getModuleMetadataURL(objectName string) (string, error) {
	url, err := urlbuilder.New(c.BaseURL, "crm/v7/settings/modules", objectName)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func (c *Connector) getModulesMetadataURL() (string, error) {
	url, err := urlbuilder.New(c.BaseURL, "crm/v7/settings/modules")
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func validateChannelId(previousResult *common.SubscriptionResult, hashedChannelId string) error {
	if previousResult == nil {
		return fmt.Errorf("%w: previous result is nil", errMissingParams)
	}

	watchResult, ok := previousResult.Result.(*WatchResult)
	if !ok {
		return fmt.Errorf("%w: expected SubscriptionResult to be type '%T', but got '%T'", errInvalidRequestType, watchResult, previousResult.Result) //nolint:err113,lll
	}

	if watchResult.Details.Events == nil {
		return fmt.Errorf("%w: no events to update", errMissingParams)
	}

	//nolint:revive
	channelIds := datautils.NewSet[string]()

	var channelId string

	for _, event := range watchResult.Details.Events {
		channelIds.AddOne(event.ChannelID)
		channelId = event.ChannelID
	}

	if len(channelIds) == 0 {
		return fmt.Errorf("%w: no channel ids found", errMissingParams)
	}

	if len(channelIds) != 1 {
		return fmt.Errorf("%w: %s", errInconsistentChannelIdsMismatch, channelIds.List())
	}

	if channelId == hashedChannelId {
		return nil
	}

	return fmt.Errorf("%w: channel id mismatch", errChannelIdMismatch)
}

func validateRequest(params common.SubscribeParams) (*SubscriptionRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(*SubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected '%T', got '%T'", errInvalidRequestType, req, params.Request)
	}

	validate := validator.New()

	if validate.Struct(req) != nil {
		return nil, fmt.Errorf("%w: request is invalid", errInvalidRequestType)
	}

	return req, nil
}

func hashString(uniqueRef string) (string, error) {
	hashedChannelID, err := hashstructure.Hash(uniqueRef, &hashstructure.HashOptions{})
	if err != nil {
		return "", fmt.Errorf("error hashing unique ref to compare with channel id: %w", err)
	}

	//nolint:gosec
	return strconv.FormatInt(int64(hashedChannelID), 10), nil
}

func formatAPIName(apiName string) string {
	parts := strings.Split(apiName, "_")

	for i, part := range parts {
		parts[i] = naming.CapitalizeFirstLetterEveryWord(part)
	}

	return strings.Join(parts, "_")
}
