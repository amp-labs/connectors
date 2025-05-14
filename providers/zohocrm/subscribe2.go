package zohocrm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	OperationCreate = "create"
	OperationEdit   = "edit"
	OperationDelete = "delete"
	OperationAll    = "all"

	maxWatchFields = 10

	ResultStatusSuccess = "SUCCESS"
)

var errInvalidModuleEvent = errors.New("invalid module event")

// ModuleEvent represents a module and operation combination.
type ModuleEvent string

// String returns the formatted string representation of the module event.
func (me ModuleEvent) ModuleAPI() (string, error) {
	parts, err := me.parts()
	if err != nil {
		return "", err
	}

	return parts[0], nil
}

func (me ModuleEvent) Operation() (string, error) {
	parts, err := me.parts()
	if err != nil {
		return "", err
	}

	return parts[1], nil
}

func (me ModuleEvent) parts() ([]string, error) {
	parts := strings.Split(string(me), ".")
	//nolint:mnd
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w: %s", errInvalidModuleEvent, me)
	}

	return parts, nil
}

type SubscriptionPayload struct {
	Watch []Watch `json:"watch"`
}

//nolint:tagliatelle
type NotificationCondition struct {
	Type           string         `json:"type"`
	Module         Module         `json:"module"`
	FieldSelection FieldSelection `json:"field_selection"`
}

//nolint:tagliatelle
type Module struct {
	APIName string `json:"api_name"`
	Id      string `json:"id"`
}

type GroupOperator string

const (
	GroupOperatorOr  = "or"
	GroupOperatorAnd = "and"
)

//nolint:tagliatelle
type FieldSelection struct {
	GroupOperator GroupOperator `json:"group_operator"`
	Group         []FieldGroup  `json:"group"`
}

//nolint:tagliatelle
type FieldGroup struct {
	Field         *Field       `json:"field,omitempty"`
	GroupOperator string       `json:"group_operator,omitempty"`
	Group         []FieldGroup `json:"group,omitempty"`
}

//nolint:tagliatelle
type Field struct {
	APIName string `json:"api_name"`
	ID      string `json:"id"`
}

//nolint:tagliatelle
type Watch struct {
	ChannelID                 string                  `json:"channel_id"`
	Events                    []ModuleEvent           `json:"events"`
	NotificationCondition     []NotificationCondition `json:"notification_condition,omitempty"`
	ChannelExpiry             *time.Time              `json:"channel_expiry,omitempty"`
	Token                     string                  `json:"token,omitempty"`
	ReturnAffectedFieldValues bool                    `json:"return_affected_field_values,omitempty"`
	NotifyURL                 string                  `json:"notify_url"`
	NotifyOnRelatedAction     bool                    `json:"notify_on_related_action,omitempty"`
}

func (c *Connector) getSubscribeURL() (string, error) {
	url, err := urlbuilder.New(c.BaseURL, "crm/v7/actions/watch")
	if err != nil {
		return "", err
	}

	return url.String(), nil
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

//nolint:tagliatelle
type SubscriptionRequest struct {
	UniqueRef       string         `json:"unique_ref"`
	WebhookEndPoint string         `json:"webhook_end_point"`
	Duration        *time.Duration `json:"duration,omitempty"`
}

var (
	errWatchFieldsAll        = errors.New("watch fields all is not supported")
	errTooManyWatchFields    = errors.New("too many watch fields")
	errSubscriptionFailed    = errors.New("subscription failed")
	errNoSubscriptionCreated = errors.New("no subscription created")
	errUnsupportedEventType  = errors.New("unsupported event type")
	errFieldNotFound         = errors.New("field not found")
	errObjectNameNotFound    = errors.New("object name not found")
)

type Result struct {
	Watch []WatchResult `json:"watch"`
}

//nolint:funlen
func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(*SubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected '%T', got '%T'", errInvalidRequestType, req, params.Request)
	}

	payload := &SubscriptionPayload{
		Watch: make([]Watch, 0),
	}

	moduleMetadataMap, err := c.getModuleMetadata(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error getting module metadata map: %w", err)
	}

	//nolint:varnamelen
	var wg sync.WaitGroup

	var subscriptionErr error

	exp := time.Now().Add(*req.Duration)

	// iterate over all objects and events
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

			formattedObjName := naming.CapitalizeFirstLetterEveryWord(string(objName))
			moduleMetadata := moduleMetadataMap[string(objName)]
			watchObject := Watch{
				ChannelID:                 req.UniqueRef + "_" + formattedObjName,
				Events:                    mappedEvents, // this will list of events for the object
				NotifyURL:                 req.WebhookEndPoint + "/objects/" + string(objName),
				Token:                     req.UniqueRef,
				ReturnAffectedFieldValues: true,
				NotifyOnRelatedAction:     false, // TODO: [ENG-2229] Enable this when association update is enabled
				ChannelExpiry:             &exp,
			}

			// get notification conditions per object
			notificationConditions, goroutineErr := c.getNotificationConditions(ctx, moduleMetadata, event)
			if goroutineErr != nil {
				subscriptionErr = errors.Join(
					subscriptionErr,
					fmt.Errorf("error getting notification conditions: %w", goroutineErr),
				)

				return
			}

			watchObject.NotificationCondition = notificationConditions

			payload.Watch = append(payload.Watch, watchObject)
		}(obj, evt)
	}

	wg.Wait()

	if subscriptionErr != nil {
		return nil, fmt.Errorf("error subscribing to events: %w", subscriptionErr)
	}

	res, err := c.enableSubscription(ctx, payload)
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

func (c *Connector) enableSubscription(ctx context.Context, payload *SubscriptionPayload) (*Result, error) {
	url, err := c.getSubscribeURL()
	if err != nil {
		return nil, fmt.Errorf("error getting subscribe URL: %w", err)
	}

	resp, err := c.Client.Post(ctx, url, payload)
	if err != nil {
		return nil, fmt.Errorf("error creating subscription: %w", err)
	}

	body, err := common.UnmarshalJSON[Result](resp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling subscription response: %w", err)
	}

	if body.Watch[0].Code != ResultStatusSuccess {
		return nil, fmt.Errorf("%w: %s", errSubscriptionFailed, body.Watch[0].Message)
	}

	if len(body.Watch) == 0 {
		return nil, errNoSubscriptionCreated
	}

	return body, nil
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

func (c *Connector) getfieldsMetadata(ctx context.Context, moduleMetadata map[string]any) (*metadataFields, error) {
	//nolint:forcetypeassert
	moduleName := moduleMetadata["module_name"].(string)

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

//nolint:cyclop,funlen
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

	if len(event.WatchFields) > 0 {
		fieldMetadata, err = c.getfieldsMetadata(ctx, moduleMetadata)
		if err != nil {
			return nil, fmt.Errorf("error getting fields metadata: %w", err)
		}
	}

	watchFieldsMetadata := make(map[string]map[string]any, 0)

	for _, field := range event.WatchFields {
		found := false

		for _, fieldMetadata := range fieldMetadata.Fields {
			//nolint:forcetypeassert
			if naming.PluralityAndCaseIgnoreEqual(fieldMetadata["api_name"].(string), field) {
				watchFieldsMetadata[field] = fieldMetadata
				found = true

				break
			}
		}

		if !found {
			return nil, fmt.Errorf("%w: %s", errFieldNotFound, field)
		}
	}

	fieldGroups := make([]FieldGroup, 0)
	//nolint:forcetypeassert
	for fieldName, fieldMetadata := range watchFieldsMetadata {
		fieldGroups = append(fieldGroups, FieldGroup{
			Field: &Field{
				APIName: fieldName,

				ID: fieldMetadata["id"].(string),
			},
		})
	}

	//nolint:forcetypeassert
	return []NotificationCondition{
		{
			Type: "field_selection",
			Module: Module{
				APIName: moduleMetadata["api_name"].(string), // this is object name
				Id:      moduleMetadata["id"].(string),       // this is object type id
			},
			FieldSelection: FieldSelection{
				GroupOperator: GroupOperatorOr,
				Group:         fieldGroups,
			},
		},
	}, nil
}
