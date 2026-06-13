# PR 2 — Verification (`WebhookVerifierConnector`)

> Part of the [Subscribe Onboarding PR Process](../../SUBSCRIBE_PR_PROCESS.md). Implementation
> reference: [`SUBSCRIBE_ONBOARDING.md`](../../SUBSCRIBE_ONBOARDING.md).

**Required.** Stacks on [PR 1](./pr-1-metadata-and-factory.md).

## Goal

Implement webhook signature verification and the typed events the caller parses out of incoming
webhooks.

## What you implement

- `WebhookVerifierConnector.VerifyWebhookMessage` on your `*Connector`.
- A provider-specific `VerificationParams` struct (the caller fills it in per installation).
- One or more event types implementing `common.SubscriptionEvent` (and
  `SubscriptionUpdateEvent` / `CollapsedSubscriptionEvent` where applicable).

## Files

- `providers/<provider>/subscribeEvent.go` (or similarly named) — verification + event types.

## Steps

1. Add the compile-time assertion: `var _ connectors.WebhookVerifierConnector = &Connector{}`.
2. Implement `VerifyWebhookMessage`: pull params via `common.AssertType`, read the signature header,
   recompute over `req.Body`, compare with `hmac.Equal`. Return `false` (not error) for untrusted
   requests.
3. Implement the event type's methods. `PreLoadData` runs first and receives the request headers/body —
   stash anything the other methods need.

## Example

```go
type SubscriptionEvent map[string]any

type AcmeVerificationParams struct {
    Secret string `json:"secret,omitempty"`
}

var (
    _ connectors.WebhookVerifierConnector = &Connector{}
    _ common.SubscriptionEvent            = SubscriptionEvent{}
)

func (c *Connector) VerifyWebhookMessage(
    ctx context.Context, req *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
    vp, err := common.AssertType[*AcmeVerificationParams](params.Param)
    if err != nil {
        return false, err
    }
    sig := req.Headers.Get("X-Acme-Signature")
    if sig == "" {
        return false, ErrMissingSignature
    }
    expected := hex.EncodeToString(hmacSHA256(vp.Secret, req.Body))
    return hmac.Equal([]byte(sig), []byte(expected)), nil
}

func (e SubscriptionEvent) EventType() (common.SubscriptionEventType, error) { /* ... */ }
func (e SubscriptionEvent) ObjectName() (string, error)                      { /* ... */ }
func (e SubscriptionEvent) RecordId() (string, error)                        { /* ... */ }
// ...RawEventName, Workspace, EventTimeStampNano, RawMap, PreLoadData
```

See [`providers/salesloft/subscribeEvent.go`](../../providers/salesloft/subscribeEvent.go) for a
complete, readable implementation.

## Checklist

- [ ] `var _ connectors.WebhookVerifierConnector = &Connector{}` assertion present.
- [ ] `VerifyWebhookMessage` uses constant-time compare (`hmac.Equal`); returns `false` (not an error)
      for untrusted requests.
- [ ] All `SubscriptionEvent` methods implemented; `SubscriptionUpdateEvent` /
      `CollapsedSubscriptionEvent` added where the provider needs them.
- [ ] Unit tests cover valid / invalid / missing-signature, plus each event method against a captured
      real payload.

## Reviewer focus

- Signature algorithm matches the provider's docs (header name, hash, encoding).
- Untrusted requests return `(false, nil)` rather than erroring.
- Event mapping (object name, record id, event type) is correct for real payloads.

## Reference

- [Verification](../../SUBSCRIBE_ONBOARDING.md#verification)
- [Event types](../../SUBSCRIBE_ONBOARDING.md#event-types)
