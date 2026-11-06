---
name: time-filters
description: "Reading only what changed, correctly: Since and Until, which order a time filter in code may declare, why positional paging keeps Until out of the request, and why a list endpoint's default filters must be overridden. Load when implementing or reviewing a read's Since/Until, a local time filter, or a list endpoint with documented defaults."
paths:
  - "providers/**"
metadata:
  tags: [connectors, build, review]
---

# Time filters and default filters

## Why

Consumers sync incrementally: each read asks for what changed between Since and Until, and the next
read starts where this one ended. A read that returns too little looks healthy: records arrive,
nothing fails, and the consumer never gets the rest. Every behaviour here is a way that happens.

- **Since and Until are inclusive** and filter on the record's update time, so an edited record is
  read again and a record stamped exactly at the boundary isn't lost between two reads. Sending
  Since to the provider as a strict "greater than" breaks that, unless Since is formatted coarser
  than the provider's timestamps.
- **Positional paging and Until.** With offset, page-number or position-like cursors, a window that
  narrows mid-read shifts the positions between pages and skips records for good. Under positional
  paging only Since goes to the provider and Until is trimmed in code; a keyset cursor may carry
  Until.
- **A time filter in code declares an order**, and the order decides when paging stops.
  `readhelper.Unordered` reads every page; `ChronologicalOrder` or `ReverseOrder` stops at the first
  page past the window. A declared order that isn't the provider's real one, or that sorts on a
  different field than the filter reads (creation time against update time), stops early and skips
  every later record with no error. An order is a contract only when the docs state the sort or the
  connector sends a sort parameter; records that looked sorted in a sample are not one.
- **A local filter trusts the update time.** It misses changes that don't bump the record's update
  time (child data edited, for instance) and never sees deletes.
- **List endpoints filter by default**: only open items, only unarchived records, only the caller's
  own, only some states. A read that doesn't override a narrowing default misses those records for
  every consumer. A status filter that takes one value at a time means reading each value in turn,
  with which one is current kept in the cursor.

## What you can use

- The provider's docs: each list endpoint's parameters, defaults, sort and filter operators.
- The repo's helpers (`readhelper.NewTimeBoundary`, `readhelper.MakeTimeFilterFunc` and its orders,
  `ExcludeSince` for the rare provider that needs it) and the connectors that use them, to see how
  the pieces fit.
- The connector's programs under `test/<name>/`, run against a real account, to see what an
  endpoint returns by default and how it sorts there; what one account shows is a sample, not a
  contract.

## Working it out

For each object, work out from the docs which filters the provider offers, how it pages, how it
sorts and what it returns by default, and choose what goes to the provider and what is trimmed in
code so that nothing in the window is skipped. Where the docs leave a behaviour open (an undocumented
default, whether child edits move the update time, how deletes show), say in the PR what the
connector does today and raise the question, rather than assuming the convenient answer. The PR also
says which objects filter in code and why.
