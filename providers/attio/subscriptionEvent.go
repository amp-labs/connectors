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
	_ common.SubscriptionEvent          = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent    = SubscriptionEvent{}
	_ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}

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

// ObjectNameWithMetadata resolves the object's name for the event using object
// metadata from the same provider.
//
// Standard and custom object changes arrive as generic record.* events that
// identify the object only by its id.object_id — a per-workspace UUID, not a name.
// This method looks that object_id up in the provided metadata. For core-object
// events (note, task, list, workspace-member) the object is already encoded in the
// event_type, so it falls back to ObjectName() without consulting metadata.
//
// The metadata must come from the same provider and be keyed by object_id (the
// same contract as GetObjectNameFromObjectMetadata).
func (evt SubscriptionEvent) ObjectNameWithMetadata(
	metadata *common.ListObjectMetadataResult,
) (string, error) {
	idMap, err := evt.idMap()
	if err != nil {
		return "", err
	}

	objectID, ok := idMap["object_id"].(string)
	if !ok {
		// No object_id: this is a core-object event whose object is in event_type.
		return evt.ObjectName()
	}

	if metadata == nil {
		return "", fmt.Errorf("%w: metadata is nil", errTypeMismatch)
	}

	return GetObjectNameFromObjectMetadata(metadata, objectID)
}

func (evt SubscriptionEvent) RawEventName() (string, error) {
	// asMap returns the single event object ({event_type, id, ...}), so event_type
	// is read directly off it.
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
// "workspace_member_id" (underscore, not the hyphenated object name).
//
// Every mapping below is confirmed against the concrete example "id" object in
// Attio's webhook reference (source of truth: https://api.attio.com/openapi/webhooks):
//
//	record.*           id: {"workspace_id","object_id","record_id"}   https://docs.attio.com/rest-api/webhook-reference/record-events/recordcreated
//	list.*             id: {"workspace_id","list_id"}                 https://docs.attio.com/rest-api/webhook-reference/list-events/listcreated
//	task.*             id: {"workspace_id","task_id"}                 https://docs.attio.com/rest-api/webhook-reference/task-events/taskcreated
//	note.*             id: {"workspace_id","note_id"}                 https://docs.attio.com/rest-api/webhook-reference/note-events/notecreated
//	note-content.*     id: {"workspace_id","note_id"}                 https://docs.attio.com/rest-api/webhook-reference/note-content-events/note-contentupdated
//	workspace-member.* id: {"workspace_id","workspace_member_id"}     https://docs.attio.com/rest-api/webhook-reference/workspace-member-events/workspace-membercreated
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

// asMap returns the single event as a StringMap. A SubscriptionEvent is one
// event object ({event_type, id, actor}) produced by
// CollapsedSubscriptionEvent.SubscriptionEventList, so no unwrapping is needed.
func (evt SubscriptionEvent) asMap() common.StringMap {
	return common.StringMap(evt)
}

// CollapsedSubscriptionEvent is the raw Attio webhook payload. Attio delivers a
// top-level object with an "events" array; each element is an individual event.
// Currently each delivery carries exactly one event, but the schema notes this
// may change to support batching, so we fan out every element.
// Ref: https://api.attio.com/openapi/webhooks
type CollapsedSubscriptionEvent map[string]any

// RawMap returns a copy of the raw payload.
func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

// SubscriptionEventList fans the top-level "events" array out into one
// SubscriptionEvent per event.
func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	rawEvents, ok := e["events"].([]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected events to be []any, got %T",
			common.ErrSubscriptionEventList, e["events"])
	}

	events := make([]common.SubscriptionEvent, 0, len(rawEvents))

	for _, raw := range rawEvents {
		eventMap, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: expected event to be map[string]any, got %T",
				common.ErrSubscriptionEventList, raw)
		}

		events = append(events, SubscriptionEvent(eventMap))
	}

	return events, nil
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
