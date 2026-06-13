# PR 5 — Maintenance (`SubscriptionMaintainerConnector`) *(provider-specific, if needed)*

> Part of the [Subscribe Onboarding PR Process](../../SUBSCRIBE_PR_PROCESS.md). Implementation
> reference: [`SUBSCRIBE_ONBOARDING.md`](../../SUBSCRIBE_ONBOARDING.md).

**Optional — skip this PR unless** the provider's subscriptions/watches expire after a TTL and must be
periodically renewed. Most providers do not.

Branches off [PR 3](./pr-3-subscribe-update-delete.md). Independent of
[PR 4 (Registration)](./pr-4-registration.md) — they don't depend on each other; do either, both, or
neither, in any order.

## Goal

Implement the scheduled renewal that keeps an expiring subscription alive.

## What you implement

`SubscriptionMaintainerConnector` on your `*Connector`:

```go
RunScheduledMaintenance(
    ctx context.Context,
    params common.SubscribeParams,
    previous *common.SubscriptionResult,
) (*common.SubscriptionResult, error)
```

And set `SubscribeRequirements.Maintenance: new(true)`.

## Files

- `providers/<provider>/...` — `RunScheduledMaintenance`.
- `providers/<provider>.go` — flip the `Maintenance` requirement flag.

## Steps

1. Add `var _ connectors.SubscriptionMaintainerConnector = &Connector{}`.
2. Implement `RunScheduledMaintenance`: renew/refresh the subscription described by `previous` (it
   carries provider-specific identifiers, timestamps, expiry) and return the **updated** state. `params`
   is typically identical to the currently active configuration.
3. Set `Maintenance: new(true)` in the provider metadata.

> The renewal **cadence** is configured by the caller, not in the connector.

## Checklist

- [ ] `var _ connectors.SubscriptionMaintainerConnector = &Connector{}` assertion present.
- [ ] `RunScheduledMaintenance` renews the subscription in `previous` and returns refreshed state.
- [ ] `Maintenance: new(true)` set in the provider metadata.

## Reviewer focus

- Returned `SubscriptionResult` carries the new expiry/identifiers so the next renewal works.
- Idempotent / safe to run repeatedly.

## Reference

- [The big picture](../../SUBSCRIBE_ONBOARDING.md#the-big-picture) · [Core types](../../SUBSCRIBE_ONBOARDING.md#core-types)
