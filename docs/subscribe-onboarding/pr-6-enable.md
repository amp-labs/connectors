# PR 6 — Enable the provider *(top)*

> Part of the [Subscribe Onboarding PR Process](../../SUBSCRIBE_PR_PROCESS.md). Implementation
> reference: [`SUBSCRIBE_ONBOARDING.md`](../../SUBSCRIBE_ONBOARDING.md).

**Required, last.** The top of the stack — a one-line switch that turns the provider on.

## Goal

Flip the activation flags so the caller starts driving subscribe for this provider.

## Prerequisites

All prerequisite PRs are merged:

- [PR 2 — Verification](./pr-2-verification.md)
- [PR 3 — Subscribe / Update / Delete](./pr-3-subscribe-update-delete.md)
- [PR 4 — Registration](./pr-4-registration.md) *(only if the provider needs it)*
- [PR 5 — Maintenance](./pr-5-maintenance.md) *(only if the provider needs it)*

## What you implement

In `providers/<provider>.go`, flip the gate:

```go
Support: Support{
    Subscribe: true, // was false
},
SubscribeRequirements: &SubscribeRequirements{
    // <provider> supports creating webhook subscriptions via API: <link to provider docs>
    SubscribeByAPI: new(true), // was new(false)
    // Registration / PostProcess / Maintenance already set by their PRs
},
```

For a **UI Subscription only** provider (no API subscribe), set `Support.Subscribe: true` and leave
`SubscribeByAPI` off.

## Files

- `providers/<provider>.go` — the activation flags only.

## Checklist

- [ ] All prerequisite PRs merged.
- [ ] End-to-end verified in a sandbox (subscribe → receive webhook → verify).
- [ ] `SubscribeByAPI: new(true)` has a code comment linking the provider docs that justify it.
- [ ] Change is just the flag flip, trivial to revert.

## Reviewer focus

- Nothing else changes in this PR — it's purely the flag flip.
- The implementation it activates is already merged and tested.

> The caller also needs a small configuration change to consume the new provider — supplying the
> per-installation verification params and subscribe request payload your connector expects. That
> change lives outside this repo and lands after this stack.

## Reference

- [PR 1 — ProviderInfo + Factory wiring](./pr-1-provider-info.md) (the flags you're flipping)
- [Why gate, and on which flags](../../SUBSCRIBE_PR_PROCESS.md#why-gate-and-on-which-flags)
