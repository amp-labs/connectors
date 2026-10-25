---
name: connector-conventions
description: "How to build a connector in this repo, in two parts: its structure, which follows how people write connectors here today, and its correctness patterns (metadata, errors, paging, per-record requests, writes and deletes, the supports registry, Proxy, tests), each with what it does, why, and when it's wrong. Load before writing, changing or reviewing anything under providers/."
paths:
  - "providers/**"
metadata:
  tags: [connectors, build, review]
---

# Connector conventions

Conventions here are of two kinds, and they're judged differently.

- **Structure** (files, naming, constructor, where things live, test layout, comments) follows how
  people write connectors in this repo today. No layout is best; a consistent one lets a reviewer
  read a new connector at a glance.
- **Correctness** (what a read returns, what an error makes the server do) is judged by what the
  code does. A pattern many connectors share can still be wrong: several send Until to the
  provider under offset paging and can skip records, and this repo's own docs once advised pacing.
  So a pattern is followed because its reason holds for this provider, never because it's common.
  `scripts/count_conventions.sh` counts patterns, as context only.

`conventions.md` explains the pieces; `time-filters` and `server-owns-limits` cover Since/Until and
rate limits.

## Structure: read like a connector a person wrote

Before writing, read as many connectors as you can, most of all recent ones written by people:

```
git log --diff-filter=A --format='%cs %an' --name-only -- 'providers/*/connector.go'
```

Leave out connectors added by `connie-from-ampersand[bot]`; they are generated, and copying them
copies their quirks. Notice which files a connector has and what goes in each, how objects and
functions are named, how the constructor is built, where handlers, supported objects, metadata and
error parsing live, how tests and fixtures are laid out, and how much is commented. Then apply that
shape to this connector's situation, so it reads like one a person contributed this month.

Two things that reading won't show:
- `components.Initialize` is marked deprecated in `internal/components/connector.go`; new code uses
  `components.Init`, which also hands the constructor the params (metadata, module).
- Comments say why and link the provider's docs page; they don't narrate the code.

## Metadata

`ListObjectMetadata` comes from a schema provider set in the constructor
(`components.SchemaProvider`). Which one depends on where the truth is:

- **A static file** generated from the provider's OpenAPI spec, when fields are fixed. The
  long-standing format is `metadata/schemas.json` through `schema.NewOpenAPISchemaProvider`; the
  maintainers' newer one is `objectsMetadata.json(.gz)` through `internal/objectcatalog` (see its
  README). Both are correct; use the one current human connectors use.
- **A request at run time**, through `schema.NewObjectSchemaProvider`, when the account defines
  fields (custom fields, per-tenant schemas): a describe endpoint is the source.
- **Both**, through `schema.NewCompositeSchemaProvider`, when some objects are static and some
  aren't.

Inferring fields from the first record of a list is common and risky: an empty account fails
metadata, a field missing or null on that record is never offered, and types are guessed. Use it
only when the provider has no describe endpoint and no spec, and say so in the PR.

## Errors

The error handler turns a failed response into a `common` error, and that error decides what the
server does: `ErrBadRequest`, `ErrCaller` and `ErrForbidden` stop the read; `ErrLimitExceeded`,
`ErrServer` and `ErrRetryable` are retried; `ErrAccessToken` refreshes the token.

- `common.InterpretError` keeps the status, body and headers (`common.HTTPError`). It's enough
  when the provider's error body needs no parsing.
- `interpreter.ErrorHandler` with a `FormatSwitch` (in `errors.go`) makes the provider's error
  envelope readable. It returns the mapped error without the status and headers; wrap it with
  `common.NewHTTPError` when callers need them.
- The shared status map (`interpreter.DefaultStatusCodeMappingToErr`) sends 408 to
  `ErrBadRequest`, so a provider timeout stops the read, and the two handlers disagree on 404
  (`InterpretError` retries it). Override a status in the connector's map when the provider
  documents a different meaning, and map a limit or transient failure the provider reports some
  other way (a 400 with a code) to the error that retries.

## Paging

The next-page token is whatever lets the next call resume: the provider's cursor or next link
(`httpkit.HeaderLink` for a `Link` header), or a page number or offset. The read is `Done` only
when the provider says there's nothing more: no cursor, no next link, or a documented count. A short
page is not an end unless the docs say so, because many providers treat page size as a hint. A read
in several passes (one status at a time, one parent at a time) keeps which pass it's on in the
token (`readhelper.AggregateReadResults` and `GetAggregateToken`). Page size comes from
`readhelper.PageSizeWithDefault`, with the provider's maximum as the default. Positional paging
also constrains Until (`time-filters`).

## Per-object logic

Most differences between objects are data: a path, a records key, a time field, kept in tables
keyed by object name. Code that only one object needs (a different read path, a multi-pass read)
gets its own function, and its own file only when that's how the repo's connectors split it.

## Fields that need a second request

When a field isn't in the list response (a transcript per meeting), each record costs another
request. That multiplies the page's cost against the provider's limit and makes one record's failure
able to fail the page. So:
- prefer a batch or expand parameter; otherwise fetch only when the caller asked for the field
  (`params.Fields`) or the related object (`AssociatedObjects`, `readhelper.HydrateAssociations`),
  and only for records that passed the time filter;
- run the requests with bounded concurrency (`internal/parallelfetch`), never paced or retried in
  the connector (`server-owns-limits`);
- map only the documented "not there" statuses (often 404) to an empty field and fail the page on
  the rest, a timeout included;
- say in the PR what a page costs in requests. When the cost is high, a separate object may be the
  better shape; that's the team's call.

## Writes and deletes

- `buildWriteRequest` validates the params and the object, then creates or updates with the method
  and path the docs give. The record id goes in the path for an update, not the body.
- `parseWriteResponse` returns the provider's id for the record. A create whose response carries no
  id is a question for the docs, not a success with an empty id; a provider that answers 2xx with an
  error in the body maps it to an error.
- A delete succeeds on the documented status. Whether a 404 means "already gone" is the provider's
  semantics to cite, not a default.

## The supports registry

The endpoint registry passed to the reader, writer and deleter rejects an object before any request
(`ErrOperationNotSupportedForObject`). `components.NewEmptyEndpointRegistry()` allows every object,
which is fine only when `buildReadRequest` (and its write and delete peers) rejects unknown objects
itself. Without either check the provider's 404 comes back, which `InterpretError` marks retryable,
so the server retries a request that can never succeed. A unit test reads an unknown object.

## The catalog entry and Proxy

- Proxy calls go to the catalog's `BaseURL` (module `BaseURL` for a module), with the builder's path
  appended (`internal/components/proxy.go`). So the BaseURL is the API's root: no version, and no
  tenant value that sits after a version (that part of the path is built in code), or every proxy
  call is pinned to it. Proxy is enabled once a connectivity test shows a proxied call succeeding.
- A value the consumer would type (a workspace, a company id) is discovered after auth when an
  endpoint returns it, because every input is a step a consumer can get wrong. When there's none,
  the PR names the endpoints checked.
- `DocsURL` is the provider's own page for the auth scheme used. Media has icon and logo, regular
  and dark mode.

## Tests and fixtures

- Each read is tested on what it returns, not only the request it sends: a first page with more
  after it (asserting the token the next request sends), the last page (`Done`), an empty page, an
  unknown object, and Since filtering where it applies. Responses are written with the `mockserver`
  helpers and requests matched with `mockcond`; semgrep rejects writing to the ResponseWriter.
- Fixtures are real response shapes with personal data removed: no real email, name or signed URL,
  in any commit on the branch, since history is public.
