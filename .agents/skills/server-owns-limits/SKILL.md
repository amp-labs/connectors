---
name: server-owns-limits
description: "What belongs to Ampersand's server rather than a connector: rate limits, pacing, retries and backoff, and what a connector does instead. Load when a provider documents rate limits, returns 429s, or a read makes extra requests per record, and when reviewing any of that."
paths:
  - "providers/**"
metadata:
  tags: [connectors, build, review]
---

# The server owns limits and retries

## Why

A connector is a library the server calls once for each request the server decides to make. The
server sees every installation's traffic to a provider, so only it can pace that traffic against
the provider's limit, retry at the right moment and back off across all of it. A connector that
paces, sleeps or retries works against the server: it holds a worker while it waits, retries what
the server will retry again, and applies one plan's limit to every consumer.

## What you can use

- The error types in `common`: a 429 interpreted by the default handler keeps its status and
  headers, `Retry-After` included, and classes as rate limited (`common/errorclass.go`);
  `common.ErrLimitExceeded` is what a limit reported some other way maps to.
- The server's limiter defaults (`shared/limiter/defaults.go` in the server repo), which is where
  a provider's published limits are recorded, with the docs link.
- The provider's docs on its limits and plans.

## Working it out

The connector's part is to tell the server the truth: return the provider's error as a limit when it
is one, and record the published limits where the server reads them. When a read makes extra
requests per record, a page's cost against the lowest plan's documented limit is a fact the team
needs, so the PR says it; whether that object belongs in scope is the team's call. A pacer, sleep or
retry loop in a connector is a defect to remove, whatever the provider's limits are.
