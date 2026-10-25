---
name: server-owns-limits
description: "What belongs to Ampersand's server rather than a connector: rate limits, pacing, retries and backoff. Load when a provider documents rate limits, returns 429s, or a read makes extra requests per record, and when reviewing any of that."
paths:
  - "providers/**"
metadata:
  tags: [connectors, build, review]
---

# The server owns limits and retries

A connector is a library the server calls once per request it decides to make. The server sees
every installation's traffic to a provider, so only it can pace them against the provider's limit,
retry at the right moment, and back off across all of them. A connector that paces or retries
itself works against that: it holds a worker while it sleeps, retries what the server retries
again, and enforces one plan's limit on every consumer. Only one connector paces its reads
(circleback, `ratelimit.go`), and it is the departure, not the pattern.

So a connector:
- returns the provider's error as it came. The default interpreter keeps a 429's status and
  headers, `Retry-After` included, and it classes as rate limited (`common/errorclass.go`); a
  provider that signals a limit another way maps it to `common.ErrLimitExceeded`, as
  `providers/google/internal/core/errors.go` does;
- never sleeps, paces, or retries a read;
- has the provider's published limits recorded where the server reads them, the limiter defaults
  in the server repo (`shared/limiter/defaults.go`), with the docs link.

When a read makes extra requests per record (a transcript per meeting), the cost of a page is a
fact the team should know: say in the PR how many requests a page costs against the lowest plan's
documented limit. Whether that object belongs in scope is the team's call; pacing it inside the
connector isn't the answer.

A review that finds a pacer, a sleep or a retry loop in a read asks for it to be removed, with this
reason, and never asks for one.
