# PR 4 — Registration (`RegisterSubscribeConnector`) *(provider-specific, if needed)*

> Part of the [Subscribe Onboarding PR Process](../../SUBSCRIBE_PR_PROCESS.md). Implementation
> reference: [`SUBSCRIBE_ONBOARDING.md`](../../SUBSCRIBE_ONBOARDING.md).

**Optional — skip this PR unless** the provider needs a one-time, installation-level setup step shared
by all object subscriptions (Salesforce → AWS EventBridge is the canonical case). Most providers do not.

Stacks on [PR 3](./pr-3-subscribe-update-delete.md).

## Goal

Implement the one-time registration that creates the shared infrastructure `Subscribe` then hangs each
object subscription off of.

## What you implement

`RegisterSubscribeConnector` on your `*Connector`:

```go
Register(ctx, params common.SubscriptionRegistrationParams) (*common.RegistrationResult, error)
DeleteRegistration(ctx, previous common.RegistrationResult) error
EmptyRegistrationParams() *common.SubscriptionRegistrationParams
EmptyRegistrationResult() *common.RegistrationResult
```

And set `SubscribeRequirements.Registration: new(true)` (and `PostProcess: new(true)` if a third-party
setup step is involved).

## Files

- `providers/<provider>/register.go` — the four methods + rollback + `RegistrationParams` / result data.
- `providers/<provider>.go` — flip the `Registration` (and maybe `PostProcess`) requirement flag.

## Steps

1. Define `RegistrationParams` (carried in `SubscriptionRegistrationParams.Request`) and the result data
   struct (stored in `RegistrationResult.Result`).
2. `Register`: validate params, create the resources, and on failure **roll back your own partial work**
   and set `Status` (`Success` / `Failed` / `FailedToRollback`).
3. `DeleteRegistration`: tear resources down in reverse creation order.
4. `Empty*`: return instances with the provider-specific `.Request` / `.Result` populated.

## Example

```go
type RegistrationParams struct {
    UniqueRef             string `json:"uniqueRef"             validate:"required"`
    Label                 string `json:"label"                 validate:"required"`
    AwsNamedCredentialArn string `json:"awsNamedCredentialArn" validate:"required"`
}

func (c *Connector) Register(
    ctx context.Context, params common.SubscriptionRegistrationParams,
) (*common.RegistrationResult, error) {
    result, err := c.register(ctx, params.Request.(*RegistrationParams))
    if err != nil {
        if rb := c.rollbackRegister(ctx, result); rb != nil {
            return &common.RegistrationResult{Status: common.RegistrationStatusFailedToRollback},
                errors.Join(rb, err)
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

See [`providers/salesforce/register.go`](../../providers/salesforce/register.go).

## PostProcess note

If the provider needs a third-party setup step the connector can't perform itself (e.g. AWS
EventBridge), that work is done **outside the connector** by the caller. Your only obligation is to
return the data the post-processor needs in `RegistrationResult.Result`. Set
`SubscribeRequirements.PostProcess: new(true)`; there's no connector method to implement, so fold it
into this PR.

## Checklist

- [ ] `Register` rolls back its own partial work on failure and sets `Status` correctly.
- [ ] `DeleteRegistration` tears down in reverse order.
- [ ] `Registration: new(true)` set (+ `PostProcess: new(true)` if applicable).
- [ ] `RegistrationResult.Result` carries everything `Subscribe` and any post-processor need.

## Reviewer focus

- Rollback ordering and idempotency (safe to retry).
- `Status` values returned on each failure path.

## Reference

- [Registration (Salesforce example)](../../SUBSCRIBE_ONBOARDING.md#registration-salesforce-example)
