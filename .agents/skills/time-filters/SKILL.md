---
name: time-filters
description: "Reading only what changed, correctly: Since and Until, which order a time filter in code may declare, and why a list endpoint's default filters must be overridden. Load when implementing or reviewing a read's Since/Until, a local time filter, or a list endpoint with documented defaults."
paths:
  - "providers/**"
metadata:
  tags: [connectors, build, review]
---

# Time filters and default filters

An incremental read that returns too little looks healthy: records arrive, nothing fails, and the
consumer never gets the rest. These are the ways it happens.

## Since and Until

- Both are inclusive: a record updated exactly at Since is read (`readhelper.NewTimeBoundary()`,
  22 connectors). An exclusive Since (`ExcludeSince`) is only lightfield's.
- They filter on the record's update time, so an edited record is read again.
- Inclusive holds at the provider too: Since goes as `>=` (`gte`, `from`), not `>` or `gt`, unless
  Since is formatted coarser than the provider's timestamps, which makes `>` inclusive in effect.
- Since goes to the provider's filter when it has one. Under positional paging (offset, page
  number, or a cursor that stands for a position), only Since goes to the provider and Until is
  trimmed in code, because a narrower window shifts the positions between pages. A keyset cursor
  may carry Until.

## A time filter in code declares only a documented order

When the provider has no time filter for an object, the connector reads every page and drops
records outside the window with `readhelper.MakeTimeFilterFunc(order, boundary, …)` (21
connectors). The order it declares decides when paging stops:

- `readhelper.Unordered`: every page is read, and records are filtered one by one.
- `readhelper.ChronologicalOrder` or `ReverseOrder`: paging stops at the first page past the
  window, so the read is shorter.

A declared order that isn't the provider's real one stops early and skips every later record, with
no error to show it. So an order is declared only when the provider's docs state the list's sort
(cite the page in a code comment), or the endpoint takes a sort parameter the connector sends.
Without that, it is `Unordered`, even when the records you saw looked sorted: a sample is not a
contract. The sort must also be on the field the filter reads: a list sorted by creation time and
filtered on update time stops at the first old record and skips older records edited since. Say in the PR which objects filter in code and why.

A local filter also depends on the provider updating the record's update time on every change, and
it never sees deletes; say so in the PR when they apply.

## A list endpoint's defaults hide records

Many list endpoints filter by default: only open or incomplete items, only unarchived records,
only the caller's own, only some invoice states. A read that doesn't override them misses those
records for every consumer, and nothing fails. For each list endpoint, read its parameters in the
docs and override every documented default that narrows the result (send all states, include
archived), or say in the PR why the default is right. When a status filter takes one value at a
time, read each value in turn and keep which one in the cursor.

When the docs don't say what an endpoint returns by default, raise it as a question in the PR, with
what the connector does today.
