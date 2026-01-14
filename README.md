<br/>
<div align="center">
    <a href="https://www.withampersand.com/?utm_source=github&utm_medium=readme&utm_campaign=connectors-repo&utm_content=logo">
    <img src="https://res.cloudinary.com/dycvts6vp/image/upload/v1723671980/ampersand-logo-black.svg" height="30" align="center" alt="Ampersand logo" >
    </a>
<br/>
<br/>

<div align="center">

[![Star us on GitHub](https://img.shields.io/github/stars/amp-labs/connectors?color=FFD700&label=Stars&logo=Github)](https://github.com/amp-labs/connectors) [![Discord](https://img.shields.io/badge/Join%20The%20Community-black?logo=discord)](https://discord.gg/BWP4BpKHvf) [![Documentation](https://img.shields.io/badge/Read%20our%20Documentation-black?logo=book)](https://docs.withampersand.com) ![PRs welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg) <img src="https://img.shields.io/static/v1?label=license&message=MIT&color=white" alt="License">
</div>

</div>

# Overview

[Ampersand](https://withampersand.com?trk=readme-github) is a declarative platform for SaaS builders who are creating product integrations. It allows you to:

* Read data from your customer’s SaaS
* Write data to your customer’s SaaS
* Subscribe to events (creates, deletes, and field changes) in your customer’s SaaS

# Ampersand Connectors

This is a Go library that makes it easier to make API calls to SaaS products such as Salesforce and Hubspot. It handles constructing the correct API requests given desired objects and fields.

It can be either be used as a standalone library, or as a part of the [Ampersand platform](https://docs.withampersand.com/), which offers additional benefits such as:
- Handling auth flows
- Orchestration of scheduled reads, real-time writes, or bulk writes
- Handling API quotas from SaaS APIs
- A dashboard for observability and troubleshooting

The key components of the Ampersand platform include:

- Manifest file (`amp.yaml`): Define all your integrations, the APIs to connect to, the objects and fields for reading or writing, and the configuration options you want to expose to your customers.

- Ampersand server: a managed service that keeps track of each of your customer’s configurations, and makes the appropriate API calls to your customer’s SaaS, while optimizing for cost, handling retries and error message parsing.

- Embeddable UI components: open-source React components that you can embed to allow your end users to customize and manage their integrations. See [the repo](https://github.com/amp-labs/react) for more info.

- Dashboard: Provides deep observability into customer integrations, allowing you to monitor & troubleshoot with detailed logs.

Add enterprise-grade integrations to your SaaS this week. **[Get started for free](https://dashboard.withampersand.com/sign-up?trk=readme-github)**.

<div align="center">
    <img src="https://res.cloudinary.com/dycvts6vp/image/upload/v1724756323/media/hqukkkmpk96zavslpmw5.png" alt="Ampersand Overview" width="80%">
</div>

# Using connectors

## Supported connectors

Browse [the providers directory](https://github.com/amp-labs/connectors/tree/main/providers) to see a list of all the connectors that Ampersand supports, and which features are supported for each connector.

## Examples

Visit the [Ampersand docs](https://docs.withampersand.com?trk=readme-github) to learn about how to use connectors as a part of the Ampersand platform. 

See the [examples directory](https://github.com/amp-labs/connectors/tree/main/examples) for examples of how to use connectors as a standalone library.

| Provider      | Auth Connector  | Deep Connector | Authorization Method |
|---------------|-----------------|----------------|----------------------|
| **Salesforce** | [example](https://github.com/amp-labs/connectors/tree/main/examples/auth_connectors/salesforce) | [example](https://github.com/amp-labs/connectors/tree/main/examples/deep_connectors/salesforce) | OAuth2, Authorization Code    |
| **Adobe** | [example](https://github.com/amp-labs/connectors/tree/main/examples/auth_connectors/adobe)      | | OAuth2, Client Credentials | 
| **Anthropic** | [example](https://github.com/amp-labs/connectors/tree/main/examples/auth_connectors/antrhopic)  | | API Key               |
| **Blueshift** | [example](https://github.com/amp-labs/connectors/tree/main/examples/auth_connectors/blueshift) | | Basic Auth            |

# Concurrency Safety

This codebase uses the `future` and `simultaneously` packages to provide safe concurrency primitives. **Do NOT use the bare `go` keyword** - always use these primitives instead.

## Using the `future` package

For launching async operations that return a result:

```go
// Instead of: go func() { ... }()
// Use future.Go for simple async operations:
result := future.Go(func() (User, error) {
    return fetchUser(id)
})
user, err := result.Await()

// With context support:
result := future.GoContext(ctx, func(ctx context.Context) (User, error) {
    return fetchUserWithContext(ctx, id)
})
user, err := result.AwaitContext(ctx)
```

## Using the `simultaneously` package

For running multiple operations in parallel with controlled concurrency:

```go
// Instead of launching multiple goroutines with: go func() { ... }()
// Use simultaneously.Do to run functions in parallel:
err := simultaneously.Do(maxConcurrent,
    func(ctx context.Context) error { return processItem1(ctx) },
    func(ctx context.Context) error { return processItem2(ctx) },
    func(ctx context.Context) error { return processItem3(ctx) },
)

// With context:
err := simultaneously.DoCtx(ctx, maxConcurrent, callbacks...)
```

**Why?** These primitives automatically handle panic recovery and prevent unbounded goroutine spawning, protecting against production outages.

# Linter

## One-time Setup

Build the custom linters:
```bash
make custom-gcl
```

Rebuild the linters from scratch. This is useful when the linter has been expanded with new plugins:
```bash
make linter-rebuild
```

## Day-to-Day Usage

Run all linters:
```bash
make lint
```

Automatically apply formatting fixes:
```bash
make fix
```

# Tests

Run the full test suite:
```bash
make test
```

Run tests with prettier, more readable output:
```bash
make test-pretty
```

Run tests in parallel to verify test isolation and correctness:
```bash
make test-parallel
```
Notes on parallelized tests:
  * `-parallel=N`: Runs up to `N` (ex:8) test functions concurrently. Useful for speeding up large test suites and for catching concurrency-related bugs.
  * `-count=M`: Runs the test `M` (ex:3) times. This helps catch flakiness or non-deterministic behavior in tests.

# Contributors

Thankful to the OSS community for making Ampersand better every day.

<a href="https://github.com/amp-labs/connectors/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=amp-labs/connectors" />
</a>
