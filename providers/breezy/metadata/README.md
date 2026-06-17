# Metadata

The static file `schemas.json` was created by manually inspecting the [Breezy HR API documentation](https://developer.breezy.hr/reference/overview) and live API responses, then translating the supported objects into our static schema format. All listed objects can be downloaded via read.

## Read filtering

- **Provider-side:** `positions` accepts `ReadParams.Filter` as the Breezy `state` query value (`published`, `draft`, `archived`, etc.). The API defaults to `published` when omitted.
- **Connector-side incremental:** `positions` (`updated_date`) and `webhook_endpoints` (`updated_at`) filter by `ReadParams.Since` / `Until` after fetching the full list. Breezy does not expose time-range query params on these endpoints.
- **Full sync only:** `companies`, `pipelines`, `categories`, `departments`, `questionnaires`, and `templates` are small lookup lists with no incremental timestamps.
