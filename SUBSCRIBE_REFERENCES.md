# Subscribe Action — Reference

This is the **reference** for adding **Subscribe** support (webhook subscriptions) to a provider in the
`github.com/amp-labs/connectors` library — the shared concepts that apply across the whole effort:
the interface ladder, the core types, and the shapes real providers take.

The work itself is shipped as a stack of small PRs. The detailed, step-by-step implementation guidance
lives in the per-PR guides — see [Writing the PRs](#writing-the-prs) at the bottom.

> **Scope.** This guide is about implementing Subscribe *in the connector*. The connector owns the
> provider implementation (subscribe/verify/registration logic) and the provider metadata. A separate
> orchestration layer in Ampersand's server — the *caller* of this connector — consumes it: it builds
> per-installation request payloads, persists results, schedules maintenance, and routes incoming
> webhook events to your event types. You implement the connector pieces described here; the caller's
> wiring is handled outside this repo.
>
> **Salesloft, Outreach, and Salesforce are the reference implementations** to model your work on.

---

## The big picture

A "subscribe-capable" connector does up to four things, each behind its own interface. The interfaces
build on one another, so you only implement the rungs your provider needs. `RegisterSubscribeConnector`
and `SubscriptionMaintainerConnector` each extend `SubscribeConnector` **independently** (neither
extends the other) — a provider can implement either, both, or neither.

```
RegisterSubscribeConnector          SubscriptionMaintainerConnector
(one-time setup, e.g. Salesforce    (renews expiring subscriptions)
 → EventBridge)
   provider-specific, if needed        provider-specific, if needed
            ▲                                  ▲
            └─────────── both extend ──────────┘
                            │
              SubscribeConnector            (create / update / delete subscriptions)
                            ▲ extends
              WebhookVerifierConnector       (verify incoming webhook signatures)
                            ▲ extends
              Connector + BatchRecordReaderConnector
                                  (base client + fetch records by id for webhook enrichment)
```

All four interfaces are defined in [`connectors.go`](./connectors.go). Note that
`WebhookVerifierConnector` embeds `Connector` (base HTTP client + provider identity) **and**
`BatchRecordReaderConnector` (`GetRecordsByIds`). This does **not** require a full read connector
(no `Read`/pagination) — only `GetRecordsByIds`, which the caller uses to fetch a record's full state
by id after a webhook event arrives.

This ladder is about *interface embedding*, **not PR/build order**. Although `RegisterSubscribeConnector`
embeds `SubscribeConnector`, registration is *sequenced before* subscribe in the PR stack because the
registration result is an input to `Subscribe` — see [the PR stack](./CONTRIBUTING_SUBSCRIBE_ACTION.md#the-stack).

| Interface | Methods you add | When you need it | PR |
|-----------|-----------------|------------------|----|
| `WebhookVerifierConnector` | `VerifyWebhookMessage` | Provider signs its webhooks (almost always). | [PR&nbsp;2](./docs/subscribe-onboarding/pr-2-verification.md) |
| `RegisterSubscribeConnector` | `Register`, `DeleteRegistration`, `EmptyRegistrationParams`, `EmptyRegistrationResult` | **Provider-specific, if needed** — only when the provider needs a one-time, installation-level setup shared by all object subscriptions (lands *before* Subscribe). | [PR&nbsp;3](./docs/subscribe-onboarding/pr-3-registration.md) |
| `SubscribeConnector` | `Subscribe`, `UpdateSubscription`, `DeleteSubscription`, `EmptySubscriptionParams`, `EmptySubscriptionResult` | Provider lets you create subscriptions programmatically via API. | [PR&nbsp;4](./docs/subscribe-onboarding/pr-4-subscribe-update-delete.md) |
| `SubscriptionMaintainerConnector` | `RunScheduledMaintenance` | **Provider-specific, if needed** — only when subscriptions/watches expire after a TTL and must be renewed on a schedule. | [PR&nbsp;5](./docs/subscribe-onboarding/pr-5-maintenance.md) |

### How the caller uses `ProviderInfo`

The interfaces above are the *how* — the code that talks to the provider. **`ProviderInfo` is what the
caller reads to know *whether* and *how* to use them.** Alongside implementing the interfaces, every
provider declares subscribe metadata in its `ProviderInfo` (in `providers/<provider>.go`):

- **`Support.Subscribe`** — the master switch: does this provider support subscribe at all?
- **`SubscribeRequirements`** — the shape of that support: `SubscribeByAPI` (subscribe via API vs.
  manual UI configuration), and whether the provider needs `Registration`, `PostProcess`, or
  `Maintenance`.

At runtime the caller reads this metadata to pick the path (API vs. manual) and to decide which of the
optional steps (registration, post-process, scheduled maintenance) to run — it never hard-codes
per-provider behavior. That's why declaring `ProviderInfo` is the first PR ([PR&nbsp;1](./docs/subscribe-onboarding/pr-1-provider-info.md))
and why a provider stays dormant until those flags are flipped on. The fields are detailed in
[Core types](#core-types) and [PR&nbsp;1](./docs/subscribe-onboarding/pr-1-provider-info.md).

A "UI Subscription only" provider (subscriptions configured in the provider's own UI, e.g. Hubspot/Gong)
only implements `WebhookVerifierConnector` — the caller reads its events but never calls `Subscribe`.

---

## Core types

From [`common/types.go`](./common/types.go). These are the provider-agnostic payloads the framework
passes you; the `any` fields are where your provider-specific structs live. They're shared across every
PR below, so they live here in the overview.

```go
type SubscribeParams struct {
    Request            any                          // your provider-specific request struct
    RegistrationResult *RegistrationResult          // set only for providers that Register first
    SubscriptionEvents map[ObjectName]ObjectEvents  // normalized desired state (objects → events)
}

type SubscriptionResult struct {
    Result       any                          // your provider's raw response (preserved verbatim)
    ObjectEvents map[ObjectName]ObjectEvents  // normalized actual state after the operation
    Status       SubscriptionStatus
    // Objects/Events/UpdateFields/PassThroughEvents are DEPRECATED — use ObjectEvents.
}

type ObjectEvents struct {
    Events            SubscriptionEventTypes  // create / update / delete
    WatchFields       []string                // fields to watch on update
    WatchFieldsAll    bool                    // watch all fields (provider-specific quirk)
    PassThroughEvents []string                // provider-specific event names
}

type VerificationParams struct{ Param any }   // wraps your provider-specific verification struct

type WebhookRequest struct {
    Headers http.Header
    Body    []byte
    URL     string
    Method  string
}

type SubscriptionRegistrationParams struct{ Request any }

type RegistrationResult struct {
    RegistrationRef string
    Result          any                  // your provider-specific registration data
    Status          RegistrationStatus
}
```

Event type constants (`common/types.go`):

```go
SubscriptionEventTypeCreate = "create"
SubscriptionEventTypeUpdate = "update"
SubscriptionEventTypeDelete = "delete"
SubscriptionEventTypeOther  = "other"
```

---

## Real providers, different shapes

| Provider | Interfaces implemented | Notes |
|----------|------------------------|-------|
| **Hubspot / Gong** | `WebhookVerifierConnector` only | UI Subscription only: subscriptions configured in the provider UI. The caller reads events; never calls `Subscribe`. |
| **Outreach / Salesloft** | `SubscribeConnector` | API subscribe + verify. Parallel create with rollback. Salesloft also needs the `WatchFieldsAuto="all"` quirk (set by the caller). |
| **Salesforce** | `RegisterSubscribeConnector` | Registers EventChannel/NamedCredential/EventRelayConfig, then subscribes PlatformEventChannelMembers. Needs `PostProcess` (AWS EventBridge, done outside the connector). |

---

## PostProcess

PostProcess is needed when subscribing requires a **second set of credentials** that the connector's
token doesn't hold. The connector authenticates to *its* provider (say, Salesforce), but the setup step
needs access to a **different** system (e.g. AWS EventBridge) — credentials the connector can't supply.
Whenever completing subscribe requires extra credentials / access beyond the connector's token, that
step is **PostProcess**.

How it plays out:

- The connector **does not implement PostProcess** — there is no connector method or interface for it.
  The post-process logic lives in **server-side code** (Ampersand's platform), which holds the
  credentials/access for the other system.
- The connector's only contribution is to **declare it** — set `SubscribeRequirements.PostProcess: new(true)`
  in `ProviderInfo` — and to **return any data the post-process step needs** in its `Subscribe` /
  `Register` result (e.g. an id created during registration), so the server side can use it.

> **Consult Ampersand staff before you start.** PostProcess spans two systems and has no standard
> connector pattern. If your provider needs it, flag it to Ampersand staff **ahead of time** so they
> can advise how to proceed — the post-process work itself is done on the server side, and there may be
> no connector-side PR for it at all.

---

## Writing the PRs

Ship subscribe support as a stack of small, gated-off PRs. The process — principles, the stack diagram,
merge order, gating, and how to manage the stack — is in
[**`CONTRIBUTING_SUBSCRIBE_ACTION.md`**](./CONTRIBUTING_SUBSCRIBE_ACTION.md).

Each PR has its own focused guide with the full implementation detail for that step (interface
signatures, examples, files, checklist, reviewer focus):

| # | PR | Guide | Required? |
|---|----|-------|-----------|
| 1 | ProviderInfo + Factory wiring | [pr-1-provider-info.md](./docs/subscribe-onboarding/pr-1-provider-info.md) | ✅ |
| 2 | Verification | [pr-2-verification.md](./docs/subscribe-onboarding/pr-2-verification.md) | ✅ |
| 3 | Registration | [pr-3-registration.md](./docs/subscribe-onboarding/pr-3-registration.md) | ⬜ if needed |
| 4 | Subscribe / Update / Delete | [pr-4-subscribe-update-delete.md](./docs/subscribe-onboarding/pr-4-subscribe-update-delete.md) | ✅ |
| 5 | Maintenance | [pr-5-maintenance.md](./docs/subscribe-onboarding/pr-5-maintenance.md) | ⬜ if needed |
| 6 | Enable the provider | [pr-6-enable.md](./docs/subscribe-onboarding/pr-6-enable.md) | ✅ (last) |
