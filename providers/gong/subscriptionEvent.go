package gong

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strconv"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// Gong's automation rules currently only supports firing webhook for Call creation events.
// https://help.gong.io/docs/create-a-webhook-rule
const (
	gongCallCreatedEvent = "callCreated"
	gongCallObject       = "Call"
)

var (
	_ connectors.WebhookVerifierConnector = &Connector{}
	_ common.SubscriptionEvent            = SubscriptionEvent{}
	_ common.CollapsedSubscriptionEvent   = CollapsedSubscriptionEvent{}
)

var (
	errTypeMismatch = errors.New("type mismatch")
	errMissingField = errors.New("missing field")
)

// VerifyWebhookMessage is a stub that always succeeds
// TODO: implement verification logic for Gong webhooks.
func (*Connector) VerifyWebhookMessage(
	_ context.Context,
	_ *common.WebhookRequest,
	_ *common.VerificationParams,
) (bool, error) {
	return true, nil
}

// CollapsedSubscriptionEvent represents the raw webhook payload from Gong.
// Gong sends one call per webhook, so this simply wraps the single event.
type CollapsedSubscriptionEvent map[string]any

func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

// SubscriptionEventList returns the event as a single-element list.
// Gong webhooks contain only one record per payload, so no fan-out is needed.
// https://help.gong.io/docs/payload-sent-to-webhooks
func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	return []common.SubscriptionEvent{SubscriptionEvent(e)}, nil
}

// SubscriptionEvent holds a single Gong webhook event. Gong's "Subscribe"
// webhook rule only fires on call creation.
// See https://help.gong.io/docs/payload-sent-to-webhooks for the payload spec.
//
// Example Gong webhook payload:
//
//	{
//	  "callData": {
//	    "metaData": {
//	      "id": "7782342341192847895",
//	      "url": "https://app.gong.io/call?id=7782342341192847895",
//	      "title": "Acme <> Gong | Intro",
//	      "scheduled": "2024-02-28T15:00:00Z",
//	      "started":   "2024-02-28T15:02:11Z",
//	      "duration":  1847,
//	      "primaryUserId": "234599920309",
//	      "direction": "Outbound",
//	      "system":    "Zoom",
//	      "scope":     "External",
//	      "media":     "Video",
//	      "language":  "eng",
//	      "workspaceId": "3498573645"
//	    },
//	    "parties":  [ ... ],
//	    "content":  { ... }
//	  },
//	  "isPrivate": false,
//	  "isTest":    false
//	}
type SubscriptionEvent map[string]any

// PreLoadData is a no-op: Gong webhooks include the full call payload inline,
// so there's no external data to hydrate.
func (evt SubscriptionEvent) PreLoadData(_ *common.SubscriptionEventPreLoadData) error {
	return nil
}

func (evt SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(evt), nil
}

// EventType returns Create when the payload includes callData; otherwise Other.
// Gong currently only supports Subscribe for Call creation events.
func (evt SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	m := common.StringMap(evt)

	if _, err := m.Get("callData"); err != nil {
		return common.SubscriptionEventTypeOther, nil //nolint:nilerr
	}

	return common.SubscriptionEventTypeCreate, nil
}

// RawEventName returns a hard-coded string because Gong's payload does not carry a raw event name.
func (evt SubscriptionEvent) RawEventName() (string, error) {
	return gongCallCreatedEvent, nil
}

// ObjectName is always "Call" for Gong subscribe webhooks.
func (evt SubscriptionEvent) ObjectName() (string, error) {
	return gongCallObject, nil
}

func (evt SubscriptionEvent) Workspace() (string, error) {
	meta, err := evt.callMetaData()
	if err != nil {
		return "", err
	}

	return meta.GetString("workspaceId")
}

// RecordId returns the Gong call ID from callData.metaData.id.
// Accepts both string (per Gong's docs) and float64.
func (evt SubscriptionEvent) RecordId() (string, error) {
	meta, err := evt.callMetaData()
	if err != nil {
		return "", err
	}

	callIDValue, err := meta.Get("id")
	if err != nil {
		return "", err
	}

	switch v := callIDValue.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("%w: expected string or number for callData.metaData.id, got %T", errTypeMismatch, callIDValue)
	}
}

// EventTimeStampNano returns 0: Gong's webhook payload does not include a
// delivery timestamp, and the call timing fields (scheduled/started) are not
// semantically the event time.
func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	return 0, nil
}

// callMetaData returns the callData.metaData sub-object, where Gong puts
// the call identity, workspace, and timing fields.
func (evt SubscriptionEvent) callMetaData() (common.StringMap, error) {
	callData, err := common.StringMap(evt).Get("callData")
	if err != nil {
		return nil, err
	}

	callMap, callMapOK := callData.(map[string]any)
	if !callMapOK {
		return nil, fmt.Errorf("%w: expected callData object, got %T", errTypeMismatch, callData)
	}

	metaData, hasMetaData := callMap["metaData"]
	if !hasMetaData {
		return nil, fmt.Errorf("%w: callData.metaData", errMissingField)
	}

	metaMap, metaMapOK := metaData.(map[string]any)
	if !metaMapOK {
		return nil, fmt.Errorf("%w: expected callData.metaData object, got %T", errTypeMismatch, metaData)
	}

	return common.StringMap(metaMap), nil
}
