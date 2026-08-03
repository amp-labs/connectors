# Onboarding a New Provider to Subscribe

This guide walks through adding subscribe support for a new provider in `subscribe/`.

## Summary of steps

1. **Create** `subscribe/<provider>.go` declaring `var <provider>Config = ProviderConfig{...}` and populating only the subcomponents the provider needs.
2. **Register** it in the `providerConfigs` map in `config.go`, keyed by `providers.<Provider>`.

That's it. The Ampersand server already reads everything through
`subscribe.GetProviderConfig(module, providerInfo, deps).<Subcomponent>.<Method>()` — you do not
touch server code.

(For "twin" providers that reuse another provider's connector, also see [Twin providers](#twin-providers--providerinfoaliases).)

---

## ProviderConfig and Registry

A `ProviderConfig` is the per-provider declarative bundle of subscribe-related configuration — the `var <provider>Config = ProviderConfig{...}` literal you write in step 1. The `providerConfigs` map in `config.go` wraps each entry in a `ProviderConfigRegistry`:

```go
type ProviderConfigRegistry struct {
    DefaultModuleConfig *ProviderConfig
    Modules             map[common.ModuleID]*ProviderConfig
}
```

**Use `DefaultModuleConfig`** when subscribe behavior is module-agnostic (Outreach, Salesloft, Hubspot, Gong, HousecallPro).

**Use `Modules`** when your provider has multiple modules with differing subscribe behavior (Salesforce, Zoho, Google all do this — their non-subscribe modules have no entry).

You can use both (one `DefaultModuleConfig` plus a few module overrides) if needed.

`GetProviderConfig(module, providerInfo, deps)` resolves the module in this order: the `module` argument (typically the installation revision's module) → `providerInfo.DefaultModule` → empty string. If the resolved module has an entry in `Modules`, that's used; otherwise `DefaultModuleConfig`; if neither matches, it returns `ErrProviderConfigNotFound`. On success the returned pointer is a fresh copy with `module` / `providerInfo` / `deps` bound onto each subcomponent, so subcomponent methods answer module-aware questions without callers threading the arguments per call.

---

## Config Components

`ProviderConfig` bundles five subcomponents — one per subscribe-related concern. Each is value-embedded (not a pointer), and its zero value is a safe no-op, so a provider that needs only `Verification` populates only that field and leaves the others alone.

| Subcomponent          | Concern                                                                                                       | Typical providers              |
|-----------------------|---------------------------------------------------------------------------------------------------------------|--------------------------------|
| `Registration`        | One-time per-installation setup (e.g. Salesforce → AWS EventBridge wiring).                                  | Salesforce, Zoho               |
| `Subscription`        | Programmatic-API support, look-up-only status, the WatchFieldsAuto Salesloft quirk, and the subscribe-time request payload builder. | Anything that subscribes via API |
| `Maintenance`         | Periodic renewal of expiring webhook subscriptions.                                                          | Gmail (daily), Zoho (weekly)   |
| `PostProcess`         | Whether third-party setup runs after the connector's `Subscribe` returns. **Derived answer only** — the setup/teardown functions themselves are Ampersand-infrastructure concerns and live in Ampersand's internal systems. | Salesforce |
| `Verification`        | Webhook signature verification: params builder, verifier connector, bypass flag, event caster.               | Anyone whose webhooks need verification |

## The `deps` resolver seam

Some per-provider functions need data that only the Ampersand server can resolve (it lives behind the server's database or customer configuration). Those capabilities are declared as narrow interfaces in the `subscribe/deps` package and carried in `deps.Dependencies`:

```go
type Dependencies struct {
    Project         ProjectResolver          // project attributes (Salesforce CDC event-flag field naming)
    CDCOptimization CDCOptimizationResolver  // per-(project, group) Salesforce CDC quota opt-in
    Subscriptions   SubscriptionResultLister // stored subscription results (Attio signing secret recovery)
}
```

The server implements these interfaces and passes a `Dependencies` value to `GetProviderConfig`, which binds it onto the returned config; subcomponent methods thread it into your per-provider functions as their `deps deps.Dependencies` parameter. Most providers ignore it (`_ deps.Dependencies`).

If your provider needs an Ampersand-owned capability that doesn't exist yet, add a new narrow resolver interface in `subscribe/deps` (one file per resolver, mirroring `project.go` / `cdcoptimization.go` / `subscriptions.go`), add a field to `Dependencies`, and coordinate the implementation with the Ampersand team. Keep the interface as small as the provider function actually needs.

## Examples

### A worked example: adding `acme`

Suppose Acme subscribes via API, builds a webhook-endpoint payload at subscribe time, and verifies webhook signatures with a per-installation secret.

**Step 1: `subscribe/acme.go`**

```go
package subscribe

import (
    "context"

    "github.com/amp-labs/amp-common/openapi"
    "github.com/amp-labs/connectors/common"
    "github.com/amp-labs/connectors/providers/acme"
    "github.com/amp-labs/connectors/subscribe/deps"
)

// acmeConfig is the per-provider subscribe-config bundle for Acme. Acme subscribes via API,
// builds a subscribe-time webhook-endpoint payload, and verifies incoming webhooks with a
// per-installation secret.
var acmeConfig = ProviderConfig{
    Subscription: SubscriptionConfig{
        buildRequestFn: getAcmeRequest,
    },
    Verification: VerificationConfig{
        paramsFn:          getAcmeVerificationParams,
        verifierConnector: &acme.Connector{},
    },
}

func getAcmeVerificationParams(
    _ context.Context,
    _ deps.Dependencies,
    req *deps.VerificationRequest,
) (*common.VerificationParams, error) {
    if req == nil || req.Installation == nil {
        return nil, errInstallationNotFound
    }

    return &common.VerificationParams{
        Param: &acme.AcmeVerificationParams{Secret: req.Installation.Id},
    }, nil
}

func getAcmeRequest(
    _ context.Context,
    _ deps.Dependencies,
    inst *openapi.Installation,
    _ *openapi.Revision,
    _ *common.RegistrationResult,
    _ *openapi.Connection,
    webhookURL string,
) (any, error) {
    return &acme.SubscriptionRequest{
        UniqueRef:       "amp_" + inst.Id,
        WebhookEndPoint: webhookURL,
        Secret:          inst.Id,
    }, nil
}
```

Note that `webhookURL` — the event-receipt endpoint — arrives pre-constructed from the caller; endpoint construction depends on Ampersand deployment configuration, so it never happens inside this package.

**Step 2: register in `providerConfigs`**

Edit [`subscribe/config.go`](https://github.com/amp-labs/connectors/blob/main/subscribe/config.go) and add an entry to the `providerConfigs` map:

```go
var providerConfigs = map[providers.Provider]ProviderConfigRegistry{
    // ...existing entries...
    providers.Acme: {DefaultModuleConfig: &acmeConfig},
}
```

Use `DefaultModuleConfig` when the provider has no module-specific differences. Use `Modules` when it does.

---

## Component reference

### `RegistrationConfig` (`registration.go`)

For providers that need a one-time installation-level setup before per-object subscriptions can be created (Salesforce → AWS EventBridge wiring is the canonical example).

Fields you populate:

| Field            | Purpose |
|------------------|---------|
| `emptyResultFn`  | Returns a fresh `*common.RegistrationResult` with provider-specific `.Result`. Must be a factory (not a singleton) — `.Result` is mutated downstream. |
| `buildParamsFn`  | Builds the payload inside `common.SubscriptionRegistrationParams.Request` for a given installation. Signature: `func(ctx context.Context, inst *openapi.Installation) (any, error)`. |

Methods:

| Method                                  | Returns | Notes |
|-----------------------------------------|---------|-------|
| `IsRequired(connector)`                 | bool    | Reads `ProviderInfo.SubscribeRequirements.Registration`, gated on the connector implementing `connectors.RegisterSubscribeConnector`. |
| `EmptyResult()`                         | `*common.RegistrationResult` | `nil` when no `emptyResultFn` declared. |
| `BuildParams(ctx, inst)`                | `(any, error)` | Returns `ErrRegistrationParamsBuilderNotDeclared` for providers that don't require registration; `(nil, nil)` for providers that do require it but declared no builder (connector accepts an empty `Request`). |

### `SubscriptionConfig` (`subscription.go`)

For providers that subscribe to webhook events, whether programmatically or by manual provider-app configuration.

Fields you populate:

| Field                       | Purpose |
|-----------------------------|---------|
| `requiresWatchFieldsAuto`   | Salesloft quirk — requires `WatchFieldsAuto="all"` for subscribe update events. |
| `buildRequestFn`            | Builds the per-installation subscribe-time payload (Pub/Sub topic, webhook URL, secret, …). Nil when no custom payload is needed. |

The `SubscriptionRequestBuilder` signature:

```go
type SubscriptionRequestBuilder func(
    ctx context.Context,
    deps deps.Dependencies,
    inst *openapi.Installation,
    rev *openapi.Revision,
    registrationResult *common.RegistrationResult,
    conn *openapi.Connection,
    webhookURL string,
) (any, error)
```

`conn` is the same value as `&inst.Connection`; `webhookURL` is the caller-constructed event-receipt endpoint (ignore it if your provider doesn't need one, e.g. Google's Pub/Sub topic builder).

Methods:

| Method                            | Returns | Notes |
|-----------------------------------|---------|-------|
| `IsSupportedViaAPI()`             | bool    | Reads `ProviderInfo.SubscribeRequirements.SubscribeByAPI`. |
| `SubscribeManually()`             | bool    | True when the integration is configured at the provider-app level rather than registered via API (`ProviderInfo.Support.Subscribe && !SubscribeByAPI`). |
| `RequiresWatchFieldsAutoAll()`    | bool    | Reads `requiresWatchFieldsAuto`. |
| `BuildRequest(ctx, inst, rev, regResult, webhookURL)` | `(any, error)` | Invokes `buildRequestFn` with the bound `deps` and `&inst.Connection` as the `conn` arg. Returns `(nil, nil)` when no builder declared. |

### `MaintenanceConfig` (`maintenance.go`)

For providers whose webhook subscriptions expire and need renewal (Gmail watch subscriptions expire after 7 days and are renewed daily; Zoho renews weekly).

Fields you populate:

| Field             | Purpose |
|-------------------|---------|
| `renewalInterval` | How often maintenance runs. |

Also add an entry to the `maintenancePeriods` map in `maintenance.go` — it backs the package function `GetMaintenancePeriod(provider)`, used by callers that have only a provider in scope (no resolved config).

Methods:

| Method            | Returns | Notes |
|-------------------|---------|-------|
| `ShouldPerform()` | bool    | Reads `ProviderInfo.SubscribeRequirements.Maintenance`. |
| `Interval()`      | `(time.Duration, bool)` | Comma-ok: returns `(0, false)` when `renewalInterval` is zero/unset. Idiomatic use: `if p, ok := cfg.Maintenance.Interval(); ok { ... }`. |

### `PostProcessConfig` (`postprocess.go`)

For providers needing third-party setup after the connector's Subscribe returns (e.g. Salesforce — AWS EventBridge wiring).

Only the derived answer lives in this package:

| Method            | Returns | Notes |
|-------------------|---------|-------|
| `ShouldPerform()` | bool    | Reads `ProviderInfo.SubscribeRequirements.PostProcess`. |

The setup/teardown functions themselves are Ampersand-infrastructure concerns — they reach AWS EventBridge and other Ampersand-managed services — and live in Ampersand's internal systems. If your provider needs post-processing, set `SubscribeRequirements.PostProcess` in its `ProviderInfo` and coordinate the internal implementation with the Ampersand team.

### `VerificationConfig` (`verification.go`)

Fields you populate:

| Field               | Purpose |
|---------------------|---------|
| `paramsFn`          | Builds the provider-specific `*common.VerificationParams`. |
| `verifierConnector` | The zero-value connector pointer used for webhook signature verification (e.g. `&hubspot.Connector{}`). Returned unwrapped — callers wanting instrumentation (e.g. metrics) decorate it themselves. |
| `bypassed`          | When true, webhook signature verification is skipped (e.g. when events are synthetic republishes carrying no provider signature). |
| `eventCaster`       | Casts raw webhook event maps into typed `common.SubscriptionEvent` slices. Usually `CastSubscriptionEvents[<provider>.SubscriptionEvent]`. |

Methods:

| Method                | Returns | Notes |
|-----------------------|---------|-------|
| `Params(ctx, req)`    | `(*common.VerificationParams, error)` | Returns `errVerificationParamsFuncNotFound` when no `paramsFn`. |
| `Connector(ctx)`      | `(connectors.WebhookVerifierConnector, error)` | Returns `errWebhookVerificationNotSupported` when no `verifierConnector`. |
| `ShouldBypass()`      | bool    | Reads `bypassed`. |
| `CastEvents(list)`    | `([]common.SubscriptionEvent, error)` | Returns `errSubscriptionEventCasterNotDeclared` when no `eventCaster`. |

The `VerificationParamsFunc` signature:

```go
type VerificationParamsFunc func(
    ctx context.Context,
    deps deps.Dependencies,
    req *deps.VerificationRequest,
) (*common.VerificationParams, error)
```

`deps.VerificationRequest` bundles the delivery being verified with its wire-type context:

```go
type VerificationRequest struct {
    Payload                 *webhook.PubsubPayload // the inbound webhook delivery being verified
    Integration             *openapi.Integration
    Installation            *openapi.Installation
    ProviderApp             *openapi.ProviderApp
    ProviderAppClientSecret string // carried separately — the openapi.ProviderApp wire type deliberately excludes secrets
}
```

`ProviderAppClientSecret` is required by providers that sign webhooks with the OAuth client secret (Hubspot, Jobber).

There is also a package function `IsHookdeckGatewayProvider(provider) bool` — true for everyone except Hubspot, Salesforce, SalesforceJWT. It's not a `VerificationConfig` method because its callers have only a provider string in scope (no `providerInfo` / module); the server uses it to choose between Hookdeck and CloudFunction event endpoints when constructing the `webhookURL` your builder receives.

---

## Twin providers — `providerInfoAliases`

If your provider in the connectors library reuses another provider's connector implementation and does not have its own `SubscribeRequirements` declared, add an entry to `providerInfoAliases` in `aliases.go`. Calls to `ResolveProviderInfoAlias` swap in the twin's `ProviderInfo`, with the original `.Name` preserved so log attribution stays accurate.

`SalesforceJWT → Salesforce` is the canonical example: SalesforceJWT shares the Salesforce connector implementation and the same modules, but its own `ProviderInfo` entry doesn't declare subscribe metadata.

```go
// subscribe/aliases.go
var providerInfoAliases = map[providers.Provider]providers.Provider{
    // SalesforceJWT shares the Salesforce connector implementation and the same modules,
    // but salesforceJWT.go does not yet declare SubscribeRequirements.
    providers.SalesforceJWT: providers.Salesforce,
}
```

Pair this with registering the same `*ProviderConfig` pointer under both provider keys in `providerConfigs`:

```go
// subscribe/config.go
var providerConfigs = map[providers.Provider]ProviderConfigRegistry{
    providers.Salesforce: {
        Modules: map[common.ModuleID]*ProviderConfig{
            providers.ModuleSalesforceCRM: &salesforceConfig,
        },
    },
    providers.SalesforceJWT: {
        Modules: map[common.ModuleID]*ProviderConfig{
            providers.ModuleSalesforceCRM: &salesforceConfig,
        },
    },
    // ...
}
```

`ResolveProviderInfoAlias` must be called by the caller exactly once, immediately after fetching `ProviderInfo`. The Ampersand server's existing entry points already do this, so unless a new entry point that fetches `ProviderInfo` is being added, you don't need to think about it.

---

## Real providers, different shapes

### Salesforce — full surface

Touches Registration, Subscription, and Verification. Registered for both `Salesforce` and `SalesforceJWT` (same `&salesforceConfig` pointer under both keys, scoped to `ModuleSalesforceCRM`). Its post-processing (AWS EventBridge wiring) is Ampersand infrastructure, so no `PostProcess` declaration appears here — `ShouldPerform` still answers true via `ProviderInfo.SubscribeRequirements.PostProcess`.

```go
var salesforceConfig = ProviderConfig{
    Registration: RegistrationConfig{
        emptyResultFn: func() *common.RegistrationResult {
            return &common.RegistrationResult{Result: &salesforce.ResultData{}}
        },
        buildParamsFn: buildSalesforceRegistrationParams,
    },
    Subscription: SubscriptionConfig{
        buildRequestFn: getSalesforceRequest,
    },
    Verification: VerificationConfig{
        paramsFn:          getSalesforceVerificationParams,
        verifierConnector: &salesforce.Connector{},
    },
}
```

Salesforce's request builder is also the only one that uses the `deps` resolvers today (project app name + CDC optimization config). See `subscribe/salesforce.go` for the per-provider helpers.

### Outreach / Salesloft — subscribe-via-API with webhook verification

Subscribes via API + builds a webhook-endpoint payload + verifies with a per-installation secret. Salesloft adds the WatchFieldsAuto quirk.

```go
var salesloftConfig = ProviderConfig{
    Subscription: SubscriptionConfig{
        buildRequestFn:          getSalesloftRequest,
        requiresWatchFieldsAuto: true,
    },
    Verification: VerificationConfig{
        paramsFn:          getSalesloftVerificationParams,
        verifierConnector: &salesloft.Connector{},
    },
}
```

### Google (Gmail / Calendar) — module-scoped with maintenance and bypass

Only the `gmail` and `calendar` modules support subscribe, so they're registered under `Modules` (as `googleConfig` and `googleCalendarConfig` respectively), not `DefaultModuleConfig`. Webhooks arrive as synthetic Pub/Sub republishes so verification is bypassed.

```go
var googleConfig = ProviderConfig{
    Subscription: SubscriptionConfig{
        buildRequestFn: getGoogleRequest,
    },
    Maintenance: MaintenanceConfig{
        renewalInterval: time.Hour * 24,
    },
    Verification: VerificationConfig{
        bypassed:    true,
        eventCaster: CastSubscriptionEvents[google.SubscriptionEvent],
    },
}

// In providerConfigs:
providers.Google: {
    Modules: map[common.ModuleID]*ProviderConfig{
        providers.ModuleGoogleGmail:    &googleConfig,
        providers.ModuleGoogleCalendar: &googleCalendarConfig,
    },
},
```

### Hubspot — look-up-only with verification

Subscriptions are configured at the provider-app level so there's no `Subscription` / `Registration` declaration; only verification data. Its params builder reads `req.ProviderAppClientSecret`.

```go
var hubspotConfig = ProviderConfig{
    Verification: VerificationConfig{
        paramsFn:          getHubspotVerificationParams,
        verifierConnector: &hubspot.Connector{},
        bypassed:          true,
        eventCaster:       CastSubscriptionEvents[hubspot.SubscriptionEvent],
    },
}
```

### Gong / HousecallPro — minimal look-up-only

The smallest possible config: a verifier connector and the bypass flag. Two-line declaration.

```go
var gongConfig = ProviderConfig{
    Verification: VerificationConfig{
        verifierConnector: &gong.Connector{},
        bypassed:          true,
    },
}
```

