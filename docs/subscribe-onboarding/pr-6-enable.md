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
    // Registration / PostProcess / Maintenance already set earlier in the stack
},
```

For a **UI Subscription only** provider (no API subscribe), set `Support.Subscribe: true` and leave
`SubscribeByAPI` off.

## Files

- `providers/<provider>.go` — the activation flags only.

## Live testing (required)

Because this PR turns the provider on, it must be verified **live on Ampersand** — not just with the
local harness from PR 3.

1. Install the provider on **Ampersand**. The **MailMonkey demo app** is a convenient test app for this.
2. Create an installation and subscribe.
3. Trigger a change in the provider so it emits a **real webhook**, and confirm it is delivered
   end-to-end. To receive and inspect deliveries without standing up your own endpoint, use
   **[Svix Play](https://play.svix.com/)** (free) as the webhook receiver.
4. **Attach a screenshot of the actual delivered webhook to the PR.** This is required.

## Checklist

- [ ] All prerequisite PRs merged.
- [ ] Live-tested on **Ampersand**: installed (e.g. via the MailMonkey demo app), created an
      installation, subscribed, and received a **real webhook** end-to-end (Svix Play works as the
      receiver).
- [ ] **Screenshot of the actual delivered webhook attached to the PR.**
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
