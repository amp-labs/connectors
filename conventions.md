# Ampersand Connector Conventions

> This document is the primary onboarding reference for building a new Ampersand connector. It synthesizes all connector development guides, auth patterns, testing conventions, and operational knowledge into a single source of truth. Treat this as your starting point before writing a single line of code.

---

## Table of Contents

1. [What is Ampersand?](#what-is-ampersand)
2. [What is a Connector?](#what-is-a-connector)
3. [Repository Structure](#repository-structure)
4. [Connector Anatomy](#connector-anatomy)
5. [Authentication Schemes](#authentication-schemes)
6. [Actions: Read, Write, Subscribe, Proxy, Search](#actions)
7. [Pagination Patterns](#pagination-patterns)
8. [Field Normalization](#field-normalization)
9. [Error Handling](#error-handling)
10. [Rate Limiting](#rate-limiting)
11. [Testing Your Connector](#testing-your-connector)
12. [Subscribe / Webhook Setup](#subscribe--webhook-setup)
13. [Common Gotchas](#common-gotchas)
14. [New Connector Checklist](#new-connector-checklist)

---

## What is Ampersand?

Ampersand is a developer-first integration infrastructure platform. It enables SaaS companies to embed native, customer-facing integrations into their products without building and maintaining the underlying plumbing.

Ampersand provides:
- **Pre-built connectors** to 100+ SaaS providers (Salesforce, HubSpot, Gong, Slack, Marketo, etc.)
- **A manifest-based API** (`amp.yaml`) for defining integration actions: reads, writes, event subscriptions, proxy, and search
- **White-labeled UI components** that end users configure to connect their own accounts
- **Managed sync infrastructure** — pagination, rate limits, retries, webhook ingestion, and field normalization are handled by Ampersand, not the application developer

**Who uses it:** SaaS developers building customer-facing integrations. Builders write `amp.yaml` manifests that declare what data to sync; Ampersand handles all the provider-specific mechanics.

---

## What is a Connector?

A **connector** is the bridge between Ampersand's unified platform and a specific SaaS provider (e.g., HubSpot, Salesforce, Gong). Each connector implements:

- **Authentication** — how to obtain and refresh tokens for a given provider
- **Read** — how to fetch records from the provider's API (with pagination, field selection, incremental sync)
- **Write** — how to create/update/delete records in the provider's API
- **Subscribe** — how to receive real-time events from the provider (webhooks, CDC, EventBridge, etc.)
- **Proxy** — routing authenticated API calls through Ampersand without Ampersand modifying the payload
- **Search** — looking up records by something other than their ID

Connectors live in the `github.com/amp-labs/connectors` repository (open source, separate from the server).

---

## Repository Structure

```
connectors/
├── providers/          # One subdirectory per provider
│   ├── hubspot/
│   │   ├── connector.go     # Main connector struct and NewConnector()
│   │   ├── params.go        # Option types (WithWorkspace, WithMetadata, etc.)
│   │   ├── read.go          # Read implementation
│   │   ├── write.go         # Write implementation
│   │   ├── metadata.go      # Field/object metadata
│   │   ├── read_test.go     # Read tests
│   │   ├── write_test.go    # Write tests
│   │   └── metadata_test.go # Metadata tests
│   ├── salesforce/
│   ├── gong/
│   └── ...
├── providers/
│   ├── hubspot.go      # ProviderInfo declaration (BaseURL, auth scheme, OAuth config, etc.)
│   ├── salesforce.go
│   └── ...
├── internal/
│   └── generated/
│       └── catalog.json    # Auto-generated from all provider declarations
├── test/               # Shared test utilities
│   └── utils.go
└── common/             # Shared utilities used across connectors
    ├── urlbuilder/
    ├── catalogreplacer/
    └── ...
```

**Key rule:** The `providers/*.go` file (e.g., `providers/gong.go`) is the *declaration* — metadata, URLs, auth config. The `providers/gong/` directory is the *implementation* — actual API calls, pagination logic, field normalization.

---

## Connector Anatomy

### The `ProviderInfo` Declaration (`providers/gong.go`)

This is the first file you create. It tells Ampersand how to talk to the provider at a high level.

```go
package providers

const Gong = "gong"

func init() {
    SetInfo(Gong, ProviderInfo{
        DisplayName: "Gong",
        AuthType:    Oauth2,
        BaseURL:     "{{.api_base_url_for_customer}}",  // templated for multi-region
        Oauth2Opts: &OauthOpts{
            GrantType:                 AuthorizationCode,
            AuthURL:                   "https://app.gong.io/oauth2/authorize",
            TokenURL:                  "https://app.gong.io/oauth2/generate-customer-token",
            ExplicitScopesRequired:    true,
            ExplicitWorkspaceRequired: false,
            TokenMetadataFields: TokenMetadataFields{
                ScopesField:      "scope",
                WorkspaceRefField: "api_base_url_for_customer",  // captured from token response
            },
        },
        Support: Support{
            BulkWrite: SupportLevel{Objects: SupportLevelNone},
            Proxy:     &ProxySupport{Enabled: true},
            Read:      &ReadSupport{Enabled: true},
            Subscribe: &SubscribeSupport{Enabled: false},
            Write:     &WriteSupport{Enabled: true},
        },
        Media: &Media{
            DarkMode: &MediaTypeConfig{IconURL: "https://..."},
            Regular:  &MediaTypeConfig{IconURL: "https://..."},
        },
    })
}
```

**Critical fields:**
- `BaseURL` — the base URL for all API calls. Use `{{.workspace}}` or `{{.api_base_url_for_customer}}` for multi-tenant/multi-region providers. The templating is resolved at runtime using the connection's metadata.
- `AuthType` — `Oauth2`, `ApiKey`, `Basic`, `NoAuth`
- `WorkspaceRefField` — which field in the OAuth token response contains the tenant-specific URL/identifier. This gets stored in `providerMetadata` on the connection.
- `Support` — declare which actions this connector implements. Be honest — if you haven't implemented write, set it to `SupportLevelNone`.

### The Connector Struct (`providers/gong/connector.go`)

```go
type Connector struct {
    BaseURL string
    Client  *common.JSONHTTPClient
    // ... other fields
}

func NewConnector(opts ...Option) (*Connector, error) {
    params, err := paramsFromOptions(opts)
    if err != nil {
        return nil, err
    }

    // Resolve the base URL using the workspace/metadata variables
    conn := &Connector{}
    
    providerInfo, err := providers.ReadInfo(providers.Gong, &catalogreplacer.CatalogVariables{
        Workspace: params.APIBaseURL,
    })
    if err != nil {
        return nil, err
    }
    
    conn.setBaseURL(providerInfo.BaseURL)
    conn.Client = common.NewJSONHTTPClient(params.Client, conn.BaseURL)
    return conn, nil
}
```

### The `params.go` File

Defines option types for `NewConnector`:

```go
type Option func(*parameters)

type parameters struct {
    Client     common.AuthenticatedHTTPClient
    APIBaseURL string  // for multi-region providers
}

func WithWorkspace(apiBaseURL string) Option {
    return func(p *parameters) {
        p.APIBaseURL = apiBaseURL
    }
}
```

---

## Authentication Schemes

### OAuth 2.0 (Most Common)

The `GrantType` is almost always `AuthorizationCode`. Ampersand handles:
- Generating the OAuth authorization URL
- Exchanging the auth code for tokens
- Storing and refreshing tokens via token-manager
- Injecting the Bearer token into every request

**You do not implement token refresh** — that's Ampersand's job. Your connector receives an already-authenticated `http.Client`.

Key fields to configure:
- `AuthURL` — the provider's OAuth authorization endpoint
- `TokenURL` — the provider's token exchange endpoint
- `ExplicitScopesRequired` — if `true`, the builder must specify scopes in their `amp.yaml`
- `Scopes` — default scopes if `ExplicitScopesRequired` is false
- `TokenMetadataFields.WorkspaceRefField` — which field in the token response contains the tenant-specific URL or identifier

**Multi-region providers:** Some providers (Gong, Zoho, QuickBooks) return a tenant-specific base URL in the token response. You must:
1. Set `WorkspaceRefField` to the correct JSON field name (read the provider's OAuth docs — do not guess)
2. Template your `BaseURL` to use that field: `"{{.api_base_url_for_customer}}"`
3. In `NewConnector`, pass the workspace value to `providers.ReadInfo` via `catalogreplacer.CatalogVariables`

### API Key

```go
AuthType: ApiKey,
ApiKeyOpts: &ApiKeyOpts{
    AttachmentType: Header,
    Header:         "Authorization",
    ValuePrefix:    "Bearer ",
},
```

### Basic Auth

```go
AuthType: Basic,
```

### No Auth

```go
AuthType: NoAuth,
```

---

## Actions

### Read

Read fetches records from a provider object (e.g., HubSpot Contacts) and delivers them to the configured destination.

**Three modes:**
1. **Backfill** — reads all historical records up to now
2. **Incremental** — scheduled, reads records updated since last sync (uses a `lastModifiedField`)
3. **Triggered read** — same as incremental but manually invoked via API; tracks state independently

**Implementation:**

```go
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
    // 1. Build URL with field selection and pagination cursor
    url, err := c.buildReadURL(config)
    if err != nil {
        return nil, err
    }

    // 2. Make the API call
    resp, err := c.Client.Get(ctx, url, nil)
    if err != nil {
        return nil, err
    }

    // 3. Parse response into common.ReadResult
    return &common.ReadResult{
        Rows:     parseRows(resp),
        NextPage: extractNextPageToken(resp),
        Done:     isLastPage(resp),
    }, nil
}
```

**Field selection:** Builders configure which fields to sync via `amp.yaml`. Your read implementation receives `config.Fields []string` — you must construct the provider API request to return only those fields (e.g., HubSpot uses `?properties=field1,field2`).

**The 414 trap:** Sending many field names in a URL query parameter can exceed provider URL length limits (Cloudflare rejects at ~8KB). If a provider uses query params for field selection, be aware of this limit. Prefer POST-based field selection if the provider supports it.

**`lastModifiedField`:** For incremental sync, the provider must have a field that indicates when a record was last updated. Declare it in your object metadata. If the provider doesn't have one, incremental sync is not possible for that object.

### Write

Write creates, updates, or deletes records in the provider. The write API endpoint is `write.withampersand.com`.

**Two modes:**
- **Synchronous** — Ampersand proxies the call, waits for the response, returns it to the caller
- **Asynchronous** — Ampersand queues the write and manages retries/backoff

**Deletes live in the write service.** "Write" is the mutation service — creates, updates, AND deletes.

```go
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
    // config.ObjectName is the object type (e.g., "contact")
    // config.RecordId is the ID for updates/deletes (empty for creates)
    // config.RecordData is the fields to write
    
    if config.RecordId == "" {
        return c.createRecord(ctx, config)
    }
    return c.updateRecord(ctx, config)
}
```

### Proxy

Proxy is the simplest action — Ampersand authenticates the request and passes it through unchanged. It is supported by 100% of providers and requires no implementation work beyond declaring `Proxy: &ProxySupport{Enabled: true}` in the `ProviderInfo`.

The proxy endpoint is `proxy.withampersand.com`. Builders use it to access provider API functionality that Ampersand's read/write/subscribe doesn't expose. It does NOT retry, does NOT rate-limit.

### Subscribe

Subscribe receives real-time events from providers. It is the most complex action and is usually implemented last.

Provider event delivery varies dramatically:
- **Simple webhook registration** — call a provider API to register a callback URL
- **Salesforce CDC** — elaborate setup sequence: Platform Events → CometD → EventBridge → SNS → Ampersand cloud function. Temporal handles setup/teardown.
- **HubSpot** — webhook subscriptions via API
- **Gong Flows** — requires private-preview tenant access to be enabled by Gong

Inbound events arrive at an Ampersand cloud function, get published to Pub/Sub, then Messenger picks them up. Often the event only tells you *what changed* (not the full record), so Messenger calls back to the provider to fetch full record data — this creates a race condition where the record may have changed again by the time you fetch it.

**Subscribe is optional and expensive to implement.** Unless the provider supports it cleanly, it's typically the last capability built.

### Search

Search allows builders to look up records by something other than a record ID.

```go
func (c *Connector) Search(ctx context.Context, config common.SearchParams) (*common.SearchResult, error) {
    // config.ObjectName — the object to search
    // config.SearchQuery — the search terms
    // config.Fields — which fields to return
}
```

Search is the newest capability and has limited provider support. If the provider's search API is poorly documented or has severe limitations, flag it rather than building a broken implementation.

---

## Pagination Patterns

Almost all provider APIs paginate. Ampersand uses a cursor-based abstraction.

### Token-based pagination (most common)

The API returns a `next_page_token` or `cursor` in the response. Pass it back on the next request.

```go
func (c *Connector) buildReadURL(config common.ReadParams) (string, error) {
    url := urlbuilder.New(c.BaseURL, "/v2/objects/contacts")
    url.AddQueryParam("limit", "100")
    
    if config.NextPage.String() != "" {
        url.AddQueryParam("after", config.NextPage.String())
    }
    return url.Build()
}

func extractNextPage(resp *ContactsResponse) common.NextPageToken {
    if resp.Paging.Next.After == "" {
        return ""
    }
    return common.NextPageToken(resp.Paging.Next.After)
}
```

### Offset-based pagination

Use the page number or offset as the cursor.

```go
// Encode offset as the next page token
nextPage := common.NextPageToken(strconv.Itoa(currentOffset + pageSize))
```

### Link-header pagination

Some providers return a `Link: <url>; rel="next"` header. Extract the URL and use it directly.

### No more pages signal

Set `Done: true` in `ReadResult` when there are no more pages. This tells Ampersand the backfill is complete and incremental sync can begin.

```go
return &common.ReadResult{
    Rows:     rows,
    NextPage: nextPage,
    Done:     nextPage == "",
}, nil
```

---

## Field Normalization

Ampersand normalizes field names so builders get consistent keys regardless of provider-specific naming quirks.

**Declare object metadata:**

```go
func (c *Connector) ListObjectMetadata(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error) {
    // Return metadata for each requested object
    return &common.ListObjectMetadataResult{
        Result: map[string]common.ObjectMetadata{
            "contact": {
                DisplayName: "Contact",
                FieldsMap: map[string]common.Field{
                    "id":         {DisplayName: "ID", ValueType: "string"},
                    "email":      {DisplayName: "Email", ValueType: "string"},
                    "first_name": {DisplayName: "First Name", ValueType: "string"},
                    // ...
                },
            },
        },
    }, nil
}
```

**Standard field conventions:**
- Use `snake_case` for field names
- Always include `id` as a field (it's required for updates/deletes)
- Declare `lastModifiedField` for objects that support incremental sync
- Map provider-specific names to human-readable display names

---

## Error Handling

### HTTP error mapping

Use the common error types:

```go
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
    resp, err := c.Client.Get(ctx, url, nil)
    if err != nil {
        // common.HTTPStatusError wraps provider errors with status codes
        var httpErr *common.HTTPStatusError
        if errors.As(err, &httpErr) {
            switch httpErr.Status {
            case http.StatusUnauthorized:
                return nil, common.ErrAccessToken
            case http.StatusForbidden:
                return nil, common.ErrForbidden
            case http.StatusNotFound:
                return nil, common.ErrObjectNotFound
            case http.StatusTooManyRequests:
                return nil, common.ErrRateLimited
            }
        }
        return nil, err
    }
    // ...
}
```

**Critical error types:**
- `common.ErrAccessToken` — signals token-manager to refresh the token
- `common.ErrRateLimited` — triggers Ampersand's backoff/retry logic
- `common.ErrObjectNotFound` — object type not supported by this connector
- `common.ErrForbidden` — permission issue (different from auth failure)

### Never swallow errors silently

If an API call fails, return the error. Do not return empty results on failure — Ampersand's retry logic depends on seeing the error.

---

## Rate Limiting

Ampersand has its own rate limiting layer (Gubernator, per-provider, per-customer). As a connector author, your job is:

1. **Return `common.ErrRateLimited`** when the provider returns a 429. Ampersand handles the backoff.
2. **Read `Retry-After` headers** if the provider includes them and propagate that information if the error type supports it.
3. **Do not implement your own retry loop** inside a connector — that's Ampersand's job at the Temporal layer.

---

## Testing Your Connector

### Test structure

Each connector has a `constructTestConnector` helper that creates a connector pointed at a mock HTTP server:

```go
func constructTestConnector(server *httptest.Server) (*Connector, error) {
    return NewConnector(
        WithWorkspace("https://api.gong.io"),  // use the full URL, not just the domain
        WithClient(server.Client()),
    )
}
```

**Critical:** If your connector uses a templated `BaseURL` (e.g., `{{.api_base_url_for_customer}}`), you MUST pass the full URL (including scheme: `https://`) to `WithWorkspace`. Passing just a domain without the scheme causes URL parsing failures.

### Test patterns

```go
func TestRead(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Serve mock responses based on the request URL/path
        switch r.URL.Path {
        case "/v2/objects/contacts":
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(mockContactsResponse)
        }
    }))
    defer server.Close()

    connector, err := constructTestConnector(server)
    require.NoError(t, err)

    result, err := connector.Read(context.Background(), common.ReadParams{
        ObjectName: "contact",
        Fields:     []string{"id", "email", "first_name"},
    })
    require.NoError(t, err)
    assert.Equal(t, 2, result.Rows)
    assert.False(t, result.Done)  // more pages available
}
```

### Mock response fixtures

Store mock API responses as JSON files in a `testdata/` directory within the connector package. This makes tests easier to read and update when the API changes.

### Linter

Run the linter before committing:

```bash
cd /path/to/connectors
golangci-lint run ./providers/gong/...
```

The linter enforces `wsl_v5` which requires blank lines before certain `if` blocks. Watch for this — it causes CI failures.

### Running tests

```bash
go test ./providers/gong/...
```

All tests must pass before submitting a PR.

---

## Subscribe / Webhook Setup

Subscribe is provider-specific and complex. Here is the general pattern:

### Simple webhook registration

1. Provider exposes an API to register a webhook URL
2. Temporal calls the registration API during installation setup (`SubscribeCreateInstallationWorkflow`)
3. The webhook URL points to an Ampersand cloud function
4. Cloud function receives events → publishes to Pub/Sub → Messenger delivers to builder

**Idempotency is critical:** If the setup workflow is interrupted and retried, it must handle "resource already exists" gracefully. `ResourceAlreadyExistsException` should be treated as success, not an error. This is a recurring source of bugs.

### Salesforce (complex case)

Salesforce uses Change Data Capture (CDC) via EventBridge:
1. Enable CDC for the desired objects in Salesforce
2. Salesforce publishes changes to AWS EventBridge (partner event bus)
3. An EventBridge rule forwards to SNS
4. SNS triggers an Ampersand cloud function
5. Cloud function publishes to Pub/Sub
6. Messenger picks up and delivers

This requires Temporal to orchestrate a multi-step setup/teardown sequence. If setup fails midway, teardown must clean up partial state.

### Event completeness

Subscribe events often only contain the changed fields, not the full record. Messenger calls back to the provider to fetch missing fields. This creates a race condition — the record may have changed again by the time you fetch it. This is a known limitation and is communicated to builders as expected behavior.

---

## Common Gotchas

### 1. Multi-region base URLs

If the provider returns a tenant-specific API URL in the OAuth token response, you MUST capture it. Do not hardcode `https://api.provider.com` if non-US customers will get a different URL.

**How to check:** Read the provider's OAuth token exchange documentation. Look for any field in the token response that looks like a URL or domain. Do not assume it's `api_base_url` — check the actual field name.

**Reference:** Look at `providers/gong.go` (`api_base_url_for_customer`) and `providers/zoho.go` (`api_domain`) for working examples.

### 2. `optionalFieldsAuto: all` and the 414 trap

If a connector's `amp.yaml` uses `optionalFieldsAuto: all`, Ampersand will attempt to sync ALL available fields for that object. For providers with many fields (HubSpot contacts can have hundreds), this creates URL query strings that exceed provider limits (typically Cloudflare rejects URLs over ~8KB with HTTP 414). The sync pauses and never recovers until the configuration is updated. Warn builders about this limitation.

### 3. Credential revocation

The single most common support ticket. When a user who authorized the connection later:
- Revokes API access in the provider's UI
- Loses the admin role that granted the permission
- Changes their password (some providers revoke all tokens on password change)

...the connection is marked `bad_credentials` and all syncs fail. **Ampersand cannot fix this automatically.** The fix is always re-authorization by the end user. Build detection into your connector (`ErrAccessToken`) so Ampersand can surface this clearly.

### 4. Private-preview API access

Some providers (e.g., Gong Flows, certain Salesforce APIs) require tenant-level feature flags to be enabled by the provider's account team. This is separate from OAuth scopes. The connector code can be correct and the credentials valid, but the API still returns 404 or 401. If you're building a connector for such an API, document this requirement clearly.

### 5. Soft-delete + unique constraint

Ampersand's database uses soft deletes (a `delete_time` column). If a connection is soft-deleted and the user re-authorizes, the INSERT can collide with the soft-deleted row's unique index. This is a platform bug pattern to be aware of — the fix is typically an UPSERT or a partial unique index (`WHERE delete_time IS NULL`).

### 6. The catalog is live-loaded

The `catalog.json` file (compiled from all `providers/*.go` declarations) is fetched from GitHub at runtime by the server. This means:
- Changes to `providers/gong.go` are not live until the connectors PR merges and the catalog refreshes
- In PR preview environments, you must set `FETCH_LATEST_PROVIDER_CATALOG=false` on ALL services that import the catalog package (api, token-manager, messenger, temporal, builder-mcp) to force use of the pinned build

### 7. Direct push to main is not allowed

Always use a feature branch and open a PR. Branch naming convention:
- `feat/` — new feature or connector
- `fix/` — bug fix
- `chore/` — maintenance, refactoring

PR title should be prefixed with `[CON-123]` if there is a Linear ticket. Use the appropriate PR template if one matches; free-hand it if none do.

---

## New Connector Checklist

Use this before submitting a PR for a new connector.

### Research phase
- [ ] Read the provider's API documentation thoroughly
- [ ] Read the provider's OAuth documentation — note the exact field names in the token response
- [ ] Check if the provider has multiple regions and returns a tenant-specific base URL
- [ ] Identify which objects are commonly needed (contacts, companies, deals, etc.)
- [ ] Identify the `lastModifiedField` for each object (needed for incremental sync)
- [ ] Check for private-preview API requirements and document them
- [ ] Check Slab for any existing research notes on this provider

### Implementation phase
- [ ] Create `providers/{provider_name}.go` with correct `ProviderInfo`
  - [ ] `BaseURL` is templated if provider is multi-region
  - [ ] `WorkspaceRefField` is set to the EXACT field name from the token response (not guessed)
  - [ ] `Support` accurately reflects what is actually implemented
- [ ] Create `providers/{provider_name}/` directory
- [ ] Implement `NewConnector` with `WithWorkspace` option if multi-region
- [ ] Implement `Read` with pagination
- [ ] Implement `ListObjectMetadata` with accurate field declarations
- [ ] Declare `lastModifiedField` for each object that supports incremental sync
- [ ] Implement `Write` if the provider supports it
- [ ] Map HTTP error codes to `common.Err*` types
- [ ] Return `common.ErrRateLimited` on 429 responses

### Testing phase
- [ ] `constructTestConnector` passes the FULL URL (with scheme) to `WithWorkspace`
- [ ] Mock server responds correctly for all tested endpoints
- [ ] Tests cover: basic read, pagination, empty results, error cases
- [ ] `go test ./providers/{provider_name}/...` passes locally
- [ ] Linter passes: `golangci-lint run ./providers/{provider_name}/...`

### PR phase
- [ ] Branch name follows convention (`feat/`, `fix/`, `chore/`)
- [ ] PR title includes `[CON-123]` if Linear ticket exists
- [ ] PR description: what changed, why, test results
- [ ] CI passes (build + lint + semgrep)
- [ ] Human review and approval
- [ ] After merge: server dependency is auto-bumped; monitor the promotion pipeline

---

## Reference: Key Files and Packages

| File/Package | Purpose |
|---|---|
| `providers/hubspot.go` | Example of a mature OAuth2 provider declaration |
| `providers/zoho.go` | Example of multi-region provider with `api_domain` workspace field |
| `providers/gong.go` | Example of `api_base_url_for_customer` workspace field pattern |
| `providers/salesforce.go` | Example of most complex subscribe setup |
| `common/urlbuilder/` | URL construction with query params |
| `common/catalogreplacer/` | Template variable substitution for BaseURL |
| `internal/generated/catalog.json` | Auto-generated catalog; do not edit by hand |
| `test/utils.go` | Shared test helpers |

---

*This document was synthesized from the Connectors > Guides Slab topic. If you find a gap or inaccuracy, update both the relevant Slab post and this file.*
