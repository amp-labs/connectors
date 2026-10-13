# Lightfield connector

Read-only deep connector for the Lightfield CRM REST API (public beta).

## Auth and requests

- API key created by a workspace admin, sent as `Authorization: Bearer sk_lf_...`
  ([API keys](https://docs.lightfield.app/using-the-api/api-keys)). The key needs the
  `accounts:read`, `contacts:read`, `opportunities:read` and `members:read` scopes.
- Base URL `https://api.lightfield.app`; every path is under `/v1`.
- Every request also carries `Lightfield-Version: 2026-03-01`, which the spec marks as required.
  The connector sets it on reads and definitions calls. The catalog entry has proxy enabled, but
  it cannot add a fixed header to ApiKey auth, so proxy callers must send `Lightfield-Version`
  themselves (reported as a common-header limitation).

## Objects

| Object          | Endpoint                | Metadata source                          |
|-----------------|-------------------------|------------------------------------------|
| `accounts`      | `GET /v1/accounts`      | catalog + `GET /v1/accounts/definitions` |
| `contacts`      | `GET /v1/contacts`      | catalog + `GET /v1/contacts/definitions` |
| `opportunities` | `GET /v1/opportunities` | catalog + `GET /v1/opportunities/definitions` |
| `members`       | `GET /v1/members`       | catalog only (no definitions endpoint)   |

`members` are the workspace users the customer asked for as "users".
The catalog (`internal/metadata/objectsMetadata.json`) is generated from the OpenAPI spec by
`go run ./scripts/openapi/lightfield/metadata`. It holds the record properties of every object
and the fixed member fields. The definitions endpoints add each workspace's system and custom
fields, select options (option ID as value, label as display value), `readOnly` and `system`.

## Record shape

Lightfield wraps values: `fields: {"$name": {"value": "Acme", "valueType": "TEXT"}}` and
`relationships: {"$owner": {"values": ["usr_1"], ...}}`. Read flattens both into top-level fields
keyed verbatim (`$name`, `$owner`, custom slugs such as `icp-score`). A field becomes its `value`;
a relationship becomes its `values` array of IDs, whatever its cardinality. `Raw` is untouched.
Custom keys have no `$`, so on a name clash the record property wins, then fields, then relationships.
Metadata types HAS_ONE relationships as `reference` (with `ReferenceTo` when the target is an
object this connector reads; `user` is not) and HAS_MANY ones as `other`, being lists of IDs.

## Pagination

`limit` (max 25) and `offset`. A page with fewer records than `limit` is the last one
([List methods](https://docs.lightfield.app/using-the-api/list-endpoints#pagination)).
`NextPage` is the request URL with the offset advanced.

## Incremental read

Accounts, contacts and opportunities filter on the server: `Since` is sent as
`$updatedAt[greaterThan]` and `Until` as `$updatedAt[lessThanOrEqual]` (RFC 3339 UTC with
milliseconds), so Since is exclusive and Until inclusive as in ReadParams; both carry over to every
page ([List methods](https://docs.lightfield.app/using-the-api/list-endpoints#filtering)).
The `$` is part of the key; `updatedAt[greaterThan]` returns 400 `unknown_field`.
`/v1/members` rejects filters (400 `unrecognized_keys`), so members are read in full and the
connector keeps those whose `updatedAt` (or `createdAt` when null) is within `(Since, Until]`.
List results come from a search index that may lag recent changes, so a read can miss them.

## Rate limits

25 requests per second per organization for each of the Search (list) and Read (definitions)
categories; 429 responses carry `Retry-After`
([Rate limits](https://docs.lightfield.app/using-the-api/rate-limits)).
