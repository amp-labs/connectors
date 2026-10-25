---
name: connector-conventions
description: "How to build a connector in this repo, in two parts: its structure, which matches how people write connectors here today, and its correctness patterns (metadata, errors, paging, per-record requests, writes and deletes, the supports registry, Proxy and the catalog entry, tests), each with what it is for and when it's wrong. Load before writing, changing or reviewing anything under providers/."
paths:
  - "providers/**"
metadata:
  tags: [connectors, build, review]
---

# Connector conventions

## Why

A connector is read by the people who maintain this repo and run by Ampersand's server for every
consumer who installs it. Two things decide whether it's good, and they are judged differently.

- **Structure** (files, naming, the constructor, where things live, test layout, comments) has no
  best answer. What matters is consistency: a connector shaped like the ones around it can be read
  and maintained at a glance, and one shaped differently costs every reviewer time.
- **Correctness** (what a read returns, what an error makes the server do) is judged by what the
  code does for this provider. A pattern being common in the repo is not a reason to use it; its
  purpose holding for this provider is.

## What you can use

- The repo itself. Read as many connectors as are useful, most of all recent ones written by
  people; `git log` shows who added each connector and when. Connectors whose author is a bot
  account are generated, so they show a generator's habits rather than the team's. `conventions.md` explains the pieces, and `scripts/count_conventions.sh` counts how
  often each pattern appears, which is context, never a reason.
- The provider's docs, for every behaviour the correctness patterns below depend on.
- `go test ./providers/<name>/...` for the unit tests, and the programs under `test/<name>/`
  (see `test/README.md`), which run the connector against a real account and show what the
  provider actually returns.
- Two facts reading won't surface: `components.Initialize` is deprecated
  (`internal/components/connector.go`) and new code uses `components.Init`; and comments here say
  why and link the provider's docs page rather than narrating the code.

## Working it out

Study how recent human-written connectors look today and give this one the same shape for its
situation, so it reads like one a person contributed this month. For correctness, understand what
each pattern below is for, then decide what this provider needs.

**Metadata.** `ListObjectMetadata` comes from a schema provider set in the constructor, and the
right one depends on where the truth about fields lives: a static file generated from the
provider's spec when fields are fixed (`schema.NewOpenAPISchemaProvider`, or the newer
`internal/objectcatalog` format; use the one current human connectors use), a describe endpoint at
run time when the account defines fields (`schema.NewObjectSchemaProvider`), or both
(`schema.NewCompositeSchemaProvider`). Inferring fields from the first record of a list is wrong
whenever the account is empty, a field is missing or null on that record, or types matter; it is a
last resort when there is no spec and no describe endpoint, and the PR says so.

**Errors.** The error a failed response becomes decides what the server does: `ErrBadRequest`,
`ErrCaller` and `ErrForbidden` stop the read, `ErrLimitExceeded`, `ErrServer` and `ErrRetryable`
are retried, and `ErrAccessToken` refreshes the token. `common.InterpretError` keeps status, body
and headers; `interpreter.ErrorHandler` with a `FormatSwitch` reads the provider's error envelope but
drops status and headers unless wrapped with `common.NewHTTPError`. The shared status map sends 408
to `ErrBadRequest` and the two handlers disagree on 404, so a mapping is wrong when it stops a read
the provider means as transient, or retries one that can never succeed. Map what the provider
documents, including limits or transient failures it reports some other way.

**Paging.** The next-page token is whatever lets the next call resume (a cursor, a next link,
`httpkit.HeaderLink` for a `Link` header, a page number or offset). A read is `Done` only when the
provider says there is nothing more; a short page is an end only if the docs say so, because many
providers treat page size as a hint. A read in several passes keeps which pass it's on in the token
(`readhelper.AggregateReadResults`, `GetAggregateToken`). Page size comes from
`readhelper.PageSizeWithDefault`, defaulting to the provider's maximum. Positional paging also
constrains Until (`time-filters`).

**Per-object logic.** Most differences between objects are data (a path, a records key, a time
field) kept in tables keyed by object name. Logic only one object needs gets its own function, and
its own file when that's how the repo's connectors split it.

**Fields that need a second request.** Each record then costs another request, and one record's
failure can fail the page. Prefer a batch or expand parameter; otherwise fetch only when the caller
asked for the field (`params.Fields`) or the related object (`readhelper.HydrateAssociations`), only
for records that passed the time filter, with bounded concurrency (`internal/parallelfetch`). Only
the documented "not there" status becomes an empty field; anything else, a timeout included, fails
the page, because an empty field delivered as truth is worse than a retried page.

**Writes and deletes.** A write validates the params and object and uses the method and path the
docs give, with the record id in the path for an update. Its response yields the provider's id; a
create with no id in its response is a question for the docs, and a 2xx carrying an error in its
body is an error. A delete succeeds on the documented status; whether a 404 means "already gone" is
the provider's semantics to cite.

**The supports registry.** The endpoint registry rejects an unsupported object before any request.
`components.NewEmptyEndpointRegistry()` allows every object, which is right only when the request
builders reject unknown objects themselves; otherwise the provider's 404 comes back as retryable and
the server retries a request that can never succeed.

**The catalog entry and Proxy.** Proxy calls go to the catalog's `BaseURL` with the builder's path
appended (`internal/components/proxy.go`), so the BaseURL is the API's root: a version or a tenant
value after it pins every proxy call. A value the consumer would otherwise type (a workspace, a
company id) is discovered after auth when an endpoint returns it, since every input is a step a
consumer can get wrong; when none does, the PR names the endpoints checked. `DocsURL` is the
provider's page for the auth scheme used, and media has icon and logo in regular and dark mode.

**Tests and fixtures.** Unit tests show what a read returns, not only the request it sends: a page
with more after it (and the token the next request sends), the last page, an empty page, an unknown
object, and Since filtering where it applies, written with the `mockserver` and `mockcond` helpers.
Fixtures are real response shapes with personal data removed in every commit on the branch, because
the history is public.
