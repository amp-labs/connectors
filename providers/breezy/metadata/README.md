# Metadata

The static file `schemas.json` was created by manually inspecting the [Breezy HR API documentation](https://developer.breezy.hr/reference/overview) and live API responses, then translating the supported objects into our static schema format. All listed objects can be downloaded via read.

Webhook endpoints (`/company/{company_id}/webhook_endpoints`) are intentionally excluded from read and metadata; a future Subscribe implementation will use those APIs directly.

## Read filtering

- **Positions list:** No `state` query param is sent; Breezy defaults to `published` positions per [company-positions](https://developer.breezy.hr/reference/company-positions).
- **Connector-side incremental:** `positions` (`updated_date`) filters by `ReadParams.Since` / `Until` after fetching the full list. Breezy does not expose time-range query params on this endpoint.
- **Full sync only:** `companies`, `pipelines`, `categories`, `departments`, `questionnaires`, and `templates` are small lookup lists with no incremental timestamps.
