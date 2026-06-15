# PR 5 — Maintenance (`SubscriptionMaintainerConnector`) *(provider-specific, if needed)*

> Part of [Contributing a Subscribe Action](../../CONTRIBUTING_SUBSCRIBE_ACTION.md). Shared concepts:
> [`SUBSCRIBE_REFERENCES.md`](../../SUBSCRIBE_REFERENCES.md).

**Optional — skip this PR unless** the provider's subscriptions/watches expire after a TTL and must be
periodically renewed. Most providers do not.

Builds on [PR 3](./pr-3-subscribe-update-delete.md).

## Goal

Implement the scheduled renewal that keeps an expiring subscription alive.

## What you implement

`SubscriptionMaintainerConnector` on your `*Connector`:

```go
RunScheduledMaintenance(
    ctx context.Context,
    params common.SubscribeParams,
    previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error)
```

And set `SubscribeRequirements.Maintenance: new(true)`.

## Files

- `providers/<provider>/...` — `RunScheduledMaintenance`.
- `providers/<provider>.go` — flip the `Maintenance` requirement flag.

## Steps

1. Add `var _ connectors.SubscriptionMaintainerConnector = &Connector{}`.
2. Implement `RunScheduledMaintenance`: renew/refresh the subscription described by `previousResult`
   (it carries provider-specific identifiers, timestamps, expiry) and return the **updated** state.
   `params` is typically identical to the currently active configuration.
3. Set `Maintenance: new(true)` in the provider metadata.

> The renewal **cadence** is configured by the caller, not in the connector.

## Checklist

- [ ] `var _ connectors.SubscriptionMaintainerConnector = &Connector{}` assertion present.
- [ ] `RunScheduledMaintenance` renews the subscription in `previousResult` and returns refreshed state.
- [ ] `Maintenance: new(true)` set in the provider metadata, with a code comment linking the provider
      docs that justify it (e.g. the subscription TTL / renewal cadence).

## Reviewer focus

- Returned `SubscriptionResult` carries the new expiry/identifiers so the next renewal works.
- Idempotent / safe to run repeatedly.

## Reference

- [The big picture](../../SUBSCRIBE_REFERENCES.md#the-big-picture) · [Core types](../../SUBSCRIBE_REFERENCES.md#core-types)
