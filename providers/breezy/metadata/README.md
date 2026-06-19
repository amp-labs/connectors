# Metadata

The static file `schemas.json` was created by manually inspecting the [Breezy HR API documentation](https://developer.breezy.hr/reference/overview) and live API responses, then translating the supported objects into our static schema format. All listed objects can be downloaded via read.

## Read filtering

- **Positions list:** No `state` query param is sent; Breezy defaults to `published` positions per [company-positions](https://developer.breezy.hr/reference/company-positions).
- **Connector-side incremental:** `positions` (`updated_date`) and `webhook_endpoints` (`updated_at`) filter by `ReadParams.Since` / `Until` after fetching the full list. Breezy does not expose time-range query params on these endpoints.
- **Full sync only:** `companies`, `pipelines`, `categories`, `departments`, `questionnaires`, and `templates` are small lookup lists with no incremental timestamps.
