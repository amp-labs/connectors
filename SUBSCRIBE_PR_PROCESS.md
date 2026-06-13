# Subscribe Onboarding — PR Process

How to **ship** subscribe support for a new provider as a clean stack of pull requests.

This doc is about *process*: how to slice the work, what each PR contains, the order to merge them, and
what reviewers should check. For the *implementation* details (interfaces, types, event parsing,
verification, metadata, factory wiring, worked examples), see the companion reference:
[**`SUBSCRIBE_ONBOARDING.md`**](./SUBSCRIBE_ONBOARDING.md).

---

## Principles

1. **Small, stacked PRs.** One concern per PR, each stacked on the one below it. A reviewer should be
   able to understand a PR without holding the whole feature in their head.
2. **Safest-first.** Start with the change that carries zero behavioral risk (metadata, gated off) and
   end with the one-line switch that turns the provider on.
3. **Gated off until the end.** The provider's subscribe flags stay `false` for the entire stack except
   the final `Enable` PR. Every intermediate PR is a safe no-op in production — nothing calls into your
   new code until you flip the switch. This means you can merge the stack incrementally without waiting
   for the whole feature to be done.
4. **One interface per PR.** The interfaces form a ladder (see
   [The big picture](./SUBSCRIBE_ONBOARDING.md#the-big-picture)); add one rung per PR.
5. **Only build what the provider needs.** `RegisterSubscribeConnector` and
   `SubscriptionMaintainerConnector` are **provider-specific** — most providers skip them. Don't add a
   PR for a rung the provider doesn't require.

---

## The stack

PRs are stacked bottom-to-top and mirror the interface ladder. The base merges first; `Enable` merges
last.

```
  Enable the provider          flip Support.Subscribe + SubscribeByAPI on   ← merge last (top)
        ▲ stacked on
  Maintenance                  SubscriptionMaintainerConnector              (if needed)
        ▲ stacked on
  Registration                 RegisterSubscribeConnector                   (if needed)
        ▲ stacked on
  Subscribe / Update / Delete  SubscribeConnector
        ▲ stacked on
  Verification                 WebhookVerifierConnector
        ▲ stacked on
  Metadata scaffold (gated off) + factory wiring               ← base, merge first (bottom)
```

| # | PR | Adds | Required? |
|---|----|------|-----------|
| 1 | Metadata scaffold + factory wiring | provider metadata (gated off), connector registration | ✅ |
| 2 | Verification | `WebhookVerifierConnector` + event types | ✅ |
| 3 | Subscribe / Update / Delete | `SubscribeConnector` | ✅ |
| 4 | Registration | `RegisterSubscribeConnector` | ⬜ if needed |
| 5 | Maintenance | `SubscriptionMaintainerConnector` | ⬜ if needed |
| 6 | Enable the provider | flips the gate on | ✅ (last) |

---

## PR-by-PR

Each PR below lists its **scope**, the **files** it typically touches, and a **checklist**. Link the
relevant section of [`SUBSCRIBE_ONBOARDING.md`](./SUBSCRIBE_ONBOARDING.md) in your PR description so
reviewers have the context.

### PR 1 — Metadata scaffold + factory wiring *(base)*

**Scope.** Establish the provider in the catalog with subscribe **gated off**, and make sure the
connector is constructible. Zero behavioral change.

**Files.**
- `providers/<provider>.go` — declare `SubscribeRequirements`, but keep `Support.Subscribe: false` and
  `SubscribeByAPI: new(false)`.
- `connector/new.go` — add the `new<Provider>Connector` constructor + map entry **only if the provider
  is brand-new** (most providers already have one for read/write).

**Checklist.**
- [ ] `Support.Subscribe` is `false` and `SubscribeByAPI` is `new(false)` (or unset).
- [ ] `SubscribeRequirements` reflects the provider's intended shape (`Registration` / `PostProcess` /
      `Maintenance` set only if applicable).
- [ ] `make build` / `go build ./...` passes; no behavior change.

→ Reference: [Provider metadata](./SUBSCRIBE_ONBOARDING.md#provider-metadata),
[Factory wiring](./SUBSCRIBE_ONBOARDING.md#factory-wiring).

### PR 2 — Verification

**Scope.** Webhook signature verification and the typed events the caller will parse.

**Files.** `providers/<provider>/subscribeEvent.go` (or similar) — `VerifyWebhookMessage`, the provider
`VerificationParams` struct, and the event type(s) implementing `SubscriptionEvent`
(+ `SubscriptionUpdateEvent` / `CollapsedSubscriptionEvent` as needed).

**Checklist.**
- [ ] `var _ connectors.WebhookVerifierConnector = &Connector{}` assertion present.
- [ ] `VerifyWebhookMessage` validates the signature with constant-time compare (`hmac.Equal`); returns
      `false` (not an error) for untrusted requests.
- [ ] Each event method (`EventType`, `ObjectName`, `RecordId`, `EventTimeStampNano`, `RawEventName`,
      `Workspace`, `PreLoadData`) implemented and unit-tested against a captured real payload.
- [ ] Unit tests cover valid / invalid / missing-signature cases.

→ Reference: [Verification](./SUBSCRIBE_ONBOARDING.md#verification),
[Event types](./SUBSCRIBE_ONBOARDING.md#event-types).

### PR 3 — Subscribe / Update / Delete

**Scope.** Programmatic subscription lifecycle.

**Files.** `providers/<provider>/subscribe.go` — `Subscribe`, `UpdateSubscription`,
`DeleteSubscription`, `EmptySubscriptionParams`, `EmptySubscriptionResult`, plus the provider
`Request` / `Result` structs. Plus a manual harness at `test/<provider>/subscribe/subscribe.go`.

**Checklist.**
- [ ] `var _ connectors.SubscribeConnector = &Connector{}` assertion present.
- [ ] `Subscribe` rolls back partially-created subscriptions on error.
- [ ] `UpdateSubscription` reconciles `previousResult` → desired state.
- [ ] `DeleteSubscription` tears down everything in `previousResult`.
- [ ] `Empty*` return instances with the provider-specific `.Request` / `.Result` populated.
- [ ] Manual harness runs against a real sandbox and a webhook is received + verified end-to-end.

→ Reference: [Interface reference](./SUBSCRIBE_ONBOARDING.md#interface-reference),
[Testing](./SUBSCRIBE_ONBOARDING.md#testing).

### PR 4 — Registration *(provider-specific, if needed)*

**Skip this PR unless** the provider needs a one-time, installation-level setup shared by all object
subscriptions (Salesforce is the canonical case).

**Scope.** `RegisterSubscribeConnector`.

**Files.** `providers/<provider>/register.go` (or similar) — `Register`, `DeleteRegistration`,
`EmptyRegistrationParams`, `EmptyRegistrationResult`, plus rollback. Set
`SubscribeRequirements.Registration: new(true)` (and `PostProcess: new(true)` if a third-party setup
step is involved).

**Checklist.**
- [ ] `Register` rolls back its own partial work on failure and sets `Status` accordingly
      (`Success` / `Failed` / `FailedToRollback`).
- [ ] `DeleteRegistration` tears down resources in reverse order.
- [ ] If `PostProcess` applies, `RegistrationResult.Result` carries the data the post-processor needs.

→ Reference: [Registration](./SUBSCRIBE_ONBOARDING.md#registration-salesforce-example).

### PR 5 — Maintenance *(provider-specific, if needed)*

**Skip this PR unless** the provider's subscriptions/watches expire after a TTL and must be renewed.

**Scope.** `SubscriptionMaintainerConnector`.

**Files.** `providers/<provider>/...` — `RunScheduledMaintenance`. Set
`SubscribeRequirements.Maintenance: new(true)`.

**Checklist.**
- [ ] `RunScheduledMaintenance` renews the subscription in `previousResult` and returns refreshed state.

→ Reference: [Maintenance](./SUBSCRIBE_ONBOARDING.md#maintenance).

### PR 6 — Enable the provider *(top)*

**Scope.** A one-line switch that turns the provider on.

**Files.** `providers/<provider>.go` — flip `Support.Subscribe: true` and `SubscribeByAPI: new(true)`.

**Checklist.**
- [ ] All prerequisite PRs (verification, subscribe, and any needed registration/maintenance) are
      merged.
- [ ] End-to-end verified in a sandbox.
- [ ] Trivial to revert (single-line change).

> The caller also needs a small configuration change to consume the new provider — supplying the
> per-installation verification params and subscribe request payload your connector expects. That
> change lives outside this repo and lands after this stack.

---

## Why gate, and on which flags

The caller activates a provider from its metadata:

- The **API subscribe** path is gated on `SubscribeRequirements.SubscribeByAPI`.
- The **manual / UI-subscription** path is gated on `Support.Subscribe`.

Keeping **both** off until PR 6 means none of the intermediate PRs can affect production, even after
they merge. That's what makes incremental merging safe. Flip both on together in the final PR. (The
*requirement* flags — `Registration` / `PostProcess` / `Maintenance` — are only consulted once subscribe
is active, so declaring them earlier is harmless.)

---

## Managing the stack

Create the branches bottom-up so each is based on the previous one:

```
main
 └─ subscribe/<provider>/metadata        (PR 1)
     └─ subscribe/<provider>/verify       (PR 2)
         └─ subscribe/<provider>/subscribe (PR 3)
             └─ ... registration / maintenance if needed
                 └─ subscribe/<provider>/enable (PR 6)
```

- If you use **Graphite**, this is a normal `gt create` stack; submit with `gt submit --stack`.
- With plain git, branch each PR off the previous branch and rebase the upstack when a lower PR changes.
- When a lower PR gets review changes, restack everything above it before re-pushing.
- You don't have to wait for PR 1 to merge before opening PR 2 — open the whole stack and let review
  proceed in parallel; merge bottom-up as each is approved.

---

## PR description checklist (copy into each PR)

```markdown
## Subscribe onboarding — <Provider> — [PR N: <name>]

Part of the subscribe onboarding stack for `<provider>`. See SUBSCRIBE_PR_PROCESS.md.

- [ ] Scope limited to this stack rung (one interface / concern)
- [ ] Provider remains gated off (Support.Subscribe / SubscribeByAPI unchanged) — except the Enable PR
- [ ] Compile-time interface assertion added (if this PR adds an interface)
- [ ] Unit tests added/updated
- [ ] Manual sandbox verification (where applicable)
- [ ] Linked the relevant SUBSCRIBE_ONBOARDING.md section
```

---

## Quick reference

| Want to… | Go to |
|----------|-------|
| Understand the interfaces & types | [`SUBSCRIBE_ONBOARDING.md`](./SUBSCRIBE_ONBOARDING.md) |
| See a worked example | [acme example](./SUBSCRIBE_ONBOARDING.md#a-worked-example-adding-acme) |
| Know what each PR contains | [PR-by-PR](#pr-by-pr) above |
| Know the merge order | [The stack](#the-stack) above |
