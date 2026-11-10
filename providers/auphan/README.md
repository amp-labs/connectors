# Auphan connector

Reads Auphan POS reports (sales, payments, tax, labor, and more) and the store list. **It is read-only on purpose.** The Auphan
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

| Object | Endpoint | Pagination |
|---|---|---|
| `stores` | `settings/stores.list` | `data.pagination.page` / `total_pages`, 200 per page |
| 31 report objects (below) | `reports/execute` | none: one call returns the whole date range |

Every report Auphan lists in `reports/types.list` is an object, named by its report type:

| Date handling | Reports |
|---|---|
| `selection` + date range | `invoices`, `modifiers`, `payments`, `sales`, `tax` (`close_date`) · `purchases` (`purchase_date`) · `screen_stats` (`start_time`) · `waitlist` (`reservation_dt`) · `worked_hours` (`sales_date`) |
| Date range, no `selection` | `adjustments`, `attendance`, `bank_drop`, `costs`, `counts`, `customer_tabs`, `customers`, `employees`, `invoice_logs`, `message_actions`, `payment_txns`, `petty_cash`, `price_change`, `production`, `receiving`, `sales_printer_stats`, `shifts`, `stock_usage`, `transfers` |
| Date-only range (`YYYY-MM-DD`) | `salerules` |
| No date range (a snapshot) | `gfs_sku`, `usergroups` |

All 31 were read live for one day and their metadata loads. Left out, because they fail live:

- `loyalty_trans` needs a loyalty program id (`tgl_id`).
- `giftcards` fails with `EXECUTION_FAILED`.
- `inventory`, `Inventory_Categories` and `inventory_modifier` fail with `INTERNAL_ERROR`, likely
  because the inventory module is off for the store.

Each report's settings live in `reportObjects` in `support.go`. To add a report, check `reports/types.get`:

- whether it has a `selection` filter (reports without one reject it with `UNKNOWN_FILTER`);
- which filters are required without a default (`attendance` needs `status`, sent as Auphan's default `[4]`);
- whether `date_from` is a `datetime`, a `date`, or absent.

### Report reads

- `ObjectName` is the report `type`. `Fields` become `columns`, sorted so the request is stable.
  Auphan requires at least one column. Some reports fail on some columns (`shifts` returns
  `INTERNAL_ERROR` for its first columns but not for others).
- **Dates:** `Since` is sent as `date_from` and `Until` as `date_to`. Auphan reads a missing
  `date_from` as the start of today (checked live), so without `Since` the read starts at
  2000-01-01 to get the whole history. A missing `date_to` is left to Auphan, which reads through today:

  | Read with | Sent | Auphan returns |
  |---|---|---|
  | `Since` and `Until` | both | that range |
  | `Since` only | `date_from` | `Since` through the end of today |
  | `Until` only | `date_from` 2000-01-01, `date_to` | 2000-01-01 through `Until` |
  | neither | `date_from` 2000-01-01 | everything through the end of today |

- **No range cap:** every dated report was read live from 2000-01-01 to today in one call, without errors and in
  under 6 s. The largest was `invoice_logs`, at 78k rows and 1 MB. This store's data only starts on 2026-09-24, so years
  of real data have not been tested.
- **Timezone:** with the `timezone` metadata (an IANA name like `America/Los_Angeles`), times are
  converted to the store's zone before sending. Without it, times keep their own location.
- Dates go out as `YYYY-MM-DD HH:MM:SS`.
- Values come back as strings (`"1411.96"`, `"36"`, `"Nov 05, 2026"`). They are returned as received
  and never converted to float, because this is accounting data.
- Reports have no paging of their own. A top-level `result_limit` or `page` is ignored, and
  `filters.result_limit` is rejected with `UNKNOWN_FILTER` (checked live). One call returns the whole
  range, so a read is always done after one call.

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

- **Timezone.** Auphan returns `null` for the store timezone, so the `timezone` input has no
  default. Without it, the caller has to pass times in the store's local time. Someone needs to
  confirm with Auphan which zone `close_date` uses.
- **Presets.** Presets are not used. System presets have negative IDs, and `presets.get` rejects
  them.
