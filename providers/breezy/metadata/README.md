# Metadata

The static file `schemas.json` was created by manually inspecting the [Breezy HR API documentation](https://developer.breezy.hr/reference/overview) and live API responses, then translating the supported objects into our static schema format. All listed objects can be downloaded via read.

Webhook endpoints (`/company/{company_id}/webhook_endpoints`) are intentionally excluded from read and metadata; a future Subscribe implementation will use those APIs directly.

## Read filtering

- **Positions list:** No `state` query param is sent; Breezy defaults to `published` positions per [company-positions](https://developer.breezy.hr/reference/company-positions).
- **Connector-side incremental:** `positions` (`updated_date`) filters by `ReadParams.Since` / `Until` after fetching the full list. Breezy does not expose time-range query params on this endpoint.
- **Full sync only:** `companies`, `pipelines`, `categories`, `departments`, `questionnaires`, and `templates` are small lookup lists with no incremental timestamps.

## Write / delete

Only **`positions`** supports Write and Delete. All other objects are read-only.

### Write (`positions`)

- **Create:** `POST /company/{company_id}/positions` — new positions are created in **`draft`** state.
- **Update:** `PUT /company/{company_id}/position/{position_id}`.
- The collection path uses plural `positions`; the resource path uses singular `position`.
- Write payloads are passed through to Breezy as JSON; there is no separate custom-field metadata layer for this connector.

### Delete (`positions`) — soft delete (archive)

The connector exposes a **Delete** operation for `positions`, but Breezy does **not** provide an HTTP `DELETE` method for positions.

Instead, Delete is implemented as a **soft delete**: `PUT /company/{company_id}/position/{position_id}/state` with `{"state":"archived"}`. The position record remains in Breezy; it is removed from the default **published** positions list (same as archiving in the Breezy UI).

**Implications for customers:**

- Calling Delete does not permanently remove the job posting.
- Archived positions will not appear in the default Read list (`GET …/positions` without a `state` filter defaults to `published`).
- To read archived positions, Breezy supports `?state=archived` on the list endpoint; the connector does not pass a `state` query param today.

### Custom fields

Breezy positions may include `custom_attributes` in API responses. This connector does not collect or schema-map custom fields separately; standard position fields are defined in `schemas.json`. Write accepts any fields Breezy accepts on create/update per [their API docs](https://developer.breezy.hr/reference/company-positions-add).
