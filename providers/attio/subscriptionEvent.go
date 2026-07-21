package attio

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/amp-labs/connectors/common"
)

type (
	SubscriptionEvent map[string]any
	//nolint: godoclint
	// Attio sends Secret in response when we subscribe to webhooks.
	// We use this secret to verify the webhook signatures.
	AttioVerificationParams struct {
		Secret string
	}
)

var (
	_ common.SubscriptionEvent       = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent = SubscriptionEvent{}

	errTypeMismatch = errors.New("type mismatch")
)

const (
	signatureHeader = "attio-signature"
)

// VerifyWebhookMessage implements WebhookVerifierConnector for Attio.
// Returns (true, nil) if signature verification succeeds.
// Returns (false, error) if verification fails or encounters an error.
// Ref: https://docs.attio.com/rest-api/guides/webhooks#authenticating
func (c *Connector) VerifyWebhookMessage(
	_ context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	if request == nil || params == nil {
		return false, fmt.Errorf("%w: request and params cannot be nil", errMissingParams)
	}

	verificationParams, err := common.AssertType[*AttioVerificationParams](params.Param)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errMissingParams, err)
	}

	signature := request.Headers.Get(signatureHeader)
	if signature == "" {
		return false, fmt.Errorf("%w: missing %s header", ErrMissingSignature, signatureHeader)
	}

	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("%w: error decoding signature: %w", ErrInvalidSignature, err)
	}

	expectedSignature := computeSignature(verificationParams.Secret, request.Body)

	if !hmac.Equal(sigBytes, expectedSignature) {
		return false, fmt.Errorf("%w: signature mismatch", ErrInvalidSignature)
	}

	return true, nil
}

func (evt SubscriptionEvent) PreLoadData(data *common.SubscriptionEventPreLoadData) error {
	return nil
}

func (evt SubscriptionEvent) UpdatedFields() ([]string, error) {
	// Attio webhooks do not provide updated field information, so we return an
	// empty list without an error.
	return []string{}, nil
}

func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	return 0, errors.New("attio webhooks do not include event timestamps") //nolint:err113
}

func (evt SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	subTypeStr, err := evt.RawEventName()
	if err != nil {
		return common.SubscriptionEventTypeOther, fmt.Errorf("error getting raw event name: %w", err)
	}

	parts := strings.Split(subTypeStr, ".")
	if len(parts) < 2 { //nolint:mnd
		// this should never happen unless the provider changes subscription event format
		return common.SubscriptionEventTypeOther, fmt.Errorf(
			"%w: '%s'", errUnsupportedSubscriptionEvent, subTypeStr,
		)
	}

	switch parts[1] {
	case "created":
		return common.SubscriptionEventTypeCreate, nil
	case "updated":
		return common.SubscriptionEventTypeUpdate, nil
	case "deleted":
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (evt SubscriptionEvent) ObjectName() (string, error) {
	name, err := evt.RawEventName()
	if err != nil {
		return "", err
	}

	objectName := strings.Split(name, ".")[0]

	return objectName, nil
}

func (evt SubscriptionEvent) RawEventName() (string, error) {
	// asMap returns the single event object ({event_type, id, ...}) extracted
	// from the top-level events array, so event_type is read directly off it.
	event := evt.asMap()

	eventName, ok := event["event_type"].(string)
	if !ok {
		return "", fmt.Errorf("%w: expected string event_type, got %T", errTypeMismatch, event["event_type"])
	}

	return eventName, nil
}

func (evt SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(evt), nil
}

// recordIDKeyByEventObject maps the object portion of an Attio webhook event_type
// (the part before the ".", e.g. "note-content" in "note-content.updated") to the
// key that holds the affected record's identifier inside the event "id" object.
//
// A naive objectName+"_id" is wrong for some events: note-content events carry a
// "note_id" (not "note-content_id"), and workspace-member events carry a
// "workspace_member_id" (underscore, not the hyphenated object name). Example
// "id" objects from the docs for these two non-obvious cases:
//
//	note-content.updated:     {"workspace_id": "...", "note_id": "..."}
//	workspace-member.created: {"workspace_id": "...", "workspace_member_id": "..."}
//
// Keys verified against Attio's webhook reference (see each event's "id" schema
// and example):
//   - source of truth: https://api.attio.com/openapi/webhooks
//   - record.*:           https://docs.attio.com/rest-api/webhook-reference/record-events/recordcreated
//   - list.*:             https://docs.attio.com/rest-api/webhook-reference/list-events/listcreated
//   - task.*:             https://docs.attio.com/rest-api/webhook-reference/task-events/taskcreated
//   - note.*:             https://docs.attio.com/rest-api/webhook-reference/note-events/notecreated
//   - note-content.*:     https://docs.attio.com/rest-api/webhook-reference/note-content-events/note-contentupdated
//   - workspace-member.*: https://docs.attio.com/rest-api/webhook-reference/workspace-member-events/workspace-membercreated
//
//nolint:gochecknoglobals
var recordIDKeyByEventObject = map[string]string{
	"record":           "record_id",
	"list":             "list_id",
	"task":             "task_id",
	"note":             "note_id",
	"note-content":     "note_id",
	"workspace-member": "workspace_member_id",
}

func (evt SubscriptionEvent) RecordId() (string, error) {
	idMap, err := evt.idMap()
	if err != nil {
		return "", err
	}

	eventName, err := evt.RawEventName()
	if err != nil {
		return "", err
	}

	// event_type has the form "{object}.{action}"; the object portion determines
	// which key inside the "id" object holds the record identifier.
	eventObject := strings.Split(eventName, ".")[0]

	idKey, ok := recordIDKeyByEventObject[eventObject]
	if !ok {
		return "", fmt.Errorf("%w: no record id key mapping for event %q", errTypeMismatch, eventName)
	}

	return lookupID(idMap, idKey)
}

func (evt SubscriptionEvent) Workspace() (string, error) {
	idMap, err := evt.idMap()
	if err != nil {
		return "", err
	}

	return lookupID(idMap, "workspace_id")
}

// idMap returns the "id" object of the event as a map[string]any.
// Attio's webhook id object holds the workspace_id and object-specific
// identifiers (e.g. note_id, record_id).
func (evt SubscriptionEvent) idMap() (map[string]any, error) {
	event := evt.asMap()

	idMap, ok := event["id"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected id to be map[string]any, got %T", errTypeMismatch, event["id"])
	}

	return idMap, nil
}

// lookupID returns the string value stored under key in the event's id map.
func lookupID(idMap map[string]any, key string) (string, error) {
	value, ok := idMap[key]
	if !ok {
		return "", fmt.Errorf("%w: id map does not contain %q", errTypeMismatch, key)
	}

	id, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%w: expected %q to be string, got %T", errTypeMismatch, key, value)
	}

	return id, nil
}

// asMap returns the event as a StringMap.
func (evt SubscriptionEvent) asMap() common.StringMap {
	// extract first event from events array
	// Attio sends an array of events, but it only contains one event per webhook call.
	// So we extract the first event for processing.
	evtsArray, ok := evt["events"].([]any)
	if ok && len(evtsArray) > 0 {
		firstEvt, ok := evtsArray[0].(map[string]any)
		if ok {
			return common.StringMap(firstEvt)
		}
	}

	// Fallback to returning the whole event if extraction fails
	return common.StringMap(evt)
}

func computeSignature(secret string, body []byte) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)

	return h.Sum(nil)
}

// GetFieldNameFromObjectMetadata looks up a field's api_slug from metadata using the object_id and attribute_id.
// It returns an error if the object or attribute is not found.
func GetFieldNameFromObjectMetadata(
	metadata *common.ListObjectMetadataResult,
	objectID string,
	attributeID string,
) (string, error) {
	obj, ok := metadata.Result[objectID]
	if !ok {
		return "", fmt.Errorf("%w: object %q", common.ErrNotFound, objectID)
	}

	for fieldName, field := range obj.Fields {
		if field.FieldId != nil && *field.FieldId == attributeID {
			return fieldName, nil
		}
	}

	return "", fmt.Errorf("%w: attribute %q in object %q", common.ErrNotFound, attributeID, objectID)
}

// GetObjectNameFromObjectMetadata looks up an object's display name from metadata using the object_id.
// It returns an error if the object is not found.
func GetObjectNameFromObjectMetadata(
	metadata *common.ListObjectMetadataResult,
	objectID string,
) (string, error) {
	obj, ok := metadata.Result[objectID]
	if !ok {
		return "", fmt.Errorf("%w: object %q", common.ErrNotFound, objectID)
	}

	return obj.DisplayName, nil
}

// Example: Webhook response
/*
{
  "webhook_id": "04731154-70d3-42bb-8320-760304c9bbfd",
  "events": [
    {
      "event_type": "note.updated",
      "id": {
        "workspace_id": "e293215c-210a-4d4a-9913-e2b33da318ab",
        "note_id": "f83d5cab-571b-47a8-8018-57146f848d19"
      },
      "parent_object_id": "ee1e6aa1-ec69-4ef4-a101-3a9abb12e281",
      "parent_record_id": "9bcad14b-55a5-478d-963b-a4ec598265c6",
      "actor": {
        "type": "workspace-member",
        "id": "f0519378-80b8-4d7c-8874-c6acc1850442"
      }
    }
  ]
}
*/
