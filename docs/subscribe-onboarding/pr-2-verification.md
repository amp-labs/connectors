# PR 2 — Verification (`WebhookVerifierConnector`)

> Part of [Contributing a Subscribe Action](../../CONTRIBUTING_SUBSCRIBE_ACTION.md). Shared concepts:
> [`SUBSCRIBE_REFERENCES.md`](../../SUBSCRIBE_REFERENCES.md).

**Required.** Stacks on [PR 1](./pr-1-provider-info.md).

## Goal

Implement webhook signature verification and the typed events the caller parses out of incoming
webhooks.

## What you implement

`WebhookVerifierConnector` (defined in [`connectors.go`](../../connectors.go)):

```go
VerifyWebhookMessage(
    ctx context.Context,
    request *common.WebhookRequest,
    params *common.VerificationParams,
) (bool, error)
```

Return `true` to allow webhook processing, `false` to reject as untrusted, and an `error` only for
*unexpected* failures.

`WebhookVerifierConnector` embeds `Connector` and `BatchRecordReaderConnector`, so implementing it also
requires `GetRecordsByIds` (the compiler enforces this). That's by design: a webhook event carries only
the affected record's **id and event metadata**, not the record body — after the event is parsed, the
server fetches the full record via `GetRecordsByIds`. So your event types only need to expose the object
name, record id, and event type.

> **Required: populate `ReadResultRow.Id` in `GetRecordsByIds`.** The server correlates each fetched
> record back to its webhook event by matching the event's record id against `ReadResultRow.Id`
> (`fetchedByID[row.Id]`). If your parser leaves `Id` empty and only fills `Raw`/`Fields`, the lookup
> misses, the record is **silently dropped**, and the event is delivered with no record body attached —
> even though the provider fetch returned `200 OK` with the record. Set `row.Id` to the record id
> (the same value the event's `RecordId()` returns) for every row you return.

Plus a provider-specific `VerificationParams` struct (the caller fills it in per installation) and one
or more event types.

Files: `providers/<provider>/subscribeEvent.go` (or similarly named).

## Event types

When a webhook arrives, the caller casts the raw payload into your typed events and asks them
provider-agnostic questions (what object? what record id? create or update?). Implement these
interfaces (from `common/types.go`) on a provider event type:

```go
type Event interface {
    RawMap() (map[string]any, error)
}

type SubscriptionEvent interface {
    Event
    EventType() (SubscriptionEventType, error)
    RawEventName() (string, error)
    ObjectName() (string, error)
    Workspace() (string, error)
    RecordId() (string, error)
    EventTimeStampNano() (int64, error)
    PreLoadData(data *SubscriptionEventPreLoadData) error  // called first; receives request headers/body
}

type SubscriptionUpdateEvent interface {  // implement if the provider reports which fields changed
    SubscriptionEvent
    UpdatedFields() ([]string, error)
}

type CollapsedSubscriptionEvent interface {  // when one webhook payload fans out to N events
    Event
    SubscriptionEventList() ([]SubscriptionEvent, error)
}
```

Salesloft models all three with a `map[string]any`:

```go
type SubscriptionEvent map[string]any
type CollapsedSubscriptionEvent map[string]any

var (
    _ common.SubscriptionEvent          = SubscriptionEvent{}
    _ common.SubscriptionUpdateEvent    = SubscriptionEvent{}
    _ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}
)
```

- **`CollapsedSubscriptionEvent`** is the raw payload as it arrives. `SubscriptionEventList()` splits it
  into individual `SubscriptionEvent`s. Salesloft sends one record per webhook, so it returns a
  single-element slice; Salesforce/Zoho batch many events into one payload and fan out here.
- **`PreLoadData`** runs first for every event and is handed the request headers/body — use it to stash
  anything (e.g. an event-name header) the other methods need.

See [`providers/salesloft/subscribeEvent.go`](../../providers/salesloft/subscribeEvent.go) for a
complete, readable implementation.

> **The concrete `CollapsedSubscriptionEvent` type must be exported** and defined as a type convertible
> from `map[string]any`. The server converts the raw webhook payload to it by name —
> `<provider>.CollapsedSubscriptionEvent(rawEvent)` in `shared/subscribe/providers/events.go` — then
> calls `SubscriptionEventList()`. This also means enabling a new provider requires adding a `case` for
> it to that server-side switch (ENG-3785 tracks refactoring this). The per-event `SubscriptionEvent` /
> `SubscriptionUpdateEvent` types only need to satisfy their `common` interfaces; they don't need to be
> exported for this path.

## Objects whose records exist only in the payload

The fetch contract above assumes the server can retrieve the record by id after the event arrives.
Some objects break that assumption: the provider exposes no singular lookup for them, while every
webhook event ships the full record. Slack messages are the motivating case — a message has no
standalone id, and is retrievable only through windowed `conversations.history` / `conversations.replies`
calls, so there is nothing for `GetRecordsByIds` to fetch.

If any object of your provider falls into this, do two things:

**1. Return `common.ErrRecordsOnlyInline` from `GetRecordsByIds` for that object**, and only that
object. The check is per object name, not per connector — the same connector keeps fetching normally
for everything else.

```go
func (c *Connector) GetRecordsByIds(
    ctx context.Context, objectName string, ids, fields []string, associations []string,
) ([]common.ReadResultRow, error) {
    if webhook.ObjectRecordsInline(objectName) {
        return nil, fmt.Errorf("%w: %s", common.ErrRecordsOnlyInline, objectName)
    }
    ...
}
```

**2. Implement `SubscriptionEventWithRecord` on the event type** so the record can be read out of the
payload. `Record(fields)` returns a `ReadResultRow` marshaled exactly as a read would: `Raw` holds the
full provider record, `Fields` holds the requested subset, and `Id` holds the same value `RecordId()`
returns.

Do not reach for `ErrGetRecordNotSupportedForObject` here. That error says only that the fetch is
unavailable, so the server delivers the event with no record body. `ErrRecordsOnlyInline` additionally
says where the record does live, which tells the server to proceed with no fetched records and read
them from the events instead of dropping the batch or retrying.

Two things follow from this that are easy to miss:

- **The record id may need composing.** An object with no standalone id needs one built from the
  fields that do identify it, and the same value must come back from both `RecordId()` and the row's
  `Id`. Slack messages use `<channel>:<ts>`.
- **Field inheritance yields nothing.** Subscribe inherits fields and mappings from the read object of
  the same name, and an object that cannot be read has none, so the manifest must not declare one.
  Expect `fields` and `mappedFields` to be empty and the record to be delivered in `raw`.

## Verification

`VerifyWebhookMessage(ctx, request *common.WebhookRequest, params *common.VerificationParams)` receives
two objects, both populated by the caller:

> **`VerifyWebhookMessage` cannot rely on the connector's authenticated client.** The verifier connector is instantiated as a bare zero value — the `verifierConnector: &<provider>.Connector{}` you'll declare in your ProviderConfig ([PR 6 — VerificationConfig](./pr-6-provider-config.md#verificationconfig-subscribeverificationgo)) — so it has **no `AuthorizationClient` and no authenticated HTTP client**. Unlike other connector methods, it must verify using only the `request` (headers/body) and `params` (the signing secret threaded in via `VerificationParams.Param`); it must not make authenticated API calls.

**`request *common.WebhookRequest`** — the raw incoming webhook HTTP request, exactly as the provider
sent it:

```go
type WebhookRequest struct {
    Headers http.Header  // request headers — the signature header lives here
    Body    []byte       // raw request body — verify over these exact bytes, don't re-marshal
    URL     string       // request URL (some providers sign method + url + body)
    Method  string       // HTTP method
}
```

**`params *common.VerificationParams`** — a thin wrapper whose single `Param any` field carries your
provider-specific verification struct:

```go
type VerificationParams struct{ Param any }   // e.g. Param = &SalesloftVerificationParams{Secret: ...}
```

Concretely, `Param` is filled per installation by the `paramsFn` you'll declare in your ProviderConfig ([PR 6 — VerificationConfig](./pr-6-provider-config.md#verificationconfig-subscribeverificationgo)); cast it back to your type with `common.AssertType`. The values inside (a signing secret, a token, an account ref, …) come from whatever your connector persisted — **anything your `Subscribe` returns in the `SubscriptionResult` is available to the `paramsFn` to thread back here** (via the `deps.Subscriptions` resolver — Attio recovers its signing secret this way). So if the provider issues a signing secret at subscribe time, return it in the `SubscriptionResult` (PR 4), and PR 6's `paramsFn` supplies it back through `VerificationParams.Param` when the caller invokes `VerifyWebhookMessage`.

Implement `VerifyWebhookMessage` on your `*Connector`. The pattern (from
[`providers/salesloft/subscribeEvent.go`](../../providers/salesloft/subscribeEvent.go)):

1. Pull your provider-specific params out of `params.Param` with `common.AssertType`.
2. Read the signature header from `req.Headers`.
3. Recompute the expected signature over `req.Body` (and sometimes method/url/timestamp) with the
   shared secret, and compare with `hmac.Equal`.

```go
type SalesloftVerificationParams struct {
    Secret string `json:"secret,omitempty"`
}

var _ connectors.WebhookVerifierConnector = &Connector{}

func (c *Connector) VerifyWebhookMessage(ctx context.Context,
    req *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
    if req == nil || params == nil {
        return false, fmt.Errorf("%w: request and params cannot be nil", errMissingParams)
    }

    vp, err := common.AssertType[*SalesloftVerificationParams](params.Param)
    if err != nil {
        return false, fmt.Errorf("%w: %w", errMissingParams, err)
    }

    signature := req.Headers.Get("x-salesloft-signature")
    if signature == "" {
        return false, fmt.Errorf("%w: missing signature header", ErrMissingSignature)
    }

    sigBytes, err := hex.DecodeString(signature)
    if err != nil {
        return false, fmt.Errorf("%w: invalid signature format", ErrInvalidSignature)
    }

    if !hmac.Equal(sigBytes, computeSignature(vp.Secret, req.Body)) {
        return false, fmt.Errorf("%w: signature mismatch", ErrInvalidSignature)
    }
    return true, nil
}
```

The provider-specific `*VerificationParams` struct (here `SalesloftVerificationParams`) is populated per installation by your ProviderConfig's `paramsFn` ([PR 6 — VerificationConfig](./pr-6-provider-config.md#verificationconfig-subscribeverificationgo)) and passed through `common.VerificationParams.Param`. Hubspot and Outreach show HMAC-SHA256 variants; Salesloft uses HMAC-SHA1. For UI Subscription only providers whose events carry no provider signature, verification is bypassed via the `bypassed` flag in the ProviderConfig (PR 6) rather than implemented here (still implement it if the provider does sign its webhooks).

See [Live Tests (Stage 1)](./live-tests.md#stage-1-webhook-verification-runwebhookconsumer) for how to verify this end-to-end.

## Checklist

- [ ] `var _ connectors.WebhookVerifierConnector = &Connector{}` assertion present.
- [ ] `VerifyWebhookMessage` uses constant-time compare (`hmac.Equal`); returns `false` (not an error)
      for untrusted requests.
- [ ] All `SubscriptionEvent` methods implemented; `SubscriptionUpdateEvent` /
      `CollapsedSubscriptionEvent` added where the provider needs them.
- [ ] Unit tests cover valid / invalid / missing-signature, plus each event method against a captured
      real payload. Use the shared table-driven suites in `test/utils/testconn/`:
      `testconn.TestCaseVerifyWebhookMessage` (webhook verification) and
      `testconn.TestCaseSubscriptionEvent` (event mapping).
- [ ] Any object the provider cannot fetch by id returns `common.ErrRecordsOnlyInline` from
      `GetRecordsByIds` and implements `SubscriptionEventWithRecord`, so its record is read from the
      payload. See [Objects whose records exist only in the
      payload](#objects-whose-records-exist-only-in-the-payload).

## Reviewer focus

- Signature algorithm matches the provider's docs (header name, hash, encoding).
- Untrusted requests return `(false, nil)` rather than erroring.
- Event mapping (object name, record id, event type) is correct for real payloads.
- Objects with no by-id fetch return `ErrRecordsOnlyInline` rather than
  `ErrGetRecordNotSupportedForObject`, and the check is scoped to those objects.

## Reference

- [Core types](../../SUBSCRIBE_REFERENCES.md#core-types)
- [`providers/salesloft/subscribeEvent.go`](../../providers/salesloft/subscribeEvent.go)
