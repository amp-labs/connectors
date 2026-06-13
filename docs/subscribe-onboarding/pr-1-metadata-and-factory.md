# PR 1 — Metadata scaffold + factory wiring *(base)*

> Part of the [Subscribe Onboarding PR Process](../../SUBSCRIBE_PR_PROCESS.md). Implementation
> reference: [`SUBSCRIBE_ONBOARDING.md`](../../SUBSCRIBE_ONBOARDING.md).

**Required.** This is the base of the stack — everything else stacks on it.

## Goal

Establish the provider in the catalog with subscribe **gated off**, and ensure the connector is
constructible. This PR changes **no runtime behavior** — it's a safe no-op until the final `Enable` PR.

## What you implement

1. **Provider metadata** — declare `SubscribeRequirements` on the provider's `ProviderInfo`, with the
   activation flags off.
2. **Factory wiring** — *only if the provider is brand-new.* Most providers already have a constructor
   (for read/write), in which case there's nothing to do here.

## Files

- `providers/<provider>.go` — the `SetInfo(...)` declaration.
- `connector/new.go` — constructor + dispatch-map entry (brand-new providers only).

## Steps

1. In `providers/<provider>.go`, set `Support.Subscribe: false` and declare `SubscribeRequirements` with
   `SubscribeByAPI: new(false)`. Set `Registration` / `PostProcess` / `Maintenance` to `new(true)`
   **only if** the provider will need them (these requirement flags are harmless while gated off).
2. If the provider has no entry in `connector/new.go`, add a `new<Provider>Connector` constructor and a
   `providers.<Provider>: wrapper(new<Provider>Connector)` map entry.

## Example

```go
// providers/<provider>.go  — gated OFF
Support: Support{
    Read:      true,
    Write:     true,
    Subscribe: false, // ← flip to true in PR 6
},
SubscribeRequirements: &SubscribeRequirements{
    SubscribeByAPI: new(false), // ← flip to new(true) in PR 6
    // Registration / PostProcess / Maintenance: new(true) only if applicable
},
```

```go
// connector/new.go — brand-new providers only
providers.Acme: wrapper(newAcmeConnector),

func newAcmeConnector(params common.ConnectorParams) (*acme.Connector, error) {
    return acme.NewConnector(acme.WithAuthenticatedClient(params.AuthenticatedClient))
}
```

## Checklist

- [ ] `Support.Subscribe` is `false` and `SubscribeByAPI` is `new(false)` (or unset).
- [ ] `SubscribeRequirements` reflects the provider's intended shape; requirement flags set only if
      applicable.
- [ ] Factory entry added **iff** the provider was not already registered.
- [ ] `go build ./...` passes; no behavioral change.

## Reviewer focus

- Confirm the provider is genuinely gated off (no path can activate it yet).
- Confirm `SubscribeRequirements` matches what later PRs will implement.

## Reference

- [Provider metadata](../../SUBSCRIBE_ONBOARDING.md#provider-metadata)
- [Factory wiring](../../SUBSCRIBE_ONBOARDING.md#factory-wiring)
