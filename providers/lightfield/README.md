# Lightfield connector

Read-only deep connector for the Lightfield CRM REST API (public beta).

## Auth and requests

- API key created by a workspace admin, sent as `Authorization: Bearer sk_lf_...`
  ([API keys](https://docs.lightfield.app/using-the-api/api-keys)). The key needs the
  `accounts:read`, `contacts:read`, `opportunities:read` and `members:read` scopes.
- Base URL `https://api.lightfield.app`; every path is under `/v1`.
- Every request also carries `Lightfield-Version: 2026-03-01`, which the spec marks as required.
  The connector sets it on reads and definitions calls. The catalog cannot add it, so proxy
  callers would have to send it themselves (proxy was not in scope for this build).

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

## Pagination

`limit` (max 25) and `offset`. A page with fewer records than `limit` is the last one
([List methods](https://docs.lightfield.app/using-the-api/list-endpoints#pagination)).
`NextPage` is the request URL with the offset advanced.

## Incremental read

The list endpoints have no updated-since filter: filters apply to entries of `fields`, and
`updatedAt` is a record property. Lists are sorted by `createdAt`, newest first. So every read
fetches all pages and the connector keeps records whose `updatedAt` (or `createdAt` when
`updatedAt` is null) is within `[Since, Until]`. Lists are served from a search index and may
lag recent writes, per the docs.

## Rate limits

25 requests per second per organization for each of the Search (list) and Read (definitions)
categories; 429 responses carry `Retry-After`
([Rate limits](https://docs.lightfield.app/using-the-api/rate-limits)).
