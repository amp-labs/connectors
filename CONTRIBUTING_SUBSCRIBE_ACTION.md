# Contributing a Subscribe Action

How to **ship** subscribe support for a new provider as a clean stack of pull requests.

This doc is about *process*: how to slice the work, what each PR contains, the order to merge them, and
what reviewers should check. For the *implementation* details (interfaces, types, event parsing,
verification, metadata, factory wiring, worked examples), see the companion reference:
[**`SUBSCRIBE_REFERENCES.md`**](./SUBSCRIBE_REFERENCES.md).

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
4. **One interface per PR.** Each PR adds a single interface's methods (verification, registration,
   subscribe, maintenance). See [The big picture](./SUBSCRIBE_REFERENCES.md#the-big-picture) for how the
   interfaces relate (note: PR order is sequenced by dependency, not by the interface ladder — see
   [The stack](#the-stack)).
5. **Only build what the provider needs.** `RegisterSubscribeConnector` and
   `SubscriptionMaintainerConnector` are **provider-specific** — most providers skip them. Don't add a
   PR for a rung the provider doesn't require.

---

## The stack

The stack is **linear** — each PR builds on the one below it. Registration (PR 3) and Maintenance
(PR 5) are **optional**; include them only if the provider needs them. Their positions are fixed by
dependency: **Registration comes before Subscribe** (it creates the shared infrastructure `Subscribe`
consumes), and **Maintenance comes after Subscribe** (it renews what Subscribe created). `Enable` is
always last.

```
  Enable the provider (PR 6)             flip Support.Subscribe + SubscribeByAPI on   ← merge last
        ▲
  Maintenance (PR 5, if needed)          SubscriptionMaintainerConnector — renews after subscribe
        ▲
  Subscribe / Update / Delete (PR 4)     SubscribeConnector
        ▲
  Registration (PR 3, if needed)         RegisterSubscribeConnector — creates infra Subscribe uses
        ▲
  Verification (PR 2)                    WebhookVerifierConnector
        ▲
  ProviderInfo + factory wiring (PR 1, gated off)          ← base, merge first
```

> The PR order is not the same as the interface ladder. `RegisterSubscribeConnector` *embeds*
> `SubscribeConnector`, yet registration is *sequenced before* subscribe because its result is an input
> to `Subscribe`. (The full `RegisterSubscribeConnector` compile-time assertion is therefore added once
> Subscribe lands in PR 4.)

| # | PR | Adds | Required? |
|---|----|------|-----------|
| 1 | ProviderInfo + Factory wiring | subscribe metadata (gated off); factory entry if brand-new | ✅ |
| 2 | Verification | `WebhookVerifierConnector` + event types | ✅ |
| 3 | Registration | `RegisterSubscribeConnector` — before Subscribe | ⬜ if needed |
| 4 | Subscribe / Update / Delete | `SubscribeConnector` | ✅ |
| 5 | Maintenance | `SubscriptionMaintainerConnector` — after Subscribe | ⬜ if needed |
| 6 | Enable the provider | flips the gate on | ✅ (last) |

---

## PR-by-PR

Each PR has its own focused guide — open the one you're writing. They share a structure: goal, what to
implement, files, step-by-step, an example, a checklist, and reviewer focus. Link the relevant
[`SUBSCRIBE_REFERENCES.md`](./SUBSCRIBE_REFERENCES.md) section in your PR description too.

| # | PR | Guide | Required? |
|---|----|-------|-----------|
| 1 | ProviderInfo + Factory wiring | [pr-1-provider-info.md](./docs/subscribe-onboarding/pr-1-provider-info.md) | ✅ |
| 2 | Verification | [pr-2-verification.md](./docs/subscribe-onboarding/pr-2-verification.md) | ✅ |
| 3 | Registration | [pr-3-registration.md](./docs/subscribe-onboarding/pr-3-registration.md) | ⬜ if needed |
| 4 | Subscribe / Update / Delete | [pr-4-subscribe-update-delete.md](./docs/subscribe-onboarding/pr-4-subscribe-update-delete.md) | ✅ |
| 5 | Maintenance | [pr-5-maintenance.md](./docs/subscribe-onboarding/pr-5-maintenance.md) | ⬜ if needed |
| 6 | Enable the provider | [pr-6-enable.md](./docs/subscribe-onboarding/pr-6-enable.md) | ✅ (last) |

## Why gate, and on which flags

The caller activates a provider from its metadata:

- `Support.Subscribe` is the **master gate** — it must be `true` for the provider to subscribe at all,
  via API or manual/UI. (Today the server keys the *API* path off `SubscribeByAPI`; enforcing
  `Support.Subscribe` as the master gate for the API path too is planned server-side work.)
- `SubscribeRequirements.SubscribeByAPI` says **whether the programmatic (API) approach is available**:
  `true` → subscribe via the connector's API; with `Support.Subscribe` on but `SubscribeByAPI` off →
  the provider is configured manually in its UI (UI Subscription only).

Keep **both `Support.Subscribe` and `SubscribeByAPI` off** for the entire stack so none of the
intermediate PRs can affect production, even after they merge — that's what makes incremental merging
safe. In the final PR, flip `Support.Subscribe` on (plus `SubscribeByAPI` for API providers). (The
*requirement* flags — `Registration` / `PostProcess` / `Maintenance` — are only consulted once subscribe
is active, so declaring them earlier is harmless.)

---

## Managing the stack

Branch each PR off the one below it: PR 1 → 2 → 3 → 4 → 5 → 6. Skip PR 3 (Registration) and/or PR 5
(Maintenance) if the provider doesn't need them — the chain just closes up.

```
main
 └─ subscribe/<provider>/provider-info       (PR 1)
     └─ subscribe/<provider>/verify           (PR 2)
         └─ subscribe/<provider>/registration  (PR 3, if needed)
             └─ subscribe/<provider>/subscribe   (PR 4)
                 └─ subscribe/<provider>/maintenance  (PR 5, if needed)
                     └─ subscribe/<provider>/enable     (PR 6, merge last)
```

- Skip PR 3 / PR 5 entirely when the provider doesn't need registration / maintenance; the next PR
  just branches off whatever precedes it.
- If you use **Graphite**, this is a normal `gt create` stack; submit with `gt submit --stack`.
- With plain git, branch each PR off its parent and rebase the upstack when a lower PR changes.
- When a lower PR gets review changes, restack everything above it before re-pushing.
- You don't have to wait for PR 1 to merge before opening PR 2 — open the whole stack and let review
  proceed in parallel; merge bottom-up as each is approved.

---

## PR description checklist (copy into each PR)

```markdown
## Subscribe Action — <Provider> — [PR N: <name>]

Part of the Subscribe Action stack for `<provider>`. See CONTRIBUTING_SUBSCRIBE_ACTION.md.

- [ ] Scope limited to this stack rung (one interface / concern)
- [ ] Provider remains gated off (Support.Subscribe / SubscribeByAPI unchanged) — except the Enable PR
- [ ] Any SubscribeRequirements flag set to new(true) has a code comment linking the provider docs
- [ ] Compile-time interface assertion added (if this PR adds an interface)
- [ ] Unit tests added/updated
- [ ] Manual sandbox verification (where applicable)
- [ ] Linked the relevant SUBSCRIBE_REFERENCES.md section
```

---

## Quick reference

| Want to… | Go to |
|----------|-------|
| Understand the interfaces & types | [`SUBSCRIBE_REFERENCES.md`](./SUBSCRIBE_REFERENCES.md) |
| See a worked example | the **Example** section in each per-PR guide (e.g. [PR 4](./docs/subscribe-onboarding/pr-4-subscribe-update-delete.md#example)) |
| Know what each PR contains | [PR-by-PR](#pr-by-pr) above |
| Know the merge order | [The stack](#the-stack) above |
