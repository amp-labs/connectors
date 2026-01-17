package salesloft

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
)

type SubscriptionEvent map[string]any

type SalesloftVerificationParams struct {
	Secret string `json:"secret,omitempty"`
}

var (
	_ common.SubscriptionEvent       = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent = SubscriptionEvent{}
)

var (
	signatureHeader = "x-salesloft-signature" //nolint:gochecknoglobals
	eventHeader     = "x-salesloft-event"     //nolint:gochecknoglobals,unused
)

func (c *Connector) VerifyWebhookMessage(ctx context.Context,
	req *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
	if req == nil || params == nil {
		return false, fmt.Errorf("%w: request and params cannot be nil", errMissingParams)
	}

	verificationParams, err := common.AssertType[*SalesloftVerificationParams](params.Param)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errMissingParams, err)
	}

	signature := req.Headers.Get(signatureHeader)

	if signature == "" {
		return false, fmt.Errorf("%w: missing %s header", ErrMissingSignature, signatureHeader)
	}

	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("%w: invalid signature format", ErrInvalidSignature)
	}

	expectedSignature := computeSignature(verificationParams.Secret, req.Body)

	if !hmac.Equal(sigBytes, expectedSignature) {
		return false, fmt.Errorf("%w: signature mismatch", ErrInvalidSignature)
	}

	return true, nil
}

func (evt SubscriptionEvent) UpdatedFields() ([]string, error) {
	// Salesloft does not provide updated fields in webhook response.
	return []string{}, errors.New("updated fields are not supported by Salesloft webhooks") //nolint:err113
}

// nolint: godoclint
// Salesloft doesn't provide deliveredAt field in webhook response.
// So we are using updated_at field as event timestamp.
func (evt SubscriptionEvent) EventTimeStampNano() (int64, error) {
	m := evt.asMap()

	updatedAtStr, err := m.GetString("updated_at")
	if err != nil {
		return 0, err
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtStr)
	if err != nil {
		return 0, fmt.Errorf("error parsing deliveredAt time: %w", err)
	}

	return updatedAt.UnixNano(), nil
}

func (evt SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	return common.SubscriptionEventType(""), errors.New("event type not provided by Salesloft webhooks") //nolint:err113
}

func (evt SubscriptionEvent) ObjectName() (string, error) {
	return "", errors.New("object name not provided by Salesloft webhooks") //nolint:err113
}

func (evt SubscriptionEvent) RawEventName() (string, error) {
	return "", errors.New("raw event name not provided by Salesloft webhooks") //nolint:err113
}

func (evt SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(evt), nil
}

func (evt SubscriptionEvent) RecordId() (string, error) {
	m := evt.asMap()

	id, err := m.GetInt("id")
	if err != nil {
		return "", err
	}

	return strconv.Itoa(int(id)), nil
}

// Workspace is not available in Salesloft.
func (evt SubscriptionEvent) Workspace() (string, error) {
	return "", nil
}

func (evt SubscriptionEvent) asMap() common.StringMap {
	return common.StringMap(evt)
}

// computeSignature returns the raw HMAC digest bytes using SHA1 (Salesloft docs).
func computeSignature(secret string, body []byte) []byte {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write(body)

	return h.Sum(nil)
}

// Example: Webhook Response
/*
{
  "_integration_id": null,
  "_integration_name": null,
  "_integration_step_id": null,
  "_integration_task_id": null,
  "_integration_task_type_label": null,
  "completed_at": null,
  "completed_by": null,
  "created_at": "2025-11-20T07:21:04.176528-05:00",
  "created_by_user": {
    "_href": "https://api.salesloft.com/v2/users/49067",
    "id": 49067
  },
  "current_state": "scheduled",
  "custom_attribute_resources": {},
  "custom_attributes": {},
  "description": "noted updated ",
  "due_at": null,
  "due_date": "2025-11-27",
  "expires_after": null,
  "id": 693873531,
  "instigator": {
    "action_caller_id": 49067,
    "action_caller_name": "Int User",
    "metadata": {},
    "reason": "api",
    "type": "manual",
    "user_guid": "0863ed13-7120-479b-8650-206a3679e2fb"
  },
  "multitouch_group_id": null,
  "object_references": [],
  "person": {
    "_href": "https://api.salesloft.com/v2/people/436664215",
    "id": 436664215
  },
  "remind_at": null,
  "reminded": false,
  "rollback_reason": null,
  "score": {
    "factors": {},
    "prioritizer_uuid": "salesloft.prioritizers/rhythm",
    "score": "2.0"
  },
  "source": "salesloft.api",
  "subject": "Follow-up with John Kelly",
  "task_type": "general",
  "updated_at": "2025-11-20T07:21:16.193200-05:00",
  "user": {
    "_href": "https://api.salesloft.com/v2/users/49067",
    "id": 49067
  }
}

*/
