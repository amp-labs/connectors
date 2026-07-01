# PR 3 — Registration (`RegisterSubscribeConnector`) *(provider-specific, if needed)*

> Part of [Contributing a Subscribe Action](../../CONTRIBUTING_SUBSCRIBE_ACTION.md). Shared concepts:
> [`SUBSCRIBE_REFERENCES.md`](../../SUBSCRIBE_REFERENCES.md).

> **Provider-specific — implement only if needed.** Most providers do not need registration. Add this
> PR only when the provider requires a one-time, installation-level setup step shared by all object
> subscriptions (Salesforce is the canonical case).

Builds on [PR 2](./pr-2-verification.md). When a provider needs registration, it comes **before**
[Subscribe (PR 4)](./pr-4-subscribe-update-delete.md): `Subscribe` consumes the registration result
(`params.RegistrationResult`), so that shared resource must exist first. (Note: the
`connectors.RegisterSubscribeConnector` interface *embeds* `SubscribeConnector`, so the full
compile-time assertion is satisfied once Subscribe lands in PR 4 — this PR adds the `Register` /
`DeleteRegistration` methods.)

## Goal

Implement the one-time registration that creates a **shared resource across objects** — the thing
`Subscribe` then hangs each object subscription off of.

## What you implement

`RegisterSubscribeConnector` (defined in [`connectors.go`](../../connectors.go)) on your `*Connector`:

```go
Register(ctx context.Context, params common.SubscriptionRegistrationParams) (*common.RegistrationResult, error)
DeleteRegistration(ctx context.Context, previousResult common.RegistrationResult) error
EmptyRegistrationParams() *common.SubscriptionRegistrationParams
EmptyRegistrationResult() *common.RegistrationResult
```

`Register` is a one-time per-installation operation that creates a shared resource used across all
objects (`Subscribe` later hangs each object subscription off it). It must roll back its own partial
work on failure and set `Status` accordingly.

And set `SubscribeRequirements.Registration: new(true)` in [PR 1](./pr-1-provider-info.md)'s
`ProviderInfo`.

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

## Serialization

The caller **persists your `RegistrationResult` and reads it back later** (for `Subscribe` and for
deletion), so it serializes/deserializes the value you place in the `Result any` field — and likewise
the `Request any` in `SubscriptionRegistrationParams`. Deserialization targets the concrete type your
`EmptyRegistrationResult()` / `EmptyRegistrationParams()` return, so those structs must round-trip
cleanly:

- **Export every field** — unexported fields are silently dropped. Add JSON tags to control the serialized names (recommended).
- **Prefer Go native types** (`string`, `int`, `bool`, `time.Time`, and slices/maps/nested structs of
  those). Avoid `any`/`interface{}`, function values, channels, and types that need custom
  (un)marshaling; they don't survive the round trip reliably.
- Anything a later step needs — the registration ref, created-resource ids — must live in `Result`; if
  it isn't serialized, it's gone.

## Checklist

- [ ] `Register` rolls back its own partial work on failure and sets `Status` correctly.
- [ ] `DeleteRegistration` tears resources down in reverse order.
- [ ] `EmptyRegistrationParams` / `EmptyRegistrationResult` populate the provider-specific structs.
- [ ] `RegistrationParams` / result structs round-trip through serialization: exported fields with JSON
      tags, Go native types.
- [ ] `Registration: new(true)` set in `ProviderInfo` (PR 1), with a code comment linking the provider
      docs that justify it.
- [ ] `RegistrationResult.Result` carries everything `Subscribe` needs.

## Reviewer focus

- Rollback ordering and idempotency (safe to retry).
- `Status` values returned on each failure path.

## Reference

- [Core types](../../SUBSCRIBE_REFERENCES.md#core-types)
- [`providers/salesforce/register.go`](../../providers/salesforce/register.go)
