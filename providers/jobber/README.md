# Jobber connector

  
## Supported Objects   
Below is an exhaustive list of objects & methods supported on the objects
  
| Object                    | Resource                    | Method       |
| --------------------------| ----------------------------| -------------|
| appAlerts                 | appAlerts                   | read         |
| apps                      | apps                        | read         |
| capitalLoans              | capitalLoans                | read         |
| clientEmails              | clientEmails                | read         |
| clientPhones              | clientPhones                | read         |
| clients                   | clients                     | read         |
| expenses                  | expenses                    | read         |
| invoices                  | invoices                    | read         |
| jobs                      | jobs                        | read         |
| paymentsRecords           | paymentsRecords             | read         |
| payoutRecords             | payoutRecords               | read         |
| products                  | products                    | read         |
| properties                | properties                  | read         |
| quotes                    | quotes                      | read         |
| requestSettingsCollection | requestSettingsCollection   | read         |
| requests                  | requests                    | read         |
| scheduledItems            | scheduledItems              | read         |
| similarClients            | similarClients              | read         |
| tasks                     | tasks                       | read         |
| taxRates                  | taxRates                    | read         |
| timeSheetEntries          | timeSheetEntries            | read         |
| users                     | users                       | read         |
| vehicles                  | vehicles                    | read         |
| vists                     | vists                       | read         |
| clients                   | clientCreate, clientEdit    | write        |
| events                    | eventCreate                 | write        |
|                           | expenseCreate               |              |
| expenses                  | expenseEdit                 | write        |
|                           | expenseDelete               |              |
| jobs                      | jobCreate, jobEdit          | write        |
| productsAndServices       | productsAndServicesCreate   | write        |
|                           | productAndServicesEdit      | write        |
| quotes                    | quoteCreate, quoteEdit      | write        |
| requests                  | requestCreate, requestEdit  | write        |
| taxes                     | taxCreate                   | write        |
| taxGroups                 | taxGroupCreate              | write        |
| vehicles                  | vehicleCreate, vehicleDelete| write        |
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
