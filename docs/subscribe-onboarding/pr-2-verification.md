# PR 2 — Verification (`WebhookVerifierConnector`)

> Part of the [Subscribe Onboarding PR Process](../../SUBSCRIBE_PR_PROCESS.md). Shared concepts:
> [`SUBSCRIBE_ONBOARDING.md`](../../SUBSCRIBE_ONBOARDING.md).

**Required.** Stacks on [PR 1](./pr-1-provider-info.md).

## Goal

Implement webhook signature verification and the typed events the caller parses out of incoming
webhooks.

## What you implement

`WebhookVerifierConnector` (defined in [`connectors.go`](../../connectors.go)):

```go
VerifyWebhookMessage(
    ctx context.Context,
    request *common.WebhookRequest,
    params *common.VerificationParams,
) (bool, error)
```

Return `true` to allow webhook processing, `false` to reject as untrusted, and an `error` only for
*unexpected* failures.

Plus a provider-specific `VerificationParams` struct (the caller fills it in per installation) and one
or more event types.

Files: `providers/<provider>/subscribeEvent.go` (or similarly named).

## Event types

When a webhook arrives, the caller casts the raw payload into your typed events and asks them
provider-agnostic questions (what object? what record id? create or update?). Implement these
interfaces (from `common/types.go`) on a provider event type:

```go
type Event interface {
    RawMap() (map[string]any, error)
}

type SubscriptionEvent interface {
    Event
    EventType() (SubscriptionEventType, error)
    RawEventName() (string, error)
    ObjectName() (string, error)
    Workspace() (string, error)
    RecordId() (string, error)
    EventTimeStampNano() (int64, error)
    PreLoadData(data *SubscriptionEventPreLoadData) error  // called first; receives request headers/body
}

type SubscriptionUpdateEvent interface {  // implement if the provider reports which fields changed
    SubscriptionEvent
    UpdatedFields() ([]string, error)
}

type CollapsedSubscriptionEvent interface {  // when one webhook payload fans out to N events
    Event
    SubscriptionEventList() ([]SubscriptionEvent, error)
}
```

Salesloft models all three with a `map[string]any`:

```go
type SubscriptionEvent map[string]any
type CollapsedSubscriptionEvent map[string]any

var (
    _ common.SubscriptionEvent          = SubscriptionEvent{}
    _ common.SubscriptionUpdateEvent    = SubscriptionEvent{}
    _ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}
)
```

- **`CollapsedSubscriptionEvent`** is the raw payload as it arrives. `SubscriptionEventList()` splits it
  into individual `SubscriptionEvent`s. Salesloft sends one record per webhook, so it returns a
  single-element slice; Salesforce/Zoho batch many events into one payload and fan out here.
- **`PreLoadData`** runs first for every event and is handed the request headers/body — use it to stash
  anything (e.g. an event-name header) the other methods need.

See [`providers/salesloft/subscribeEvent.go`](../../providers/salesloft/subscribeEvent.go) for a
complete, readable implementation.

## Verification

Implement `VerifyWebhookMessage` on your `*Connector`. The pattern (from
[`providers/salesloft/subscribeEvent.go`](../../providers/salesloft/subscribeEvent.go)):

1. Pull your provider-specific params out of `params.Param` with `common.AssertType`.
2. Read the signature header from `req.Headers`.
3. Recompute the expected signature over `req.Body` (and sometimes method/url/timestamp) with the
   shared secret, and compare with `hmac.Equal`.

```go
type SalesloftVerificationParams struct {
    Secret string `json:"secret,omitempty"`
}

var _ connectors.WebhookVerifierConnector = &Connector{}

func (c *Connector) VerifyWebhookMessage(ctx context.Context,
    req *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
    if req == nil || params == nil {
        return false, fmt.Errorf("%w: request and params cannot be nil", errMissingParams)
    }

    vp, err := common.AssertType[*SalesloftVerificationParams](params.Param)
    if err != nil {
        return false, fmt.Errorf("%w: %w", errMissingParams, err)
    }

    signature := req.Headers.Get("x-salesloft-signature")
    if signature == "" {
        return false, fmt.Errorf("%w: missing signature header", ErrMissingSignature)
    }

    sigBytes, err := hex.DecodeString(signature)
    if err != nil {
        return false, fmt.Errorf("%w: invalid signature format", ErrInvalidSignature)
    }

    if !hmac.Equal(sigBytes, computeSignature(vp.Secret, req.Body)) {
        return false, fmt.Errorf("%w: signature mismatch", ErrInvalidSignature)
    }
    return true, nil
}
```

The provider-specific `*VerificationParams` struct (here `SalesloftVerificationParams`) is populated
per installation by the caller and passed through `common.VerificationParams.Param`. Hubspot and
Outreach show HMAC-SHA256 variants; Salesloft uses HMAC-SHA1. For UI Subscription only providers whose
events carry no provider signature, verification is bypassed by the caller rather than implemented here
(still implement it if the provider does sign its webhooks).

## Checklist

- [ ] `var _ connectors.WebhookVerifierConnector = &Connector{}` assertion present.
- [ ] `VerifyWebhookMessage` uses constant-time compare (`hmac.Equal`); returns `false` (not an error)
      for untrusted requests.
- [ ] All `SubscriptionEvent` methods implemented; `SubscriptionUpdateEvent` /
      `CollapsedSubscriptionEvent` added where the provider needs them.
- [ ] Unit tests cover valid / invalid / missing-signature, plus each event method against a captured
      real payload (table-driven; see `test/utils/testroutines/`).

## Reviewer focus

- Signature algorithm matches the provider's docs (header name, hash, encoding).
- Untrusted requests return `(false, nil)` rather than erroring.
- Event mapping (object name, record id, event type) is correct for real payloads.

## Reference

- [Core types](../../SUBSCRIBE_ONBOARDING.md#core-types)
- [`providers/salesloft/subscribeEvent.go`](../../providers/salesloft/subscribeEvent.go)
