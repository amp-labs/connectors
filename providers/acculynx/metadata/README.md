# AccuLynx Metadata

`schemas.json` is auto-generated from the AccuLynx V2 OpenAPI specification by
`scripts/openapi/acculynx/metadata`. Do not edit it by hand.

The OpenAPI spec lives at `openapi/openapi.json` in this directory. AccuLynx
hosts the spec on ReadMe (the docs platform they use) but does not link it from
their developer site — the URL is embedded as the `oasPublicUrl` value in the
HTML of any page under [apidocs.acculynx.com/reference/](https://apidocs.acculynx.com/reference/).

To refresh after AccuLynx ships a new API version:

1. Open any reference page, e.g. https://apidocs.acculynx.com/reference/getcalendars
2. View page source and search for `api-registry`. The matching URL looks like
   `https://dash.readme.com/api/v1/api-registry/<id>`.
3. Fetch that URL and save the response as `openapi/openapi.json`.
4. From the repo root, run:

   ```sh
   go run ./scripts/openapi/acculynx/metadata
   ```

The script projects the spec down to the connector's supported object
inventory (defined in `objectEndpoints` in the script). To add or remove an
object, edit that map and re-run the generator.

## Why some endpoints aren't here

The spec contains 97 V2 paths but we only expose ~33 as connector objects.
Excluded categories:

- **By-id-only endpoints** (`/invoices/{invoiceId}`, `/financials/{financialsId}`,
  `/jobs/{jobId}` etc.) — no list semantics.
- **Singletons-per-job** (`/jobs/{jobId}/adjuster`, `/insurance`, `/financials`,
  `/initial-appointment`, `/accounting/integration-status`,
  `/representatives/ar-owner`/`sales-owner`/`company`,
  `/milestones/current`) — return one record, not a collection.
- **Write-only endpoints** (`/jobs/{jobId}/documents`, `/messages`,
  `/photos-videos`, `/payments/*`) — POST-only in the spec.
- **Aggregate dashboards** (`/jobs/{jobId}/payments`, `/payments/overview`) —
  shape is a stats blob, not a list.
- **Search POSTs** (`/jobs/search`, `/contacts/search`) — `/jobs` GET already
  supports time-based filtering via `dateFilterType=ModifiedDate`.
- **Reports** (`/reports/scheduled-reports/*`) — require pre-known IDs from
  outside the API.
- **Pre-known-id histories** (`/leads/{leadId}/history`) — same.
- **Two-level fan-out** (`/estimates/{estimateId}/sections/{sectionId}/items`,
  `/financials/{financialsId}/amendments`, `/worksheet/items`) — deferred to a
  follow-up; section-level data is still exposed via `estimates/sections`.
- **External-reference bulk lookups** (`/jobs/external-references`) — not a
  data object.
- **Diagnostics** (`/diagnostics/ping`) — operational health check.
