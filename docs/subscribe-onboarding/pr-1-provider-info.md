# PR 1 — `ProviderInfo` + Factory wiring *(base)*

> Part of [Contributing a Subscribe Action](../../CONTRIBUTING_SUBSCRIBE_ACTION.md). Shared concepts:
> [`SUBSCRIBE_REFERENCES.md`](../../SUBSCRIBE_REFERENCES.md).

**Required.** This is the base of the stack — everything else stacks on it.

## Goal

Declare the provider's subscribe metadata on its `ProviderInfo` — `Support.Subscribe` and
`SubscribeRequirements` — with the gate (`Support.Subscribe`) **off**, and wire the connector into the
factory if it's brand-new. This PR changes **no runtime behavior**; it's a safe no-op until the final
[`Enable`](./pr-6-enable.md) PR.

## What you implement

Declare subscribe capability in the provider's handwritten info file, `providers/<provider>.go`,
inside `SetInfo(...)`. Two pieces matter:

- **`Support.Subscribe`** — the master "this provider supports subscribe at all" switch.
- **`SubscribeRequirements`** — what the provider *needs* in order to subscribe. All fields are `*bool`.
  From `providers/types.gen.go`:

```go
type SubscribeRequirements struct {
    // Maintenance: subscriptions/watches expire after a TTL and must be renewed on a schedule.
    Maintenance *bool
    // PostProcess: subscribing needs a third-party setup step the connector itself cannot perform
    // (e.g. Salesforce → AWS EventBridge).
    PostProcess *bool
    // Registration: a one-time setup step shared across all subscribed objects is required.
    Registration *bool
    // SubscribeByAPI: subscriptions can be created programmatically via API. If false, the provider
    // may still support webhooks via manual configuration in its UI (UI Subscription only).
    SubscribeByAPI *bool
}
```

In **this** PR, keep `Support.Subscribe` off — it's the gate (see [Gating rule](#gating-rule) below).
Declare the rest of `SubscribeRequirements` at its real value; with the gate off the provider stays
dormant regardless:

```go
// providers/<provider>.go — gated OFF (Support.Subscribe is the gate)
Support: Support{
    Read:      true,
    Write:     true,
    Subscribe: false, // ← the gate; flip to true in PR 6
},
SubscribeRequirements: &SubscribeRequirements{
    // <provider> supports creating webhook subscriptions via API: <link to provider docs>
    SubscribeByAPI: new(true), // the provider's real capability; omit/false for UI-Subscription-only
    // Registration / PostProcess / Maintenance: new(true) only if the provider will need them
},
```

Set `Registration` / `PostProcess` / `Maintenance` to `new(true)` **only if** the provider needs them.
These requirement flags are harmless while gated off — they're only consulted once subscribe is active.
`Registration` and `Maintenance` each pair with a connector interface implemented in its own PR
([PR 3](./pr-3-registration.md) / [PR 5](./pr-5-maintenance.md)). **`PostProcess` has no connector code
at all** — it's a pure indicator declared here; see [PostProcess](#postprocess-indicator-only) below.

> **Always link the provider docs.** Whenever you set a `SubscribeRequirements` flag to `new(true)`,
> precede it with a code comment linking the provider documentation that establishes it — that the
> provider supports API subscriptions, requires registration, needs a post-process setup step, or
> expires subscriptions on a schedule. Reviewers rely on these links to verify each flag, so a PR that
> adds a flag without a doc link should not pass review.

### PostProcess (indicator only)

`PostProcess` covers **any setup that must happen in a third-party system the connector has no access
to**. For a concrete, fully explained example see the `PostProcess` comment in
[`providers/salesforce.go`](../../providers/salesforce.go) (search for `PostProcess`). It is
**independent of registration**: Salesforce happens to need both, but a provider can require one without
the other.

The post-process logic itself lives in **server-side code**, not in the connector — there is **no
connector method or interface** for it. The connector's entire contribution is to **indicate whether it
is required**, by setting `SubscribeRequirements.PostProcess: new(true)` in `ProviderInfo` here (with
the doc-link comment above). If a post-process step will need data the connector produces (e.g. an id
created during [registration](./pr-3-registration.md) or returned by `Subscribe`), make sure that data
is returned in the corresponding result so the server side can consume it.

> **Consult Ampersand staff ahead of time** if your provider needs PostProcess — it spans two systems
> and the work is done server-side, so coordinate before you build. See
> [PostProcess in the reference](../../SUBSCRIBE_REFERENCES.md#postprocess).

### Examples from real providers

Salesloft (API subscribe, no registration/maintenance/post-process — shown here in its final enabled
form):

```go
// providers/salesloft.go
Support: Support{ Read: true, Write: true, Proxy: true, Subscribe: true },
SubscribeRequirements: &SubscribeRequirements{
    // Salesloft supports creating webhook subscriptions via API:
    // https://developers.salesloft.com/docs/platform/webhooks
    SubscribeByAPI: new(true),
},
```

Salesforce declares it **per module** (`ModuleSalesforceCRM` only) — note other modules keep
`Subscribe: false`:

```go
// providers/salesforce.go (within ModuleSalesforceCRM)
Support: Support{ Subscribe: true, /* ... */ },
SubscribeRequirements: &SubscribeRequirements{
    // One-time EventRelay/EventChannel setup per installation:
    // https://developer.salesforce.com/docs/platform/pub-sub-api/guide/event-relay-intro.html
    Registration:   new(true),
    // AWS EventBridge wiring after subscribe (done outside the connector):
    // https://developer.salesforce.com/docs/platform/pub-sub-api/guide/aws-event-relay.html
    PostProcess:    new(true),
    // Subscriptions created via the Salesforce API:
    // https://developer.salesforce.com/docs/platform/pub-sub-api/guide/intro.html
    SubscribeByAPI: new(true),
},
```

## Gating rule

> `Support.Subscribe` is the **gate** — it must be `true` for the provider to subscribe at all (via API
> or manual/UI); `SubscribeByAPI` says whether the programmatic API approach is available. **Keep
> `Support.Subscribe` off** for the entire stack and flip it on only in the final
> [`Enable`](./pr-6-enable.md) PR — that's what keeps every intermediate PR a safe no-op even after it
> merges.

## Factory wiring

"The factory" is [`connector/new.go`](../../connector/new.go) — it maps each provider to a constructor so
a connector can be built by provider name. Which case you're in:

- **Existing connector** (already does read/write — the common case). It **already has a factory entry**,
  so there's nothing to do here — subscribe work in later PRs just attaches new methods to the existing
  `*Connector`.
- **Brand-new provider** (no connector yet). Build the connector first, following
  [Adding a Proxy Connector](../../CONTRIBUTING.md#adding-a-proxy-connector) and
  [Adding a Deep Connector](../../CONTRIBUTING.md#adding-a-deep-connector), then add a constructor +
  dispatch-map entry in [`connector/new.go`](../../connector/new.go):

```go
var connectorConstructors = map[providers.Provider]outputConstructorFunc{
    // ...
    providers.Acme: wrapper(newAcmeConnector),
}

func newAcmeConnector(params common.ConnectorParams) (*acme.Connector, error) {
    return acme.NewConnector(acme.WithAuthenticatedClient(params.AuthenticatedClient))
}
```

There is **nothing subscribe-specific** to register: the caller obtains the connector via `New(...)`
and type-asserts it to `SubscribeConnector` / `WebhookVerifierConnector` / etc., so the assertion just
succeeds once your `*Connector` implements the interface methods (in later PRs).

## Twin providers

If your provider reuses another provider's connector implementation (same `*Connector`, same modules),
register the same constructor in `connector/new.go` for both provider keys so the twin shares the
implementation. Declare `Support.Subscribe` and `SubscribeRequirements` directly on the twin's
`providers/<twin>.go` (mirroring the original) so the twin carries its own subscribe metadata. The
canonical example is **[SalesforceJWT](../../providers/salesforceJWT.go)**, which shares the
**[Salesforce](../../providers/salesforce.go)** connector and modules — compare the two `ProviderInfo`
declarations.

## Files

- `providers/<provider>.go` — the `SetInfo(...)` declaration.
- `connector/new.go` — constructor + map entry (brand-new providers only).

## Checklist

- [ ] `Support.Subscribe` is `false` (the gate stays off).
- [ ] `SubscribeRequirements` reflects the provider's intended shape; `Registration` / `PostProcess` /
      `Maintenance` set only if applicable.
- [ ] Every `SubscribeRequirements` flag set to `new(true)` has a code comment linking the provider
      docs that justify it.
- [ ] For multi-module providers, subscribe declared on the right module(s); others stay
      `Subscribe: false`.
- [ ] Factory entry added **iff** the provider was not already registered.
- [ ] Twin providers (if any) declare their own metadata.
- [ ] No behavioral change (the provider stays gated off).

## Reviewer focus

- The provider is genuinely gated off (no path can activate it yet).
- `SubscribeRequirements` matches what later PRs will implement.
- Per-module scoping is correct for multi-module providers.

## Reference

- [The big picture](../../SUBSCRIBE_REFERENCES.md#the-big-picture) · [Core types](../../SUBSCRIBE_REFERENCES.md#core-types)
- Real declarations: [`providers/salesloft.go`](../../providers/salesloft.go),
  [`providers/salesforce.go`](../../providers/salesforce.go),
  [`connector/new.go`](../../connector/new.go)
