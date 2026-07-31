package mail

import (
	"encoding/json"
	"maps"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
)

// Zoho Mail delivers two webhook entities on the same endpoint: Mail (new
// incoming email) and Tasks (task activity in a group). They have disjoint
// payload shapes and are told apart by their keys ("messageId" vs "entityId").
// The zoho package routes a delivery here via IsWebhookPayload, since the
// platform uses one collapsed-event type per provider.

// rawNameNewMail is the synthetic event name for a Mail webhook. The Mail entity
// fires only on newly received mail and carries no action of its own; Task
// events use their payload "action" field verbatim.
const rawNameNewMail = "newMail"

// Webhook payload keys (see the WEBHOOK RESPONSE SAMPLE and the Task data table
// in the docs).
const (
	// Mail keys.
	keyMessageID = "messageId"
	// keyMessageIDString is the string twin of messageId. The numeric id exceeds
	// 2^53, so this lossless form is preferred wherever both are present.
	keyMessageIDString = "messageIdString"
	keyFolderID        = "folderId"
	keyReceivedTime    = "receivedTime"
	keyZuid            = "zuid"

	// Task keys.
	keyEntityID    = "entityId"    // task id
	keyAction      = "action"      // task action that triggered the webhook
	keyNamespaceID = "nameSpaceId" // group id the task belongs to
)

// Zoho CRM webhook discriminator keys — always present in CRM notification
// payloads and never in Zoho Mail outgoing-webhook payloads.
const (
	crmKeyAffectedValues = "affected_values"
	crmKeyModule         = "module"
	crmKeyOperation      = "operation"
)

// IsWebhookPayload reports whether a decoded webhook body is a Zoho Mail-module
// payload (Mail or Task), as opposed to a Zoho CRM one. The shapes are disjoint:
// CRM payloads carry "affected_values"/"module"/"operation", Mail carries
// "messageId", Task carries "entityId". The CRM discriminator keys must also be
// absent, so a CRM payload that ever grows a "messageId"/"entityId" field is
// not misrouted to the Mail parser.
func IsWebhookPayload(m map[string]any) bool {
	if !isMailPayload(m) && !isTaskPayload(m) {
		return false
	}

	for _, crmKey := range []string{crmKeyAffectedValues, crmKeyModule, crmKeyOperation} {
		if _, ok := m[crmKey]; ok {
			return false
		}
	}

	return true
}

func isMailPayload(m map[string]any) bool {
	_, ok := m[keyMessageID]

	return ok
}

func isTaskPayload(m map[string]any) bool {
	_, ok := m[keyEntityID]

	return ok
}

// CollapsedSubscriptionEvent is a Zoho Mail-module webhook payload. Zoho Mail
// delivers one entity per webhook request, so there is no fan-out.
type CollapsedSubscriptionEvent map[string]any

// SubscriptionEvent is a single Zoho Mail-module webhook event (Mail or Task).
type SubscriptionEvent map[string]any

var (
	_ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}
	_ common.SubscriptionEvent          = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent    = SubscriptionEvent{}
)

func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	return []common.SubscriptionEvent{SubscriptionEvent(e)}, nil
}

func (evt SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(evt), nil
}

// PreLoadData is a no-op: Zoho Mail webhooks carry the full entity inline.
func (evt SubscriptionEvent) PreLoadData(_ *common.SubscriptionEventPreLoadData) error {
	return nil
}

// EventType maps the event to a normalized type. Mail webhooks only fire on new
// mail, so they are always Create. Task webhooks carry an "action"; Zoho does
// not document its enum, so it is mapped best-effort (see taskEventType) with
// the raw value preserved by RawEventName.
func (evt SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	if evt.isTask() {
		return taskEventType(numberToString(evt[keyAction])), nil
	}

	return common.SubscriptionEventTypeCreate, nil
}

func (evt SubscriptionEvent) RawEventName() (string, error) {
	if evt.isTask() {
		return numberToString(evt[keyAction]), nil
	}

	return rawNameNewMail, nil
}

func (evt SubscriptionEvent) ObjectName() (string, error) {
	if evt.isTask() {
		return objectNameTasks, nil
	}

	return objectNameMessages, nil
}

// RecordId returns a composite "<parent>/<child>" identifier that carries the
// extra id GetRecordsByIds needs to fetch the record:
//   - Mail: "<folderId>/<messageId>" (the get-message endpoint also needs the
//     folder). If the payload has no folderId, the bare messageId is returned;
//     it still identifies the event, but GetRecordsByIds cannot fetch it — the
//     fetch requires the composite. Real Mail deliveries always carry folderId.
//   - Task: "<groupId>/<taskId>" for a group task, or the bare taskId for a
//     personal task (no group).
//
// Ids are large 64-bit integers read as strings to avoid float precision loss.
// The message id prefers the payload's lossless "messageIdString" twin;
// folderId has no such twin, so callers decoding the webhook body must use
// json.Decoder.UseNumber to keep it exact.
func (evt SubscriptionEvent) RecordId() (string, error) {
	if evt.isTask() {
		return joinCompositeID(numberToString(evt[keyNamespaceID]), numberToString(evt[keyEntityID])), nil
	}

	return joinCompositeID(numberToString(evt[keyFolderID]), evt.messageID()), nil
}

// joinCompositeID builds "<parent>/<child>", or just child when parent is empty.
func joinCompositeID(parent, child string) string {
	if parent == "" {
		return child
	}

	return parent + recordIDSeparator + child
}

// EventTimeStampNano returns the mail receivedTime (epoch ms) in nanoseconds.
// Task webhook payloads carry no trigger timestamp, so they return 0.
func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	if evt.isTask() {
		return 0, nil
	}

	ms, err := strconv.ParseInt(numberToString(evt[keyReceivedTime]), 10, 64)
	if err != nil {
		return 0, nil //nolint:nilerr // absent/odd timestamp is not fatal for event routing
	}

	return time.UnixMilli(ms).UnixNano(), nil
}

// Workspace returns the mailbox owner's zuid for mail, or the task's group id
// (nameSpaceId) for tasks; "" when absent.
func (evt SubscriptionEvent) Workspace() (string, error) {
	if evt.isTask() {
		return numberToString(evt[keyNamespaceID]), nil
	}

	return numberToString(evt[keyZuid]), nil
}

// UpdatedFields returns nil: neither a new-mail nor a task event carries a
// field-level change set.
func (evt SubscriptionEvent) UpdatedFields() ([]string, error) {
	return nil, nil
}

func (evt SubscriptionEvent) isTask() bool {
	return isTaskPayload(evt)
}

// messageID prefers the lossless messageIdString over the numeric messageId,
// which a plain json.Unmarshal (no UseNumber) decodes as float64 and rounds
// past 2^53.
func (evt SubscriptionEvent) messageID() string {
	if s := numberToString(evt[keyMessageIDString]); s != "" {
		return s
	}

	return numberToString(evt[keyMessageID])
}

// taskEventType maps a Zoho Mail task "action" to a normalized event type. The
// action enum is undocumented, so this matches on substrings and falls back to
// Other; the raw action is always available via RawEventName.
func taskEventType(action string) common.SubscriptionEventType {
	a := strings.ToLower(action)

	switch {
	case strings.Contains(a, "add"), strings.Contains(a, "create"):
		return common.SubscriptionEventTypeCreate
	case strings.Contains(a, "delet"), strings.Contains(a, "remov"), strings.Contains(a, "trash"):
		return common.SubscriptionEventTypeDelete
	case strings.Contains(a, "updat"), strings.Contains(a, "edit"), strings.Contains(a, "modif"),
		strings.Contains(a, "complet"), strings.Contains(a, "status"), strings.Contains(a, "assign"),
		strings.Contains(a, "move"), strings.Contains(a, "reopen"), strings.Contains(a, "close"):
		return common.SubscriptionEventTypeUpdate
	default:
		return common.SubscriptionEventTypeOther
	}
}

// numberToString renders a JSON scalar (string, json.Number, or float64) as a
// string without losing integer precision or using scientific notation.
func numberToString(v any) string {
	switch n := v.(type) {
	case nil:
		return ""
	case string:
		return n
	case json.Number:
		return n.String()
	case float64:
		return strconv.FormatFloat(n, 'f', -1, 64)
	default:
		return ""
	}
}
