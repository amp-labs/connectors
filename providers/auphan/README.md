# Auphan connector

Reads daily sales and payment totals from Auphan POS. **It is read-only on purpose.** The Auphan
API has write endpoints, but tokens are often live production tokens (`"write_mode": "live"`), so
Write is not implemented and `Support.Write` is false.

API reference: https://api1.auphansoftware.com/resources/

## Auth and base URLs

- API key in the `X-Auphan-Api-Token` header.
| Provider        | Constructor           | Base URL                                               | Store ID comes from  |
|-----------------|-----------------------|--------------------------------------------------------|----------------------|
| `auphan`        | `NewConnector`        | `https://{workspace}.gw1.auphansoftware.com/api/v1.php` | workspace            |
| `auphanSandbox` | `NewSandboxConnector` | `https://api1.auphansoftware.com/{workspace}/api/v1.php` | `storeId` metadata  |

- In production, the workspace is the store ID.
- In the sandbox, the workspace is the demo name. Auphan provisions it per partner with its own
  seeded store ID and token, so the store ID is a separate input.
- Every call is a `POST {base}/{domain}/{action}` with the body `{"store_id": <int>, "params": {...}}`.
- One connection reads one store. A franchise HQ token can read many stores by changing
  `store_id`, which this connector does not do yet.

## Supported objects

| Object     | Endpoint                | Read | Pagination                                      |
|------------|-------------------------|------|-------------------------------------------------|
| `payments` | `reports/execute`       | yes  | none, one page                                  |
| `sales`    | `reports/execute`       | yes  | none, one page                                  |
| `stores`   | `settings/stores.list`  | yes  | `data.pagination.page` / `total_pages`, 200 per page |

To add another report type, add it to `reportObjects` in `support.go`.

### Report reads

- `ObjectName` is the report `type`. `Fields` become `columns`, sorted so the request is stable.
  Auphan requires at least one column.
- `filters.selection` is always `close_date`.
- `Since` is required, and the read fails without it. A report needs a date range, and the
  connector does not pull all history silently. `Until` defaults to now.
- Dates go out as `YYYY-MM-DD HH:MM:SS`.
- Values come back as strings (`"1411.96"`, `"36"`, `"Nov 05, 2026"`). They are returned as received
  and never converted to float, because this is accounting data.
- Reports have no paging. A top-level `result_limit` or `page` is ignored, and `filters.result_limit`
  is rejected with `UNKNOWN_FILTER` (checked live). Every read returns one page with `Done: true`.

### Paired rows in the summary layout

In `layout: "summary"` each group arrives as two rows. The dimension row has every measure null,
and the measure row that follows has every dimension null:

```json
["Nov 05, 2026", "MC", "", null, null, null],
[null, null, null, "1411.96", "180.76", "36"]
```

Each pair is merged into one record keyed by column `id`. Rows are checked one at a time, so a
layout that returns whole rows is kept as is.

### Metadata

- Report objects: built from `reports/types.get`. Each column `id` becomes the field name, `label`
  the display name, and `kind` (dimension or measure) the provider type.
- `stores`: sampled from the first `stores.list` record.

## Errors

Auphan may return failures with HTTP 200 and `{"success": false, "error": {"code", "message"}}`.
Every response is checked for this.

| Code | Error |
|---|---|
| `MISSING_TOKEN`, `INVALID_TOKEN` | `common.ErrAccessToken` |
| `USER_DISABLED`, `STORE_ACCESS_DENIED`, `MODULE_ACCESS_DENIED`, `ACCESS_DENIED`, `GROUP_NOT_ASSIGNABLE` | `common.ErrForbidden` |
| `INVALID_PARAM`, `VALIDATION_ERROR`, `INVALID_REQUEST`, `UNKNOWN_FILTER` | `common.ErrBadRequest` |
| `NOT_FOUND` | `common.ErrBadRequest` + `common.ErrNotFound` |
| any other code | `common.ErrRequestFailed` |

Non-2xx statuses go through the standard status mapping, so 429 becomes `common.ErrLimitExceeded`.

The docs list `401 INVALID_TOKEN`, but live, the gateway rejects an unknown token on its own with a
plain-text `Forbidden` and HTTP 403. A non-JSON 403 is therefore mapped to `common.ErrAccessToken`,
so a revoked token shows up as bad credentials. JSON 403s from the store, like
`MODULE_ACCESS_DENIED`, stay `common.ErrForbidden`.

## Live tests

`test/auphan/read` reads `payments` for 2026-11-05 and checks the result against known totals.
`test/auphan/metadata` lists metadata for every object. Both read `./auphan-creds.json`:

```json
{"provider": "auphan", "apiKey": "<token>", "substitutions": {"workspace": "<store id>"}}
```

They only call read endpoints.

## Open questions

- **Timezone.** Auphan returns `null` for the store timezone. `Since`/`Until` are formatted in the
  location of the `time.Time` they come in, with no conversion to UTC. The caller has to pass
  times in the store's local time. Someone needs to confirm with Auphan which zone `close_date`
  uses.
- **Presets.** Presets are not used. System presets have negative IDs, and `presets.get` rejects
  them.
