# PR 3 — Subscribe / Update / Delete (`SubscribeConnector`)

> Part of the [Subscribe Onboarding PR Process](../../SUBSCRIBE_PR_PROCESS.md). Shared concepts:
> [`SUBSCRIBE_ONBOARDING.md`](../../SUBSCRIBE_ONBOARDING.md).

**Required.** Stacks on [PR 2](./pr-2-verification.md).

## Goal

Implement the programmatic subscription lifecycle: create, update, and delete subscriptions in the
provider.

## What you implement

`SubscribeConnector` (defined in [`connectors.go`](../../connectors.go)) on your `*Connector`:

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

Plus your provider-specific `Request` / `Result` structs, and a manual test harness.

Files: `providers/<provider>/subscribe.go` and `test/<provider>/subscribe/subscribe.go`.

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

func (c *Connector) UpdateSubscription(
    ctx context.Context, params common.SubscribeParams, previous *common.SubscriptionResult,
) (*common.SubscriptionResult, error) { /* reconcile previous → desired */ }

func (c *Connector) DeleteSubscription(
    ctx context.Context, previous common.SubscriptionResult,
) error { /* delete everything in previous.Result */ }
```

The per-installation request payload (`SubscriptionRequest{WebhookEndpoint, Secret}`) is *built by the
caller* and handed to you in `params.Request` — the caller constructs the webhook endpoint URL and
secret. You only define the struct and consume it.

See [`providers/salesloft/subscribe.go`](../../providers/salesloft/subscribe.go) and
[`providers/outreach/subscribe.go`](../../providers/outreach/subscribe.go) for the parallel-create-with-
rollback pattern.

## Testing

1. **Compile-time assertion** — `var _ connectors.SubscribeConnector = &Connector{}` so a missing method
   fails the build.
2. **Manual end-to-end harness** — add `test/<provider>/subscribe/subscribe.go`, a small `main` that
   loads creds, builds the connector, and calls `Subscribe` against a real sandbox. Model it on
   [`test/outreach/subscribe/subscribe.go`](../../test/outreach/subscribe/subscribe.go):

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

   Trigger a change in the provider sandbox and confirm the webhook is received and verifies end-to-end.

See [`CONTRIBUTING.md`](../../CONTRIBUTING.md) for credential setup (`creds.json`) and the dev
environment.

## Checklist

- [ ] `var _ connectors.SubscribeConnector = &Connector{}` assertion present.
- [ ] `Subscribe` rolls back partially-created subscriptions on error.
- [ ] `UpdateSubscription` reconciles `previous` → desired state (handles additions and removals).
- [ ] `DeleteSubscription` tears down everything in `previous.Result`.
- [ ] `Empty*` populate the provider-specific `.Request` / `.Result`.
- [ ] `SubscriptionResult.ObjectEvents` reflects the actual post-operation state.
- [ ] Manual harness runs against a real sandbox; a triggered change yields a verified webhook
      end-to-end.

## Reviewer focus

- Rollback correctness on partial failure (no orphaned provider-side subscriptions).
- Update diffing handles both additions and removals.
- `Empty*` types match what's persisted and re-read.

## Reference

- [The big picture](../../SUBSCRIBE_ONBOARDING.md#the-big-picture) · [Core types](../../SUBSCRIBE_ONBOARDING.md#core-types)
- [`providers/salesloft/subscribe.go`](../../providers/salesloft/subscribe.go),
  [`providers/outreach/subscribe.go`](../../providers/outreach/subscribe.go),
  [`test/outreach/subscribe/subscribe.go`](../../test/outreach/subscribe/subscribe.go)
