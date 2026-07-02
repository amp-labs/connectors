# Jobber connector

## Supported Objects

Below is an exhaustive list of objects & methods supported on the objects
| Object                    | Resource                     | Method |
| ------------------------- | ---------------------------- | ------ |
| appAlerts                 | appAlerts                    | read   |
| apps                      | apps                         | read   |
| capitalLoans              | capitalLoans                 | read   |
| clientEmails              | clientEmails                 | read   |
| clientPhones              | clientPhones                 | read   |
| clients                   | clients                      | read   |
| expenses                  | expenses                     | read   |
| invoices                  | invoices                     | read   |
| jobs                      | jobs                         | read   |
| paymentsRecords           | paymentsRecords              | read   |
| payoutRecords             | payoutRecords                | read   |
| products                  | products                     | read   |
| properties                | properties                   | read   |
| quotes                    | quotes                       | read   |
| requestSettingsCollection | requestSettingsCollection    | read   |
| requests                  | requests                     | read   |
| scheduledItems            | scheduledItems               | read   |
| similarClients            | similarClients               | read   |
| tasks                     | tasks                        | read   |
| taxRates                  | taxRates                     | read   |
| timeSheetEntries          | timeSheetEntries             | read   |
| users                     | users                        | read   |
| vehicles                  | vehicles                     | read   |
| vists                     | vists                        | read   |
| clients                   | clientCreate, clientEdit     | write  |
| events                    | eventCreate                  | write  |
|                           | expenseCreate                |        |
| expenses                  | expenseEdit                  | write  |
|                           | expenseDelete                |        |
| jobs                      | jobCreate, jobEdit           | write  |
| productsAndServices       | productsAndServicesCreate    | write  |
|                           | productAndServicesEdit       | write  |
| quotes                    | quoteCreate, quoteEdit       | write  |
| requests                  | requestCreate, requestEdit   | write  |
| taxes                     | taxCreate                    | write  |
| taxGroups                 | taxGroupCreate               | write  |
| vehicles                  | vehicleCreate, vehicleDelete | write  |

## Incremental read

Jobber's guides don't document it, but the GraphQL schema exposes `filter`
arguments (introspectable without auth) that support timestamp ranges via
`Iso8601DateTimeRangeInput{before, after}` (both inclusive). `Since`/`Until`
map to them as follows:

| Object                                                                          | Strategy                                                                                                                                                              |
| ------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| clients, expenses, invoices, quotes, requests, timeSheetEntries, payoutRecords  | `filter: { updatedAt: { after, before } }`                                                                                                                             |
| visits, tasks, capitalLoans                                                     | `filter: { createdAt: { after, before } }` — these have no `updatedAt`, so updates to existing records are not captured                                                |
| jobs                                                                            | No `updatedAt` filter exists; the query sorts `UPDATED_AT DESCENDING` and the connector drops records older than `Since` and stops paginating (see `incremental.go`)   |
| everything else                                                                 | `Since`/`Until` are ignored (full read)                                                                                                                                |

## Subscribe (webhooks)

Webhook endpoints are managed per account via the `webhookEndpointCreate` /
`webhookEndpointDelete` GraphQL mutations — one endpoint per topic, all
pointing at the same webhook URL. Jobber has no update or list operation for
webhook endpoints, so `UpdateSubscription` reconciles by deleting stale
endpoints and creating new ones, and `DeleteSubscription` relies on the
endpoint IDs stored at creation time.

Subscribable objects and their topics (`{OBJECT}_{CREATE|UPDATE|DESTROY}`):

| Object           | Topic root         | Events                 | Pass-through topics        |
| ---------------- | ------------------ | ---------------------- | -------------------------- |
| clients          | CLIENT             | create, update, delete |                            |
| properties       | PROPERTY           | create, update, delete |                            |
| requests         | REQUEST            | create, update, delete |                            |
| quotes           | QUOTE              | create, update, delete | QUOTE_SENT, QUOTE_APPROVED |
| jobs             | JOB                | create, update, delete | JOB_CLOSED                 |
| visits           | VISIT              | create, update, delete | VISIT_COMPLETE             |
| invoices         | INVOICE            | create, update, delete |                            |
| expenses         | EXPENSE            | create, update, delete |                            |
| users            | USER               | create, update         |                            |
| timeSheetEntries | TIMESHEET          | create, update, delete |                            |
| payoutRecords    | PAYOUT             | create, update, delete |                            |
| products         | PRODUCT_OR_SERVICE | create, update, delete |                            |

Notes:

- Webhook payloads are minimal (`topic`, `accountId`, `itemId`, `occurredAt`);
  full records are hydrated via `GetRecordsByIds`, which issues the singular
  GraphQL getters (`client(id:)`, `job(id:)`, ...).
- Signature verification: `X-Jobber-Hmac-SHA256` header carries
  `Base64(HMAC-SHA256(app client secret, raw body))`.
- Receiving a topic requires the app to hold the matching read OAuth scope.
- Delivery is at least once; endpoints must respond within 1 second or Jobber
  may disable the app's webhooks.
- `WatchFields` is not supported — Jobber webhooks carry no changed-field data.
