# Onboarding a New Provider to Subscribe

This guide walks through adding **Subscribe** support (webhook subscriptions) for a provider in the
`github.com/amp-labs/connectors` library. It covers the interfaces you implement, the provider
metadata you declare, the event types you parse, how the connector is wired into the factory, how to
test it, and the **recommended way to break the work into PRs**.

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
form a ladder — each extends the one below it — so you only implement the rungs your provider needs.

```
SubscriptionMaintainerConnector   (renews expiring subscriptions — provider-specific, if needed)
        ▲ extends
RegisterSubscribeConnector        (one-time per-installation setup, e.g. Salesforce → EventBridge — provider-specific, if needed)
        ▲ extends
SubscribeConnector                (create / update / delete subscriptions)
        ▲ extends
WebhookVerifierConnector          (verify incoming webhook signatures)
        ▲ extends
Connector + BatchRecordReaderConnector   (base client + fetch records by id for webhook enrichment)
```

All four interfaces are defined in [`connectors.go`](./connectors.go). Note that
`WebhookVerifierConnector` embeds `Connector` (base HTTP client + provider identity) **and**
`BatchRecordReaderConnector` (`GetRecordsByIds`). This does **not** require a full read connector
(no `Read`/pagination) — only `GetRecordsByIds`, which the caller uses to fetch a record's full state
by id after a webhook event arrives.

| Interface | Methods you add | When you need it |
|-----------|-----------------|------------------|
| `WebhookVerifierConnector` | `VerifyWebhookMessage` | Provider signs its webhooks (almost always). |
| `SubscribeConnector` | `Subscribe`, `UpdateSubscription`, `DeleteSubscription`, `EmptySubscriptionParams`, `EmptySubscriptionResult` | Provider lets you create subscriptions programmatically via API. |
| `RegisterSubscribeConnector` | `Register`, `DeleteRegistration`, `EmptyRegistrationParams`, `EmptyRegistrationResult` | **Provider-specific, if needed** — only when the provider needs a one-time, installation-level setup shared by all object subscriptions. |
| `SubscriptionMaintainerConnector` | `RunScheduledMaintenance` | **Provider-specific, if needed** — only when subscriptions/watches expire after a TTL and must be renewed on a schedule. |

A "UI Subscription only" provider (subscriptions configured in the provider's own UI, e.g. Hubspot/Gong)
only implements `WebhookVerifierConnector` — the caller reads its events but never calls `Subscribe`.

---

## Interface reference

All signatures below are from [`connectors.go`](./connectors.go). Implement them as methods on your
provider's `*Connector` type, and add a compile-time assertion (see
[`providers/salesloft/subscribe.go`](./providers/salesloft/subscribe.go)):

```go
var _ connectors.SubscribeConnector = &Connector{}
```

### `WebhookVerifierConnector`

```go
VerifyWebhookMessage(
    ctx context.Context,
    request *common.WebhookRequest,
    params *common.VerificationParams,
) (bool, error)
```

Return `true` to allow webhook processing, `false` to reject as untrusted, and an `error` only for
*unexpected* failures.

### `SubscribeConnector`

```go
Subscribe(ctx context.Context, params common.SubscribeParams) (*common.SubscriptionResult, error)

UpdateSubscription(
    ctx context.Context,
    params common.SubscribeParams,
    previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error)

DeleteSubscription(ctx context.Context, previousResult common.SubscriptionResult) error

EmptySubscriptionParams() *common.SubscribeParams
EmptySubscriptionResult() *common.SubscriptionResult
```

- **`Subscribe`** translates the normalized `params.SubscriptionEvents` (objects → event types) into
  provider-specific API calls and returns the resulting state. On partial failure it should roll back
  what it created (see Salesloft/Outreach for the parallel-create-with-rollback pattern).
- **`UpdateSubscription`** reconciles the existing subscription (`previousResult`) with the new desired
  state (`params`). The framework only calls this after it detects a change.
- **`DeleteSubscription`** tears down everything identified by `previousResult`.
- **`Empty*`** return zero-value instances with the provider-specific `.Request` / `.Result` populated
  so the framework can deserialize stored DB state back into your concrete types.

### `RegisterSubscribeConnector`

> **Provider-specific — implement only if needed.** Most providers do not need registration. Add this
> interface only when the provider requires a one-time, installation-level setup step (Salesforce is
> the canonical case).

```go
Register(ctx context.Context, params common.SubscriptionRegistrationParams) (*common.RegistrationResult, error)
DeleteRegistration(ctx context.Context, previousResult common.RegistrationResult) error
EmptyRegistrationParams() *common.SubscriptionRegistrationParams
EmptyRegistrationResult() *common.RegistrationResult
```

`Register` is a one-time per-installation operation that creates shared infrastructure
(`Subscribe` later hangs each object subscription off it). It must roll back its own partial work on
failure and set `Status` accordingly — see the Salesforce example below.

### `SubscriptionMaintainerConnector`

> **Provider-specific — implement only if needed.** Add this interface only when the provider's
> subscriptions/watches expire after a TTL and must be periodically renewed.

```go
RunScheduledMaintenance(
    ctx context.Context,
    params common.SubscribeParams,
    previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error)
```

Called on a schedule (the cadence is configured by the caller). It renews/refreshes
the subscription in `previousResult` and returns the updated state.

---

## Core types

From [`common/types.go`](./common/types.go). These are the provider-agnostic payloads the framework
passes you; the `any` fields are where your provider-specific structs live.

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
SubscriptionEventTypeCreate            = "create"
SubscriptionEventTypeUpdate            = "update"
SubscriptionEventTypeDelete            = "delete"
SubscriptionEventTypeAssociationUpdate = "associationUpdate"
SubscriptionEventTypeOther             = "other"
```

---

## Event types

When a webhook arrives, the caller casts the raw payload into your typed events and asks them
provider-agnostic questions (what object? what record id? create or update?). You implement these
interfaces (from `common/types.go`) on a provider event type — see
[`providers/salesloft/subscribeEvent.go`](./providers/salesloft/subscribeEvent.go) for a complete,
readable example.

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

---

## Verification

Implement `VerifyWebhookMessage` on your `*Connector`. The pattern (from
[`providers/salesloft/subscribeEvent.go`](./providers/salesloft/subscribeEvent.go)):

1. Pull your provider-specific params out of `params.Param` with `common.AssertType`.
2. Read the signature header from `req.Headers`.
3. Recompute the expected signature over `req.Body` (and sometimes method/url/timestamp) with the
   shared secret, and compare with `hmac.Equal`.

```go
type SalesloftVerificationParams struct {
    Secret string `json:"secret,omitempty"`
}

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

---

## Provider metadata

Declare subscribe capability in the provider's hand-written info file,
`providers/<provider>.go`, inside `SetInfo(...)`. Two pieces matter:

- **`Support.Subscribe`** — the master "this provider supports subscribe at all" switch.
- **`SubscribeRequirements`** — what the provider *needs* in order to subscribe. All fields are `*bool`;
  use the package pointer helper `new(true)`. From `providers/types.gen.go`:

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

Salesloft (API subscribe, no registration/maintenance/post-process):

```go
// providers/salesloft.go
Support: Support{
    Read:      true,
    Write:     true,
    Proxy:     true,
    Subscribe: true,
},
SubscribeRequirements: &SubscribeRequirements{
    SubscribeByAPI: new(true),
},
```

Salesforce declares it **per module** (`ModuleSalesforceCRM` only) — note other modules keep
`Subscribe: false`:

```go
// providers/salesforce.go (within ModuleSalesforceCRM)
Support: Support{ Subscribe: true, /* ... */ },
SubscribeRequirements: &SubscribeRequirements{
    Registration:   new(true),
    PostProcess:    new(true),
    SubscribeByAPI: new(true),
},
```

> **Gating rule (important for PR staging).** The caller activates a provider based on these flags:
> the API path is gated on `SubscribeByAPI`, the manual / UI-subscription path on `Support.Subscribe`. **Keep
> both off until the implementation is complete and verified**, then flip them on in a final
> "enable" PR. Declaring the *requirement* flags (`Registration`/`PostProcess`/`Maintenance`) earlier
> is harmless — they're only consulted once subscribe is active.

---

## Factory wiring

The caller obtains a connector through the factory in [`connector/new.go`](./connector/new.go), which
maps a provider to its constructor:

```go
func New(provider providers.Provider, params common.ConnectorParams) (connectors.Connector, error) {
    constructor, ok := connectorConstructors[provider]
    if !ok {
        return nil, ErrInvalidProvider
    }
    return constructor(params)
}

var connectorConstructors = map[providers.Provider]outputConstructorFunc{
    // ...
    providers.Salesloft: wrapper(newSalesloftConnector),
    // ...
}

func newSalesloftConnector(params common.ConnectorParams) (*salesloft.Connector, error) {
    return salesloft.NewConnector(salesloft.WithAuthenticatedClient(params.AuthenticatedClient))
}
```

The caller then type-asserts the returned `connectors.Connector` to `SubscribeConnector` /
`WebhookVerifierConnector` / etc. **So there is nothing subscribe-specific to register in the
factory** — once your `*Connector` implements the interface methods, the assertion succeeds.

**Most providers being onboarded for subscribe already have a factory entry** (they already do
read/write), so subscribe work just attaches new methods to the existing `*Connector`. A brand-new
provider also needs the constructor + map entry above.

---

## A worked example: adding `acme`

Acme subscribes via API, builds a webhook-endpoint payload at subscribe time, and signs webhooks with
a per-installation secret. Files live under `providers/acme/`.

**Verification + event type** (`providers/acme/subscribeEvent.go`):

```go
package acme

type SubscriptionEvent map[string]any

type AcmeVerificationParams struct {
    Secret string `json:"secret,omitempty"`
}

var _ common.SubscriptionEvent = SubscriptionEvent{}

func (c *Connector) VerifyWebhookMessage(
    ctx context.Context, req *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
    vp, err := common.AssertType[*AcmeVerificationParams](params.Param)
    if err != nil {
        return false, err
    }
    sig := req.Headers.Get("X-Acme-Signature")
    expected := hex.EncodeToString(hmacSHA256(vp.Secret, req.Body))
    return hmac.Equal([]byte(sig), []byte(expected)), nil
}

func (e SubscriptionEvent) ObjectName() (string, error) { /* ... */ }
func (e SubscriptionEvent) RecordId() (string, error)   { /* ... */ }
func (e SubscriptionEvent) EventType() (common.SubscriptionEventType, error) { /* ... */ }
// ...RawEventName, Workspace, EventTimeStampNano, RawMap, PreLoadData
```

**Subscribe** (`providers/acme/subscribe.go`):

```go
package acme

var _ connectors.SubscribeConnector = &Connector{}

type SubscriptionRequest struct {
    WebhookEndpoint string `json:"webhookEndpoint"`
    Secret          string `json:"secret"`
}

type SubscriptionResult struct {
    SubscriptionID string `json:"subscriptionId"`
}

func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
    return &common.SubscribeParams{Request: &SubscriptionRequest{}}
}

func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
    return &common.SubscriptionResult{Result: &SubscriptionResult{}}
}

func (c *Connector) Subscribe(
    ctx context.Context, params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
    req, err := common.AssertType[*SubscriptionRequest](params.Request)
    if err != nil {
        return nil, err
    }
    // For each object+event in params.SubscriptionEvents, call Acme's API.
    // Roll back created subscriptions on partial failure. Return the actual state.
    // ...
}

func (c *Connector) UpdateSubscription(
    ctx context.Context, params common.SubscribeParams, previous *common.SubscriptionResult,
) (*common.SubscriptionResult, error) { /* reconcile previous → desired */ }

func (c *Connector) DeleteSubscription(
    ctx context.Context, previous common.SubscriptionResult,
) error { /* delete everything in previous.Result */ }
```

**Metadata** (`providers/acme.go`) — declared but gated off until the final PR:

```go
Support: Support{ Read: true, Write: true, Subscribe: false /* flip in final PR */ },
SubscribeRequirements: &SubscribeRequirements{
    SubscribeByAPI: new(false), // flip to new(true) in final PR
},
```

The per-installation request payload (`SubscriptionRequest{WebhookEndpoint, Secret}`) is *built by the
caller* and handed to you in `params.Request` — the caller constructs the webhook endpoint URL and
secret. You only define the struct and consume it.

---

## Real providers, different shapes

| Provider | Interfaces implemented | Notes |
|----------|------------------------|-------|
| **Hubspot / Gong** | `WebhookVerifierConnector` only | UI Subscription only: subscriptions configured in the provider UI. The caller reads events; never calls `Subscribe`. |
| **Outreach / Salesloft** | `SubscribeConnector` | API subscribe + verify. Parallel create with rollback. Salesloft also needs the `WatchFieldsAuto="all"` quirk (set by the caller). |
| **Salesforce** | `RegisterSubscribeConnector` | Registers EventChannel/NamedCredential/EventRelayConfig, then subscribes PlatformEventChannelMembers. Needs `PostProcess` (AWS EventBridge, done outside the connector). |

---

## Registration (Salesforce example)

For providers needing one-time installation-level setup. From
[`providers/salesforce/register.go`](./providers/salesforce/register.go):

```go
type RegistrationParams struct {
    UniqueRef             string `json:"uniqueRef"             validate:"required"`
    Label                 string `json:"label"                 validate:"required"`
    AwsNamedCredentialArn string `json:"awsNamedCredentialArn" validate:"required"`
}

type ResultData struct {
    EventChannel     *EventChannel     `validate:"required"`
    NamedCredential  *NamedCredential  `validate:"required"`
    EventRelayConfig *EventRelayConfig `validate:"required"`
}

func (c *Connector) EmptyRegistrationParams() *common.SubscriptionRegistrationParams {
    return &common.SubscriptionRegistrationParams{Request: &RegistrationParams{}}
}
func (c *Connector) EmptyRegistrationResult() *common.RegistrationResult {
    return &common.RegistrationResult{Result: &ResultData{}}
}

func (c *Connector) Register(
    ctx context.Context, params common.SubscriptionRegistrationParams,
) (*common.RegistrationResult, error) {
    sfParams, ok := params.Request.(*RegistrationParams)
    // ...create resources...
    result, err := c.register(ctx, sfParams)
    if err != nil {
        if rollbackErr := c.rollbackRegister(ctx, result); rollbackErr != nil {
            return &common.RegistrationResult{Status: common.RegistrationStatusFailedToRollback},
                errors.Join(rollbackErr, err)
        }
        return &common.RegistrationResult{Status: common.RegistrationStatusFailed}, err
    }
    return &common.RegistrationResult{
        RegistrationRef: result.EventRelayConfig.Id,
        Result:          result,
        Status:          common.RegistrationStatusSuccess,
    }, nil
}
```

Key points: `Register` **rolls back its own partial work** on failure and reports `Status`
(`Success` / `Failed` / `FailedToRollback`); `DeleteRegistration` tears resources down in reverse
order. Set `SubscribeRequirements.Registration: new(true)`.

### PostProcess

`PostProcess` work (e.g. wiring AWS EventBridge after Salesforce subscribes) is performed outside the
connector by the caller, not in the connector. The connector's only obligation is to **return the
data the post-processor needs** in `RegistrationResult.Result` (for Salesforce, the `EventChannel` id,
etc.). Set `SubscribeRequirements.PostProcess: new(true)`; there is no connector method to implement,
so it does not get its own connector PR — fold the flag into the registration/enable PR.

---

## Maintenance

For providers whose subscriptions expire. Implement `RunScheduledMaintenance` (renew the watch/
subscription in `previousResult`, return the refreshed state) and set
`SubscribeRequirements.Maintenance: new(true)`. The renewal *cadence* is configured by the caller.

---

## Twin providers

If your provider reuses another provider's connector implementation (same `*Connector`, same modules),
register the same constructor in `connector/new.go` for both provider keys so the twin shares the
implementation. Declare `Support.Subscribe` and `SubscribeRequirements` directly on the twin's
`providers/<twin>.go` (mirroring the original) so the twin carries its own subscribe metadata. The
canonical example is **SalesforceJWT**, which shares the Salesforce connector and modules.

---

## Testing

1. **Compile-time assertions** — add `var _ connectors.SubscribeConnector = &Connector{}` (and the
   verifier/registration variants) so a missing method fails the build.
2. **Unit tests** — table-driven tests for `VerifyWebhookMessage` (valid/invalid/missing signature) and
   for each event method (`EventType`, `ObjectName`, `RecordId`, …) using captured real payloads. Use
   the existing testroutines helpers (e.g. `test/utils/testroutines/`).
3. **Manual end-to-end harness** — add `test/<provider>/subscribe/subscribe.go`, a small `main` that
   loads creds, builds the connector, and calls `Subscribe` against a real sandbox. Model it on
   [`test/outreach/subscribe/subscribe.go`](./test/outreach/subscribe/subscribe.go):

   ```go
   conn := connTest.GetOutreachConnector(ctx)
   result, err := conn.Subscribe(ctx, common.SubscribeParams{
       SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
           "account": {Events: []common.SubscriptionEventType{
               common.SubscriptionEventTypeCreate,
               common.SubscriptionEventTypeUpdate,
               common.SubscriptionEventTypeDelete,
           }},
       },
       Request: &outreach.SubscriptionRequest{ /* ... */ },
   })
   ```

   Trigger a change in the provider sandbox and confirm the webhook is received and verifies.

See [`CONTRIBUTING.md`](./CONTRIBUTING.md) for credential setup (`creds.json`) and the dev environment.

---

## Recommended PR breakdown

Land subscribe support as a **stack** of small, reviewable PRs. Each PR is stacked on the one below it,
and the provider stays **gated off** until everything is implemented and verified — so every PR in the
stack is a safe no-op until the final `Enable` PR at the top.

The stack mirrors the interface ladder from [The big picture](#the-big-picture): the base establishes
the provider, then each rung adds one interface, bottom-up.

```
  Enable the provider          flip Support.Subscribe + SubscribeByAPI on   ← merge last (top)
        ▲ stacked on
  Maintenance                  SubscriptionMaintainerConnector              (optional)
        ▲ stacked on
  Registration                 RegisterSubscribeConnector                   (optional)
        ▲ stacked on
  Subscribe / Update / Delete  SubscribeConnector
        ▲ stacked on
  Verification                 WebhookVerifierConnector
        ▲ stacked on
  Metadata scaffold (gated off) + factory wiring               ← base, merge first (bottom)
```

Create the branches bottom-to-top (the base merges first):

1. **Metadata scaffold + factory wiring** *(base)* — declare the `SubscribeRequirements` shape in
   `providers/<provider>.go`, but keep `Support.Subscribe: false` and `SubscribeByAPI: new(false)`
   (gated off). Add the `connector/new.go` constructor + map entry **only if the provider is
   brand-new**. Zero behavioral risk.
2. **Verification** — `VerifyWebhookMessage` + provider `VerificationParams` struct + the event type(s)
   implementing `SubscriptionEvent` (+ `SubscriptionUpdateEvent` / `CollapsedSubscriptionEvent` as
   needed) + compile-time assertion + unit tests.
3. **Subscribe / Update / Delete** — `SubscribeConnector` methods + `Request`/`Result` structs +
   `Empty*` + the `test/<provider>/subscribe/` harness.
4. **Registration** *(optional — only if the provider needs one-time setup)* —
   `RegisterSubscribeConnector` methods + rollback; set `Registration: new(true)` (+
   `PostProcess: new(true)` and ensure `RegistrationResult.Result` carries what the post-processor
   needs).
5. **Maintenance** *(optional — only if subscriptions expire)* — `RunScheduledMaintenance` + set
   `Maintenance: new(true)`.
6. **Enable the provider** *(top)* — flip `Support.Subscribe: true` and `SubscribeByAPI: new(true)`.
   A one-line change that's trivial to review and to revert. Once this and the caller-side
   configuration are in place, the provider is live.

Why this order: metadata first establishes the provider's intended shape with zero behavioral risk;
verification and the subscribe methods are independent files that compile on their own; optional rungs
only appear for providers that need them; the final enable PR is a one-line flip. This matches the
"small, stacked, safest-first" PR convention.

> The caller also needs a small configuration change to consume the new provider — supplying the
> per-installation verification params and subscribe request payload your connector expects. That
> change lives outside this repo and lands after the connectors stack.
