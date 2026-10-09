# Jump connector

Jump exposes a single GraphQL endpoint: `https://my.jumpapp.com/enterprise/graphql`.
API reference (login required): https://my.jumpapp.com/enterprise/documentation

## Supported Objects
Below is an exhaustive list of objects & methods supported on the objects

| Object            | Resource          | Method |
|-------------------|-------------------|--------|
| contacts          | contacts          | read   |
| documents         | documents         | read   |
| integrations      | integrations      | read   |
| meetingPreps      | meetingPreps      | read   |
| meetings          | meetings          | read   |
| notes             | notes             | read   |
| pulses            | pulses            | read   |
| scorecards        | scorecards        | read   |
| signalDefinitions | signalDefinitions | read   |
| tasks             | tasks             | read   |
| users             | users             | read   |

## Incremental Read
All objects support incremental read via the `updatedAfter` / `updatedBefore` query arguments.
Jump compares timestamps to the microsecond and both bounds are inclusive, so `Since` and `Until`
are sent with their fractional seconds.

## Metadata
Jump does not publish an OpenAPI specification and disables GraphQL introspection for API keys.
Object metadata is generated from the published GraphQL schema, see `metadata/README.md`.
Nested relations (for example `meetings.transcript`) are only queried when requested.

## Query Complexity
Jump rejects queries above 1000 complexity points. A list query costs `first × (1 + selected fields)`,
and some fields have fixed costs (for example `Meeting.rawTranscript` is 20). The page size is
therefore computed per object from the requested fields, see `objects.go` and `read.go`.
The per-record costs were measured against the live API and must be updated if query templates change.

## Known Limitations
- **Contacts reject every pagination cursor**: the `contacts` query returns `INVALID_PAGINATION` for any
  `after` value, including the `endCursor` it returned itself. Contacts are ordered by `insertedAt` and
  then `id`, so subsequent pages are requested with the inclusive `insertedAfter` argument and records
  already returned are dropped. If more contacts than fit on a page share the same `insertedAt`,
  pagination stops with an error instead of looping.
- **Rate limit is per IP**: 1000 requests per minute per IP address, shared by every token used from it.
  In practice a burst of a few hundred requests per minute, followed by more requests while throttled,
  locked the IP out for weeks (`Retry-After` of about 1.96 million seconds). Stop sending requests on
  the first `429`.
- **`users` requires an account-scoped API key**: OAuth tokens cannot query users.
