package zohocrm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/amp-labs/connectors/common"
)

var (
	_ common.SubscriptionEvent       = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent = SubscriptionEvent{}

	errTypeMismatch = errors.New("type mismatch")
	errMatcherDiff  = errors.New("matcher ref does not match token")
)

// SubscriptionEvent represents a webhook event from Zoho CRM.
type SubscriptionEvent map[string]any

// VerifyWebhookMessage verifies the signature of a webhook message from Zoho CRM.
func (*Connector) VerifyWebhookMessage(
	_ context.Context, params *common.WebhookVerificationParameters,
) (bool, error) {
	var body map[string]any

	err := json.Unmarshal(params.Body, &body)
	if err != nil {
		return false, err
	}

	//nolint:varnamelen
	token, ok := body["token"]
	if !ok {
		return false, fmt.Errorf("%w: %s", errFieldNotFound, "token")
	}

	tokenStr, ok := token.(string)
	if !ok {
		return false, fmt.Errorf("%w: %s, expected string, got %T", errTypeMismatch, "token", token)
	}

	matcherRef, ok := params.MatcherRefs["uniqueRef"]
	if !ok {
		return false, fmt.Errorf("%w: %s", errFieldNotFound, "uniqueRef")
	}

	matcherRefStr, ok := matcherRef.(string)
	if !ok {
		return false, fmt.Errorf("%w: %s, expected string, got %T", errTypeMismatch, "matcherRef", matcherRef)
	}

	if tokenStr != matcherRefStr {
		return false, errMatcherDiff
	}

	return true, nil
}

var (
	_ common.SubscriptionEvent                      = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent                = SubscriptionEvent{}
	_ common.ReferenceIdentifiableSubscriptionEvent = SubscriptionEvent{}
)

type CollapsedSubscriptionEvent map[string]any

var _ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}

//nolint:funlen
func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	/*
		{
			"server_time": 1750102639787,
			"affected_values": [
				{
					"record_id": "6756839000000575405",
					"values": {
						"Company": "Rangoni Of Test",
						"Phone": "555-555-1111"
					}
				}
			],
			"query_params": {},
			"module": "Leads",
			"resource_uri": "https://www.zohoapis.com/crm/v2/Leads",
			"ids": [
				"6756839000000575405"
			],
			"affected_fields": [
				{
					"6756839000000575405": [
						"Company",
						"Phone"
					]
				}
			],
			"operation": "update",
			"channel_id": "1105420521999070702",
			"token": "c3504777-db15-4332-8286-478a1b5006bc"
		}
	*/
	evts := make([]common.SubscriptionEvent, 0)

	//nolint:varnamelen
	m := common.StringMap(e)

	affectedValues, err := m.Get("affected_values")
	if err != nil {
		return nil, err
	}

	affectedValuesArr, ok := affectedValues.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: %s, expected []any, got %T", errTypeMismatch, "affectedValues", affectedValues)
	}

	//nolint:varnamelen
	for _, affectedValue := range affectedValuesArr {
		affectedValueMap, ok := affectedValue.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: %s, expected map[string]any, got %T", errTypeMismatch, "affectedValue", affectedValue)
		}

		vm := common.StringMap(affectedValueMap)

		recordId, err := vm.GetString("record_id")
		if err != nil {
			return nil, err
		}

		valuesAny, err := vm.Get("values")
		if err != nil {
			return nil, err
		}

		values, ok := valuesAny.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: %s, expected map[string]any, got %T", errTypeMismatch, "values", valuesAny)
		}

		subscriptionEvent := maps.Clone(m)

		subscriptionEvent["affected_values"] = []any{
			values,
		}

		subscriptionEvent["affected_fields"] = []any{}

		fieldsMap := make(map[string][]string)
		for field := range values {
			fieldsMap[recordId] = append(fieldsMap[recordId], field)
		}

		subscriptionEvent["affected_fields"] = []any{
			fieldsMap,
		}

		subscriptionEvent["ids"] = []string{
			recordId,
		}

		evts = append(evts, SubscriptionEvent(subscriptionEvent))
	}

	return evts, nil
}

// EventType returns the type of event (create, update, delete).
func (evt SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	operation, err := evt.RawEventName()
	if err != nil {
		return common.SubscriptionEventTypeOther, fmt.Errorf("error getting raw event name: %w", err)
	}

	switch operation {
	case "create", "insert":
		return common.SubscriptionEventTypeCreate, nil
	case "edit", "update":
		return common.SubscriptionEventTypeUpdate, nil
	case "delete":
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

// RawEventName returns the raw event name from the zoho crm.
func (evt SubscriptionEvent) RawEventName() (string, error) {
	m := evt.asMap()

	operation, err := m.GetString("operation")
	if err != nil {
		return "", err
	}

	return operation, nil
}

// ObjectName returns the name of the object that triggered the event.
func (evt SubscriptionEvent) ObjectName() (string, error) {
	m := evt.asMap()

	module, err := m.GetString("module")
	if err != nil {
		return "", err
	}

	return module, nil
}

var errNotImplemented = errors.New("not implemented")

// Workspace returns the workspace ID.
func (evt SubscriptionEvent) Workspace() (string, error) {
	return "", fmt.Errorf("%w: %s", errNotImplemented, "workspace")
}

// RecordId returns the ID of the record that triggered the event.
func (evt SubscriptionEvent) RecordId() (string, error) {
	m := evt.asMap()

	idsAny, err := m.Get("ids")
	if err != nil {
		return "", fmt.Errorf("error getting record id: %w", err) //nolint:err113
	}

	// convert it to array
	ids, ok := idsAny.([]string)
	if !ok || len(ids) == 0 {
		return "", fmt.Errorf("%w: %s, expected []string, got %T", errTypeMismatch, "ids", idsAny) //nolint:err113
	}

	// Get the first ID.
	id := ids[0]

	return id, nil
}

// EventTimeStampNano returns the timestamp of the event in nanoseconds.
func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	m := evt.asMap()

	serverTime, err := m.AsInt("server_time")
	if err != nil {
		return 0, err
	}

	return time.UnixMilli(serverTime).UnixNano(), nil
}

// UpdatedFields returns the fields that were updated in the event.
func (evt SubscriptionEvent) UpdatedFields() ([]string, error) {
	m := evt.asMap()

	affectedFieldsAny, err := m.Get("affected_fields")
	if err != nil {
		return nil, err
	}

	affectedFieldsArr, ok := affectedFieldsAny.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: %s, expected []any, got %T", errTypeMismatch, "affectedFieldsAny", affectedFieldsAny)
	}

	affectedFieldsMap, ok := affectedFieldsArr[0].(map[string][]string)
	if !ok {
		return nil, fmt.Errorf("%w: %s, expected map[string][]string, got %T",
			errTypeMismatch,
			"affectedFieldsArr",
			affectedFieldsArr,
		)
	}

	recordId, err := evt.RecordId()
	if err != nil {
		return nil, err
	}

	fields := affectedFieldsMap[recordId]

	// fieldsStr := make([]string, 0, len(fields))

	// for _, field := range fields {
	// 	fieldStr, ok := field.(string)
	// 	if !ok {
	// 		return nil, errors.New("invalid field")
	// 	}
	// 	fieldsStr = append(fieldsStr, fieldStr)
	// }

	return fields, nil
}

// asMap returns the event as a StringMap.
func (evt SubscriptionEvent) asMap() common.StringMap {
	return common.StringMap(evt)
}

// Example : Webhook response
/*
{
  "server_time": 1745564776307,
  "affected_values": [
    {
      "record_id": "6172731000000939010",
      "values": {
        "Modified_By": {
          "name": "Integration User",
          "id": "6172731000000457001",
          "email": "integration.user@withampersand.com"
        },
        "Record_Status__s": 0
      }
    }
  ],
  "query_params": {},
  "module": "Leads",
  "resource_uri": "https://www.zohoapis.com/crm/v2/Leads",
  "ids": [
    "6172731000000939010"
  ],
  "affected_fields": [
    {
      "6172731000000939010": [
        "Modified_By",
        "Record_Status__s"
      ]
    }
  ],
  "operation": "delete",
  "channel_id": "1745564708612968000",
  "token": null
}

*/

func (evt SubscriptionEvent) Reference() (string, error) {
	m := evt.asMap()

	token, err := m.Get("token")
	if err != nil {
		return "", fmt.Errorf("error getting token: %w", err)
	}

	tokenStr, ok := token.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s, expected string, got %T", errTypeMismatch, "token", token)
	}

	return tokenStr, nil
}
