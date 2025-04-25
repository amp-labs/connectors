package zohocrm

import (
	"context"
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
)

// ModuleEvent represents a module and operation combination
type ModuleEvent string

// String returns the formatted string representation of the module event
func (me ModuleEvent) moduleAPI() (string, error) {
	parts, err := me.parts()
	if err != nil {
		return "", err
	}

	return parts[0], nil
}

func (me ModuleEvent) operation() (string, error) {
	parts, err := me.parts()
	if err != nil {
		return "", err
	}

	return parts[1], nil
}

func (me ModuleEvent) parts() ([]string, error) {
	parts := strings.Split(string(me), ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid module event: %s", me)
	}

	return parts, nil
}

type SubscriptionPayload struct {
	Watch []Watch `json:"watch"`
}

type NotificationCondition struct {
	Type           string         `json:"type"`
	Module         Module         `json:"module"`
	FieldSelection FieldSelection `json:"field_selection"`
}

type Module struct {
	APIName string `json:"api_name"`
	Id      string `json:"id"`
}

type GroupOperator string

const GroupOperationOr = "or"
const GroupOperatorAnd = "and"

type FieldSelection struct {
	GroupOperator GroupOperator `json:"group_operator"`
	Group         []FieldGroup  `json:"group"`
}

type FieldGroup struct {
	Field         *Field       `json:"field,omitempty"`
	GroupOperator string       `json:"group_operator,omitempty"`
	Group         []FieldGroup `json:"group,omitempty"`
}

type Field struct {
	APIName string `json:"api_name"`
	ID      string `json:"id"`
}

type Watch struct {
	ChannelID                    string                  `json:"channel_id"`
	Events                       []ModuleEvent           `json:"events"`
	NotificationCondition        []NotificationCondition `json:"notification_condition,omitempty"`
	ChannelExpiry                *time.Time              `json:"channel_expiry,omitempty"`
	Token                        string                  `json:"token,omitempty"`
	ReturnAffectedFieldValues    bool                    `json:"return_affected_field_values,omitempty"`
	NotifyURL                    string                  `json:"notify_url"`
	NotifyOnRelatedRelatedAction bool                    `json:"notify_on_related_related_action,omitempty"`
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

type SubscriptionRequest struct {
	UniqueRef string `json:"unique_ref"`
}

var (
	errMissingParams      = fmt.Errorf("missing parameters")
	errInvalidRequestType = fmt.Errorf("invalid request type")
)

func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	subscribeURL, err := c.getSubscribeURL()
	if err != nil {
		return nil, err
	}

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

	var wg sync.WaitGroup
	var err error

	for obj, evt := range params.SubscriptionEvents {
		wg.Add(1)
		go func(objName common.ObjectName, event common.ObjectEvents) {
			var goroutineErr error
			ctx, cancel := context.WithCancel(ctx)
			defer wg.Done()

			formattedObjName := naming.CapitalizeFirstLetterEveryWord(string(objName))
			moduleMetadata := moduleMetadataMap[string(objName)]
			watchObject := Watch{
				ChannelID: req.UniqueRef + "_" + formattedObjName,
				Events:    mapEvents(string(objName), event.Events),
			}

			var fieldMetadata *metadataFields
			if len(event.WatchFields) > 0 {
				fieldMetadata, goroutineErr = c.getfieldsMetadata(ctx, moduleMetadata)
				if goroutineErr != nil {
					err = fmt.Errorf("error getting fields metadata: %w", goroutineErr)
					cancel()
					return
				}
			}

			fieldGroups := make([]FieldGroup, 0)

			for _, field := range event.WatchFields {
				for _, fieldMetadata := range fieldMetadata.Fields {
			}

			watchObject.NotificationCondition = []NotificationCondition{
				{
					Type: "field_selection",
					Module: Module{
						APIName: moduleMetadata["api_name"].(string),
						Id:      moduleMetadata["id"].(string),
					},
					FieldSelection: FieldSelection{
						GroupOperator: GroupOperator,
						Group:         []FieldGroup{

						},
					},
				},
			}
		}(obj, evt)
	}

	return nil, nil
}

func mapEvents(apiName string, events []common.SubscriptionEventType) []ModuleEvent {
	moduleEvents := make([]ModuleEvent, 0)

	for _, event := range events {
		switch event {
		case common.SubscriptionEventTypeCreate:
			moduleEvents = append(moduleEvents, ModuleEvent(apiName+"."+OperationCreate))
		case common.SubscriptionEventTypeUpdate:
			moduleEvents = append(moduleEvents, ModuleEvent(apiName+"."+OperationEdit))
		case common.SubscriptionEventTypeDelete:
			moduleEvents = append(moduleEvents, ModuleEvent(apiName+"."+OperationDelete))
		}
	}

	return moduleEvents
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

func modulesMetadataToMap(metadata *ModuleMetadata) map[string]any {
	modules := make(map[string]any)

	for _, module := range metadata.Modules {
		modules["module_name"] = module
	}

	return modules
}

func (c *Connector) getModuleMetadata(ctx context.Context, params common.SubscribeParams) (map[string]map[string]any, error) {
	objectNames := make([]string, 0)
	for obj, _ := range params.SubscriptionEvents {
		objectNames = append(objectNames, string(obj))
	}

	var metadataURL string
	var err error

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
			if naming.PluralityAndCaseIgnoreEqual(objName, module["module_name"].(string)) {
				objectNameMatchedModule[objName] = module
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("object name '%s' not found in module metadata", objName)
		}
	}

	return objectNameMatchedModule, nil
}

func (c *Connector) getfieldsMetadata(ctx context.Context, moduleMetadata map[string]any) (*metadataFields, error) {
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
