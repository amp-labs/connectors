# PR 3 — Subscribe / Update / Delete (`SubscribeConnector`)

> Part of the [Subscribe Onboarding PR Process](../../SUBSCRIBE_PR_PROCESS.md). Implementation
> reference: [`SUBSCRIBE_ONBOARDING.md`](../../SUBSCRIBE_ONBOARDING.md).

**Required.** Stacks on [PR 2](./pr-2-verification.md).

## Goal

Implement the programmatic subscription lifecycle: create, update, and delete subscriptions in the
provider.

## What you implement

`SubscribeConnector` on your `*Connector`:

```go
Subscribe(ctx, params common.SubscribeParams) (*common.SubscriptionResult, error)
UpdateSubscription(ctx, params common.SubscribeParams, previous *common.SubscriptionResult) (*common.SubscriptionResult, error)
DeleteSubscription(ctx, previous common.SubscriptionResult) error
EmptySubscriptionParams() *common.SubscribeParams
EmptySubscriptionResult() *common.SubscriptionResult
```

Plus your provider-specific `Request` / `Result` structs, and a manual test harness.

## Files

- `providers/<provider>/subscribe.go` — the five methods + structs.
- `test/<provider>/subscribe/subscribe.go` — a small `main` harness for end-to-end testing.

## Steps

1. Add `var _ connectors.SubscribeConnector = &Connector{}`.
2. Define `Request` (carried in `params.Request`) and `Result` (stored in
   `SubscriptionResult.Result`).
3. `Subscribe`: translate `params.SubscriptionEvents` (objects → event types) into provider API calls.
   **Roll back** anything created if a later call fails.
4. `UpdateSubscription`: diff `previous` against the desired `params` and reconcile (create the new,
   delete the gone).
5. `DeleteSubscription`: remove everything identified by `previous.Result`.
6. `Empty*`: return instances with the provider-specific `.Request` / `.Result` populated so stored DB
   state deserializes into your types.

## Example

```go
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
    // For each object+event in params.SubscriptionEvents, call the provider API.
    // Track successes; roll back on partial failure. Return the actual state.
}
```

See [`providers/salesloft/subscribe.go`](../../providers/salesloft/subscribe.go) and
[`providers/outreach/subscribe.go`](../../providers/outreach/subscribe.go) for the parallel-create-with-
rollback pattern, and [`test/outreach/subscribe/subscribe.go`](../../test/outreach/subscribe/subscribe.go)
for the harness.

## Checklist

- [ ] `var _ connectors.SubscribeConnector = &Connector{}` assertion present.
- [ ] `Subscribe` rolls back partially-created subscriptions on error.
- [ ] `UpdateSubscription` reconciles `previous` → desired state.
- [ ] `DeleteSubscription` tears down everything in `previous.Result`.
- [ ] `Empty*` populate the provider-specific `.Request` / `.Result`.
- [ ] Manual harness runs against a real sandbox; a triggered change yields a verified webhook
      end-to-end.

## Reviewer focus

- Rollback correctness on partial failure (no orphaned provider-side subscriptions).
- Update diffing handles both additions and removals.
- `SubscriptionResult.ObjectEvents` reflects the actual post-operation state.

## Reference

- [Interface reference](../../SUBSCRIBE_ONBOARDING.md#interface-reference)
- [Core types](../../SUBSCRIBE_ONBOARDING.md#core-types)
- [Testing](../../SUBSCRIBE_ONBOARDING.md#testing)
