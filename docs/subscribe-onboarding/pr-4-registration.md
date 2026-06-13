# PR 4 — Registration (`RegisterSubscribeConnector`) *(provider-specific, if needed)*

> Part of the [Subscribe Onboarding PR Process](../../SUBSCRIBE_PR_PROCESS.md). Shared concepts:
> [`SUBSCRIBE_ONBOARDING.md`](../../SUBSCRIBE_ONBOARDING.md).

> **Provider-specific — implement only if needed.** Most providers do not need registration. Add this
> PR only when the provider requires a one-time, installation-level setup step shared by all object
> subscriptions (Salesforce → AWS EventBridge is the canonical case).

Builds on [PR 3](./pr-3-subscribe-update-delete.md). It has **no dependency** on
[PR 5 (Maintenance)](./pr-5-maintenance.md) — neither interface extends the other — so do either, both,
or neither, in any order. You can keep them as separate branches off PR 3 or stack them; either is fine.

## Goal

Implement the one-time registration that creates the shared infrastructure `Subscribe` then hangs each
object subscription off of.

## What you implement

`RegisterSubscribeConnector` (defined in [`connectors.go`](../../connectors.go)) on your `*Connector`:

```go
Register(ctx context.Context, params common.SubscriptionRegistrationParams) (*common.RegistrationResult, error)
DeleteRegistration(ctx context.Context, previousResult common.RegistrationResult) error
EmptyRegistrationParams() *common.SubscriptionRegistrationParams
EmptyRegistrationResult() *common.RegistrationResult
```

`Register` is a one-time per-installation operation that creates shared infrastructure (`Subscribe`
later hangs each object subscription off it). It must roll back its own partial work on failure and set
`Status` accordingly.

And set `SubscribeRequirements.Registration: new(true)` in [PR 1](./pr-1-provider-info.md)'s metadata
(and `PostProcess: new(true)` if a third-party setup step is involved).

Files: `providers/<provider>/register.go` (or similar) and `providers/<provider>.go` (flag).

## Example

From [`providers/salesforce/register.go`](../../providers/salesforce/register.go):

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
order.

## PostProcess

`PostProcess` work (e.g. wiring AWS EventBridge after Salesforce subscribes) is performed **outside the
connector** by the caller. The connector's only obligation is to **return the data the post-processor
needs** in `RegistrationResult.Result` (for Salesforce, the `EventChannel` id, etc.). Set
`SubscribeRequirements.PostProcess: new(true)`; there's no connector method to implement, so fold the
flag into this PR.

## Checklist

- [ ] `Register` rolls back its own partial work on failure and sets `Status` correctly.
- [ ] `DeleteRegistration` tears resources down in reverse order.
- [ ] `EmptyRegistrationParams` / `EmptyRegistrationResult` populate the provider-specific structs.
- [ ] `Registration: new(true)` set (+ `PostProcess: new(true)` if applicable), each with a code
      comment linking the provider docs that justify it.
- [ ] `RegistrationResult.Result` carries everything `Subscribe` and any post-processor need.

## Reviewer focus

- Rollback ordering and idempotency (safe to retry).
- `Status` values returned on each failure path.

## Reference

- [Core types](../../SUBSCRIBE_ONBOARDING.md#core-types)
- [`providers/salesforce/register.go`](../../providers/salesforce/register.go)
