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

PRs 1–3 are a linear stack and mirror the interface ladder. PRs 4 and 5 are **optional** and have
**no dependency on each other** — `RegisterSubscribeConnector` and `SubscriptionMaintainerConnector`
each extend `SubscribeConnector` (not each other). Because neither depends on the other, the order is
up to you: branch each off PR 3, or stack them in a line if that fits your workflow — either is fine.
`Enable` merges last, once PR 3 and any needed PR 4 / PR 5 are in.

```
  Enable the provider (PR 6)   flip Support.Subscribe + SubscribeByAPI on
        ▲ merge last, after PR 3 (+ PR 4 / PR 5 if used)
        │
  Registration (PR 4)          Maintenance (PR 5)     ← optional; no dependency between
  RegisterSubscribeConnector   SubscriptionMaintainer    them — branch each off PR 3, or
  (if needed)                  (if needed)               stack them; either is fine
        └──────────── both build on ─────────────┘
                          │
  Subscribe / Update / Delete (PR 3)   SubscribeConnector
        ▲ stacked on
  Verification (PR 2)                  WebhookVerifierConnector
        ▲ stacked on
  ProviderInfo + factory wiring (PR 1, gated off)          ← base, merge first
```

| # | PR | Adds | Required? |
|---|----|------|-----------|
| 1 | ProviderInfo + Factory wiring | subscribe metadata (gated off); factory entry if brand-new | ✅ |
| 2 | Verification | `WebhookVerifierConnector` + event types | ✅ |
| 3 | Subscribe / Update / Delete | `SubscribeConnector` | ✅ |
| 4 | Registration | `RegisterSubscribeConnector` | ⬜ if needed |
| 5 | Maintenance | `SubscriptionMaintainerConnector` | ⬜ if needed |
| 6 | Enable the provider | flips the gate on | ✅ (last) |

---

## PR-by-PR

Each PR has its own focused guide — open the one you're writing. They share a structure: goal, what to
implement, files, step-by-step, an example, a checklist, and reviewer focus. Link the relevant
[`SUBSCRIBE_ONBOARDING.md`](./SUBSCRIBE_ONBOARDING.md) section in your PR description too.

| # | PR | Guide | Required? |
|---|----|-------|-----------|
| 1 | ProviderInfo + Factory wiring | [pr-1-provider-info.md](./docs/subscribe-onboarding/pr-1-provider-info.md) | ✅ |
| 2 | Verification | [pr-2-verification.md](./docs/subscribe-onboarding/pr-2-verification.md) | ✅ |
| 3 | Subscribe / Update / Delete | [pr-3-subscribe-update-delete.md](./docs/subscribe-onboarding/pr-3-subscribe-update-delete.md) | ✅ |
| 4 | Registration | [pr-4-registration.md](./docs/subscribe-onboarding/pr-4-registration.md) | ⬜ if needed |
| 5 | Maintenance | [pr-5-maintenance.md](./docs/subscribe-onboarding/pr-5-maintenance.md) | ⬜ if needed |
| 6 | Enable the provider | [pr-6-enable.md](./docs/subscribe-onboarding/pr-6-enable.md) | ✅ (last) |

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

Branch PRs 1→2→3 in a line. PRs 4 and 5, if needed, both build on PR 3; since they have no dependency
on each other you can branch each off PR 3 (siblings, any order) **or** stack them in a line — whatever
fits your workflow. Enable branches last, once PR 3 and any needed PR 4 / PR 5 are merged. One possible
layout:

```
main
 └─ subscribe/<provider>/provider-info     (PR 1)
     └─ subscribe/<provider>/verify         (PR 2)
         └─ subscribe/<provider>/subscribe   (PR 3)
             ├─ subscribe/<provider>/registration  (PR 4, if needed) ┐ optional; no dependency
             ├─ subscribe/<provider>/maintenance    (PR 5, if needed) ┘ between them
             └─ subscribe/<provider>/enable          (PR 6, merge last)
```

- PRs 4 and 5 have no dependency on each other — implement either, both, or neither, in any order.
  Keeping them as siblings off PR 3 (shown above) or stacking them one-on-the-other are both fine.
- If you use **Graphite**, PRs 1–3 are a normal `gt create` stack; PRs 4/5 are branches off PR 3.
  Submit with `gt submit --stack`.
- With plain git, branch each PR off its parent and rebase the upstack when a lower PR changes.
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
- [ ] Any SubscribeRequirements flag set to new(true) has a code comment linking the provider docs
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
| See a worked example | the **Example** section in each per-PR guide (e.g. [PR 3](./docs/subscribe-onboarding/pr-3-subscribe-update-delete.md#example)) |
| Know what each PR contains | [PR-by-PR](#pr-by-pr) above |
| Know the merge order | [The stack](#the-stack) above |
